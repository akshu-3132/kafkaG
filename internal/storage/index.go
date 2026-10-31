package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type Index struct {
	File                  *os.File
	LastIndexedByteOffset int64
	IndexInterval         int64
	EntryCount            int64
}

func NewIndex(FilePath string, IndexInterval int64) (*Index, error) {
	dirPath := filepath.Dir(FilePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}
	file, err := os.OpenFile(FilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	return &Index{
		File:                  file,
		LastIndexedByteOffset: 0,
		IndexInterval:         IndexInterval,
		EntryCount:            0,
	}, nil
}

func (idx *Index) ShouldIndexEntry(currentByteOffset int64) bool {
	if currentByteOffset-idx.LastIndexedByteOffset >= idx.IndexInterval {
		return true
	} else if idx.EntryCount == 0 {
		return true
	}
	return false
}

func (idx *Index) AddEntry(Bytes []byte) error {
	bytesToBeWritten := 16
	if len(Bytes) != bytesToBeWritten {
		return fmt.Errorf("Invalid index entry size")
	}
	for bytesToBeWritten > 0 {
		bytesWritten, err := idx.File.Write(Bytes[len(Bytes)-bytesToBeWritten:])
		if err != nil {
			return fmt.Errorf("Error writing index entry to file: %v\n", err)
		}
		if bytesWritten == 0 {
			return fmt.Errorf("No bytes written to index file")
		}
		bytesToBeWritten -= bytesWritten
	}
	idx.EntryCount++
	return nil
}
