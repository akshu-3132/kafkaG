package main

import (
	"bufio"
	"fmt"
	"kafkaG/internal/protocol"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9092")
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	var choice int64
	var correlationID int32 = 1
	for {
		fmt.Print("Enter 1 to append, 2 to get, and 0 to exit: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			return
		}
		input = strings.TrimSpace(input)
		if _, err := fmt.Sscan(input, &choice); err != nil {
			fmt.Println("Invalid input.")
			continue
		}
		switch choice {
		case 0:
			fmt.Println("Exiting...")
			return

		case 1:
			fmt.Print("Enter message to append: ")
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error:", err)
				return
			}
			message = strings.TrimRight(message, "\r\n")
			produceRequest := &protocol.ProduceRequest{
				Topic:   "test-topic",
				Message: []byte(message),
			}
			request := &protocol.Request{
				APIKey:        1,
				Version:       1,
				CorrelationID: correlationID,
				Body:          produceRequest.Encode(),
			}
			correlationID++
			frame := &protocol.Frame{
				Payload: request.Encode(),
			}
			if err := frame.WriteTo(conn); err != nil {
				fmt.Println("Failed to send request:", err)
				return
			}
			responseFrame, err := protocol.ReadFrame(conn)
			if err != nil {
				fmt.Println("Failed to read response:", err)
				return
			}
			decodedResponse, err := protocol.DecodeResponse(responseFrame.Payload)
			if err != nil {
				fmt.Println("Failed to decode response:", err)
				return
			}
			fmt.Println("Decoded response:", string(decodedResponse.Body))
		case 2:
			fmt.Print("Enter offset to fetch: ")

			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error:", err)
				return
			}
			var offset int64
			if _, err := fmt.Sscan(strings.TrimSpace(input), &offset); err != nil {
				fmt.Println("Invalid offset.")
				continue
			}
			fetchRequest := &protocol.FetchRequest{
				Offset: offset,
			}
			body := fetchRequest.Encode()
			if err != nil {
				fmt.Println("Failed to encode fetch request:", err)
				continue
			}
			request := &protocol.Request{
				APIKey:        2,
				Version:       1,
				CorrelationID: correlationID,
				Body:          body,
			}
			correlationID++
			frame := &protocol.Frame{
				Payload: request.Encode(),
			}
			if err := frame.WriteTo(conn); err != nil {
				fmt.Println("Failed to send request:", err)
				return
			}
			responseFrame, err := protocol.ReadFrame(conn)
			if err != nil {
				fmt.Println("Failed to read response:", err)
				return
			}
			decodedResponse, err := protocol.DecodeResponse(responseFrame.Payload)
			if err != nil {
				fmt.Println("Failed to decode response:", err)
				return
			}
			fmt.Println("Decoded response:", string(decodedResponse.Body))

		default:
			fmt.Println("Invalid choice. Enter 1, 2, or 0.")
		}
	}
}
