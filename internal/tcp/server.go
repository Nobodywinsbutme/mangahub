package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	listener  net.Listener
	clients   map[net.Conn]bool
	mu        sync.Mutex
	broadcast chan []byte
}

func Start(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	server := &Server{
		listener:  listener,
		clients:   make(map[net.Conn]bool),
		broadcast: make(chan []byte),
	}

	log.Printf("TCP Sync Server listening on port %s", port)

	go broadcastLoop(server)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("TCP Accept error: %v", err)
			continue
		}
		go handleTCPConnection(conn, server)
	}
}

func handleTCPConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	server.mu.Lock()
	server.clients[conn] = true
	server.mu.Unlock()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		message := fmt.Sprintf("Progress: %s", line)
		server.broadcast <- []byte(message)
	}

	server.mu.Lock()
	delete(server.clients, conn)
	server.mu.Unlock()
}

func broadcastLoop(server *Server) {
	for {
		msg := <-server.broadcast
		server.mu.Lock()
		for conn := range server.clients {
			_, err := conn.Write(append(msg, '\n'))
			if err != nil {
				conn.Close()
				delete(server.clients, conn)
			}
		}
		server.mu.Unlock()
	}
}
