package protocol

import (
	"encoding/binary"
	"fmt"
)

type Response struct {
	CorrelationID int32
	ErrorCode     int16
	Body          []byte
}

func (r *Response) Encode() ([]byte, error) {
	buffer := make([]byte, 4+2+len(r.Body))
	binary.BigEndian.PutUint32(buffer[0:4], uint32(r.CorrelationID))
	binary.BigEndian.PutUint16(buffer[4:6], uint16(r.ErrorCode))
	copy(buffer[6:], r.Body)
	return buffer, nil
}

func DecodeResponse(data []byte) (*Response, error) {
	if len(data) < 6 {
		return nil, fmt.Errorf("data too short")
	}
	r := Response{
		CorrelationID: int32(binary.BigEndian.Uint32(data[0:4])),
		ErrorCode:     int16(binary.BigEndian.Uint16(data[4:6])),
		Body:          data[6:],
	}
	return &r, nil
}
