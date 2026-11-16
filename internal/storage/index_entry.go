package storage

import (
	"encoding/binary"
	"fmt"
)

type IndexEntry struct {
	Offset     int64
	ByteOffset int64
}

func (idx *IndexEntry) encode() []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], uint64(idx.Offset))
	binary.BigEndian.PutUint64(buf[8:16], uint64(idx.ByteOffset))
	return buf
}

func decodeIndexEntry(data []byte) (*IndexEntry, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("invalid index entry data length")
	}
	offset := int64(binary.BigEndian.Uint64(data[0:8]))
	byteOffset := int64(binary.BigEndian.Uint64(data[8:16]))
	return &IndexEntry{
		Offset:     offset,
		ByteOffset: byteOffset,
	}, nil
}
