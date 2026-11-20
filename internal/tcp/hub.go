package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	listenAddr string			
	ln         net.Listener
	quitch chan struct{}
	mu sync.RWMutex
	clients map[int64]map[net.Conn]bool //allow one user to have multiple connection
}

type AuthMessage struct {
	Type string `json:"type"`
	Token string `json:"token"`
}

type ProgressEvent struct {
	Type           string `json:"type"`
	MangaID        int64  `json:"manga_id"`
	Status         string `json:"status,omitempty"`
	CurrentChapter int    `json:"chapter"`
	Version        int    `json:"version"`
}

//take the port number as the parameter (ex ":9090")
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch: make(chan struct{}),
		clients: map[int64]map[net.Conn]bool{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return  err
	}
	defer ln.Close()

	s.ln = ln
	fmt.Println("TCP server listening on:", s.listenAddr)
	go s.acceptLoop()

	//wait for the quitch channel
	<- s.quitch 

	return nil
}

//handling connection 
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		fmt.Println("new connection to the server:", conn.RemoteAddr())

		go s.readLoop(conn)
	}
}

func (s *Server) addClient(userID int64, conn net.Conn) {
	s.mu.Lock() //prevent goroutine to modify shared data at the same time
	//useful when multiple user is trying to access
	defer s.mu.Unlock()

	if s.clients[userID] == nil {
		s.clients[userID] = make(map[net.Conn]bool)
	}
	s.clients[userID][conn] = true
}

func (s * Server) removeClient(userID int64, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conns, ok := s.clients[userID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(s.clients, userID)
		}
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(scanner.Bytes(), &auth); err != nil || auth.Type != "auth" {
		conn.Write([]byte(`{"type":"error","message":"required auth"}` + "\n"))
		return
	}

	userID, err := validateJWT(auth.Token)
	if err != nil {
		conn.Write([]byte(`{"type":"error","message":"invalid token"}` + "\n"))
		return
	}

	s.addClient(userID, conn)
	defer s.removeClient(userID, conn)

	conn.Write([]byte(`{"type":"auth_ok"}` + "\n"))

	for scanner.Scan() {
		// keep connection open until client disconnects
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}

	
}

func (s *Server) BroadcastProgress(userID, mangaID int64, chap, version int) {
	payload, err := json.Marshal(ProgressEvent{
		Type:           "progress_updated",
		MangaID:        mangaID,
		CurrentChapter: chap,
		Version:        version,
	})
	if err != nil {
		return
	}

	payload = append(payload, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.clients[userID] {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err = conn.Write(payload); err != nil {
			_ = conn.Close()
			delete(s.clients[userID], conn)
		}
	}
	if len(s.clients[userID]) == 0 {
		delete(s.clients, userID)
	}
}

func validateJWT(tokenString string) (int64, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("missing user_id")
	}

	return int64(userIDFloat), nil
}