package protocol

import (
	"encoding/binary"
	"fmt"
)

type FetchRequest struct {
	Offset int64
}

func (fr *FetchRequest) Encode() ([]byte, error) {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, uint64(fr.Offset))
	return buffer, nil
}

func DecodeFetchRequest(data []byte) (*FetchRequest, error) {
	if len(data) != 8 {
		return nil, fmt.Errorf("data too short")
	}
	offset := int64(binary.BigEndian.Uint64(data))
	return &FetchRequest{
		Offset: offset,
	}, nil
}
