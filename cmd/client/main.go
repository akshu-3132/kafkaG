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

	frame := protocol.Frame{Payload: []byte("Hello, Broker!")}
	err = frame.WriteTo(conn)
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent frame to broker:", string(frame.Payload))
}
