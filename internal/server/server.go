package server

import (
	"fmt"
	"net"

	"kafkaG/internal/protocol"
)

type Server struct {
	addr string
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Broker listening on", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	frame, err := protocol.ReadFrame(conn)
	if err != nil {
		fmt.Println("Failed to read frame:", err)
		return
	}

	fmt.Println("Received:", string(frame.Payload))
}
