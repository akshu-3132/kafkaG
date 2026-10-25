package protocol

import (
	"encoding/binary"
	"fmt"
)

type Request struct {
	APIKey        int16
	Version       int16
	CorrelationID int32
	Body          []byte
}

func (r *Request) Encode() []byte {
	buffer := make([]byte, 2+2+4+len(r.Body))
	binary.BigEndian.PutUint16(buffer[0:2], uint16(r.APIKey))
	binary.BigEndian.PutUint16(buffer[2:4], uint16(r.Version))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(r.CorrelationID))
	copy(buffer[8:], r.Body)
	return buffer
}

func Decode(data []byte) (*Request, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}
	r := &Request{}
	r.APIKey = int16(binary.BigEndian.Uint16(data[0:2]))
	r.Version = int16(binary.BigEndian.Uint16(data[2:4]))
	r.CorrelationID = int32(binary.BigEndian.Uint32(data[4:8]))
	r.Body = data[8:]
	return r, nil
}
