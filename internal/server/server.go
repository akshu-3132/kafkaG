package server

import (
	"fmt"
	"io"
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
	for {
		frame, err := protocol.ReadFrame(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client Disconnected")
			} else {
				fmt.Println("Error occured while reading frame")
			}
			return
		}
		req, err := protocol.Decode(frame.Payload)
		if err != nil {
			fmt.Println("Error occured while decoding request:", err)
			return
		}
		res, err := s.handleRequest(req)
		if err != nil {
			fmt.Println("Error occured while handling request:", err)
		}
		fmt.Println("Response:", res)
		if res == nil {
			fmt.Println("Response is nil, skipping sending response")
			continue
		}
		responsePayload, err := res.Encode()
		if err != nil {
			fmt.Println("Error occured while encoding response:", err)
			return
		}
		responseFrame := &protocol.Frame{
			Payload: responsePayload,
		}
		if err := responseFrame.WriteTo(conn); err != nil {
			fmt.Println("Error occured while writing response frame:", err)
			return
		}

	}
}
