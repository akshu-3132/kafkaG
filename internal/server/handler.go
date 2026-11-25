package server

import (
	"encoding/binary"
	"fmt"
	"kafkaG/internal/protocol"
	"net"
)

const (
	APIKeyProduce uint16 = 1
)

func (s *Server) handleRequest(conn net.Conn) error {
	defer conn.Close()
	var header [4]byte
	var apiKey uint16
	for {
		buffer, err := protocol.Read(conn, header[:], s.BufferPool)
		if err != nil || buffer == nil {
			return fmt.Errorf("error reading")
		}

		apiKey = binary.BigEndian.Uint16(buffer[0:2])

		if apiKey == APIKeyProduce {
			fmt.Println("Received produce request")
		}

		//TODO sending temporary response to make it working
		s.BufferPool.ReturnBuffer(buffer)

	}

}
