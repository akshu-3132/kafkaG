package server

import (
	"fmt"
	"kafkaG/internal/config"
	"kafkaG/internal/storage"
	"net"
)

type Server struct {
	addr           string
	segmentManager *storage.SegmentManager
	config         *config.Config
	BufferPool     *BufferPool
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
		BufferPool:     NewBufferPool(1024 * 1024 * 32), // 32 MB buffer pool
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

		go s.handleRequest(conn)
	}
}
