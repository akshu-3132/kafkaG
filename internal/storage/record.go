package storage

import (
	"encoding/binary"
	"fmt"
)

const recordHeaderSize = 12

type Record struct {
	Timestamp int64
	Message   []byte
}

func (r *Record) Encode() ([]byte, error) {
	if uint64(len(r.Message)) > uint64(^uint32(0)) {
		return nil, fmt.Errorf("message too large")
	}
	if len(r.Message) <= 0 {
		return nil, fmt.Errorf("message is empty")
	}
	buffer := make([]byte, recordHeaderSize+len(r.Message))
	binary.BigEndian.PutUint64(buffer[0:8], uint64(r.Timestamp))
	binary.BigEndian.PutUint32(buffer[8:12], uint32(len(r.Message)))
	copy(buffer[12:], r.Message)

	return buffer, nil
}

func DecodeRecord(data []byte) (*Record, error) {
	if len(data) < recordHeaderSize {
		return nil, fmt.Errorf("data too short")
	}
	messageLength := uint64(binary.BigEndian.Uint32(data[8:12]))
	if messageLength != uint64(len(data)-recordHeaderSize) {
		return nil, fmt.Errorf("invalid message length")
	}
	record := &Record{
		Timestamp: int64(binary.BigEndian.Uint64(data[0:8])),
		Message:   make([]byte, int(messageLength)),
	}
	copy(record.Message, data[recordHeaderSize:])
	return record, nil
}
