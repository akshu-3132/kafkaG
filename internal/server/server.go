package server

import (
	"fmt"
	"io"
	"kafkaG/internal/config"
	"kafkaG/internal/protocol"
	"kafkaG/internal/storage"
	"net"
	"time"
)

type Server struct {
	addr           string
	segmentManager *storage.SegmentManager
	config         *config.Config
}

func New(addr string) *Server {
	config, err := config.LoadConfig("/home/akshu/projects/kafkaG/configs/broker.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}
	segmentManager, err := storage.NewSegmentManager(config.StorageConfig.DirPath, &config.StorageConfig)
	if err != nil {
		fmt.Println("Error creating segment manager:", err)
		return nil
	}
	err = segmentManager.RecoverSegments() // Recover segments from disk
	if err != nil {
		fmt.Println("Error recovering segments:", err)
		return nil
	}
	return &Server{
		addr:           addr,
		segmentManager: segmentManager,
		config:         config,
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
			offset, err := s.segmentManager.WriteRecord(record)
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
			record, err := s.segmentManager.ReadRecord(fetchRequest.Offset)
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
