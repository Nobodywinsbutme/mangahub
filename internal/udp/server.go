package udp

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	conn    *net.UDPConn
	clients map[string]*net.UDPAddr
	mu      sync.Mutex
}

func Start(port string) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}

	server := &Server{
		conn:    conn,
		clients: make(map[string]*net.UDPAddr),
	}

	log.Printf("UDP Notification Server listening on port %s", port)

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("UDP Read error: %v", err)
			continue
		}
		msg := string(buffer[:n])
		handleUDPMessage(strings.TrimSpace(msg), remoteAddr, server)
	}
}

func handleUDPMessage(msg string, remoteAddr *net.UDPAddr, server *Server) {
	parts := strings.Split(msg, "|")
	if len(parts) < 2 {
		return
	}

	switch parts[0] {
	case "REGISTER":
		username := parts[1]
		server.mu.Lock()
		server.clients[username] = remoteAddr
		server.mu.Unlock()
		log.Printf("Registered UDP client: %s at %s", username, remoteAddr.String())
	case "NOTIFY":
		if len(parts) == 3 {
			mangaTitle := parts[1]
			chapter := parts[2]
			broadcastNotification(server, mangaTitle, chapter)
		}
	}
}

func broadcastNotification(server *Server, mangaTitle, chapter string) {
	notification := fmt.Sprintf("New Chapter: %s - Chapter %s\n", mangaTitle, chapter)

	server.mu.Lock()
	defer server.mu.Unlock()

	for username, addr := range server.clients {
		_, err := server.conn.WriteToUDP([]byte(notification), addr)
		if err != nil {
			log.Printf("Failed to notify %s: %v", username, err)
		}
	}
}
