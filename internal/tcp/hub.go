package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/handler"
	"net"
	"sync"
)

type Server struct {
	listenAddr string			
	ln         net.Listener
	quitch chan struct{}
	mu sync.RWMutex
	clients map[int64]map[net.Conn]bool
}


type AuthMessage struct {
	Type string `json:"type"`
	Token string `json:"token"`
}

type ProgressEvent struct {
	Type string `json:"type"`
	MangaID int64 `json:"manga_id"`
	Status string `json:"status"`
	CurrentChapter int `json:"current_chap"`
	Version int `json:"version"`
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

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(scanner.Bytes(), &auth); err != nil || auth.Type != "auth" {
		conn.Write([]byte(`{
			"type":"error", 
			"message":"required auth"
		}` + "\n"))
		return
	}

	userID, err := handler.ValidateJWT(auth.Token)
	if err != nil {
		conn.Write([]byte(`{
			"type":"error", 
			"message":"invalid token"
		}` + "\n"))
		return
	}

	s.addClient(userID, conn)

	defer s.removeClient(userId, conn)

	conn.Write([]byte(`{"type":"auth_ok}` + "\n"))
}



func main() {
	server := NewServer(":9090")

	go func()  {
		for msg := range server.msgch {
			//payload is in []byte so we need to transfer it to string.
			fmt.Printf("receive message from %v: %v", msg.from, string(msg.payload))
		}
	}()
	
	log.Fatal(server.Start())
} 