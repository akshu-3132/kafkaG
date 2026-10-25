package main

import (
	"fmt"
	"kafkaG/internal/protocol"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9092")
	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	defer conn.Close()
	messages := []string{
		"Hello, Broker!",
		"This is a test message.",
		"Goodbye, Broker!",
	}
	for _, msg := range messages {
		frame := protocol.Frame{Payload: []byte(msg)}
		if err := frame.WriteTo(conn); err != nil {
			fmt.Println("Failed to send message:", err)
			return
		}
	}
}
