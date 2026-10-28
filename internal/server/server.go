package server

import (
	"fmt"
	"io"
	"kafkaG/internal/protocol"
	"kafkaG/internal/storage"
	"net"
	"time"
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

	fl, err := storage.OpenFileLog("./data/test.log")
	if err != nil {
		fmt.Println("Error opening log:", err)
		return
	}
	defer fl.Close()

	for {
		frame, err := protocol.ReadFrame(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected")
			} else {
				fmt.Println("Error reading frame:", err)
			}
			return
		}

		request, err := protocol.Decode(frame.Payload)
		if err != nil {
			fmt.Println("Error decoding request:", err)
			return
		}

		switch request.APIKey {
		case 1: // Produce
			produceRequest, err := protocol.DecodeProduceRequest(request.Body)
			if err != nil {
				fmt.Println("Error decoding produce request:", err)
				return
			}
			record := &storage.Record{
				Timestamp: time.Now().Unix(),
				Message:   produceRequest.Message,
			}
			offset, err := fl.Append(record)
			if err != nil {
				fmt.Println("Error appending record:", err)
				return
			}
			fmt.Println("Appended record at offset:", offset)
			response := &protocol.Response{
				CorrelationID: request.CorrelationID,
				Body:          []byte("Message appended successfully"),
			}
			responseFrame := &protocol.Frame{
				Payload: response.Encode(),
			}
			err = responseFrame.WriteTo(conn)
		case 2: // Fetch
			fetchRequest, err := protocol.DecodeFetchRequest(request.Body)
			if err != nil {
				fmt.Println("Error decoding fetch request:", err)
				return
			}
			record, err := fl.ReadAt(fetchRequest.Offset)
			if err != nil {
				fmt.Println("Error reading record:", err)
				return
			}
			response := &protocol.Response{
				CorrelationID: request.CorrelationID,
				Body:          record.Message,
			}
			responseFrame := &protocol.Frame{
				Payload: response.Encode(),
			}
			err = responseFrame.WriteTo(conn)
			if err != nil {
				fmt.Println("Error writing response:", err)
				return
			}
		}
	}
}
