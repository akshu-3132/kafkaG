package protocol

import (
	"encoding/binary"
	"fmt"
)

type ProduceRequest struct {
	Topic   string
	Message []byte
}

func (r *ProduceRequest) Encode() []byte {
	lengthTopic := len(r.Topic)
	lengthMessage := len(r.Message)
	buffer := make([]byte, 4+lengthTopic+4+lengthMessage)
	binary.BigEndian.PutUint32(buffer[:4], uint32(lengthTopic))
	copy(buffer[4:], r.Topic)
	binary.BigEndian.PutUint32(buffer[4+lengthTopic:8+lengthTopic], uint32(lengthMessage))
	copy(buffer[8+lengthTopic:], r.Message)
	return buffer
}

func DecodeProduceRequest(data []byte) (*ProduceRequest, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("Invalid produce request")
	}
	lengthTopic := binary.BigEndian.Uint32(data[:4])
	topic := string(data[4 : 4+lengthTopic])
	message := string(data[8+lengthTopic:])
	return &ProduceRequest{
		Topic:   topic,
		Message: []byte(message),
	}, nil
}
