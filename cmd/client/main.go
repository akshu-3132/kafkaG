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
	request := protocol.Request{
		APIKey:        0,
		Version:       0,
		CorrelationID: 0,
		Body:          []byte("Hello, Broker!"),
	}
	frame := protocol.Frame{
		Payload: request.Encode(),
	}
	if err := frame.WriteTo(conn); err != nil {
		fmt.Println("Failed to send frame:", err)
		return
	}
	ResponseFrame, err := protocol.ReadFrame(conn)
	if err != nil {
		fmt.Println("Failed to read frame:", err)
		return
	}
	response, err := protocol.DecodeResponse(ResponseFrame.Payload)
	if err != nil {
		fmt.Println("Failed to decode response:", err)
		return
	}
	fmt.Println("Received response from server:", string(response.Body))

}
