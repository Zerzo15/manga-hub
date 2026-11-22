package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"mangahub/internal/auth"
)

type Client struct {
	conn    net.Conn
	writeMu sync.Mutex
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	mu         sync.RWMutex
	clients    map[int64]map[*Client]bool //allow one user to have multiple connection
	jwtService *auth.JWTService
}

type AuthMessage struct {
	Type  string `json:"type"`
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
func NewServer(listenAddr string, jwtService *auth.JWTService) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		clients:    map[int64]map[*Client]bool{},
		jwtService: jwtService,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln
	fmt.Println("TCP server listening on:", s.listenAddr)
	go s.acceptLoop()

	//wait for the quitch channel
	<-s.quitch

	return nil
}

//handling connection
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select { //handle shutdown
			case <-s.quitch:
				return
			default:
				fmt.Println("accept error:", err)
				continue
			}
		}

		fmt.Println("new connection to the server:", conn.RemoteAddr())

		go s.readLoop(conn)
	}
}

func (s *Server) addClient(userID int64, client *Client) {
	s.mu.Lock() //prevent goroutine to modify shared data at the same time
	//useful when multiple user is trying to access
	defer s.mu.Unlock()

	if s.clients[userID] == nil {
		s.clients[userID] = make(map[*Client]bool)
	}
	s.clients[userID][client] = true
}

func (s *Server) removeClient(userID int64, client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conns, ok := s.clients[userID]; ok {
		delete(conns, client)
		if len(conns) == 0 {
			delete(s.clients, userID)
		}
	}
}

func (s *Server) readLoop(conn net.Conn) {
	client := &Client{
		conn: conn,
	}

	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	//give the client 60 second window to enter an authentication
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return
	}

	if !scanner.Scan() {
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(scanner.Bytes(), &auth); err != nil || auth.Type != "auth" {
		conn.Write([]byte(`{"type":"error","message":"required auth"}` + "\n"))
		return
	}

	userID, err := s.validateJWT(auth.Token)
	if err != nil {
		conn.Write([]byte(`{"type":"error","message":"invalid token"}` + "\n"))
		return
	}

	//authentication is sucess
	//remove the previous timer so the connection stay open
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return
	}

	s.addClient(userID, client)
	defer s.removeClient(userID, client)

	if err := client.WriteMessage([]byte(`{"type":"auth_ok"}` + "\n")); err != nil {
		return
	}

	for scanner.Scan() {
		// keep connection open until client disconnects
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scanner error:", err)
	}
}

//guarantee one go routine write to one connection at a time
func (c *Client) WriteMessage(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}

	_, err := c.conn.Write(payload)
	return err
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

	s.mu.RLock()

	userClients := s.clients[userID]
	clients := make([]*Client, 0, len(userClients))

	for client := range userClients {
		clients = append(clients, client)
	}

	s.mu.RUnlock()

	for _, client := range clients {
		if err := client.WriteMessage(payload); err != nil {
			_ = client.conn.Close()
			s.removeClient(userID, client)
		}
	}
}

func (s *Server) validateJWT(tokenString string) (int64, error) {
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
