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
	for i := 0; i < 10; i++ {
		producerequest := protocol.ProduceRequest{
			Topic:   "test-topic",
			Message: []byte(fmt.Sprintf("Hello Kafka %d", i)),
		}
		request := protocol.Request{
			APIKey:        1,
			Version:       1,
			CorrelationID: int32(i),
			Body:          producerequest.Encode(),
		}

		frame := protocol.Frame{
			Payload: request.Encode(),
		}
		if err := frame.WriteTo(conn); err != nil {
			fmt.Println("Failed to send request", err)
			return
		}
		responseFrame, err := protocol.ReadFrame(conn)
		if err != nil {
			fmt.Println("Failed to read response frame", err)
			return
		}
		response, err := protocol.DecodeResponse(responseFrame.Payload)
		if err != nil {
			fmt.Println("Failed to decode response", err)
			return
		}
		fmt.Println("Request ID:", response.CorrelationID, "ResponseId:", response.CorrelationID, "Response Body:", string(response.Body))

	}

}
