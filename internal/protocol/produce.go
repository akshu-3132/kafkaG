package protocol

import "encoding/binary"

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
