package storage

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type FileLog struct {
	File       *os.File
	ByteOffset int64
	Index      *Index
}

func (fl *FileLog) Append(record *Record) (int64, error) {
	recordBytes, err := record.Encode()
	if err != nil {
		return -1, fmt.Errorf("Error encoding record: %v\n", err)
	}
	bytesToWrite := len(recordBytes)
	for bytesToWrite > 0 {
		// Writing only remaining bytes to the file to avoid partial writes
		bytesWritten, err := fl.File.Write(recordBytes[len(recordBytes)-bytesToWrite:])
		if err != nil {
			return -1, fmt.Errorf("Error writing record to file: %v\n", err)
		}
		if bytesWritten == 0 {

			return -1, fmt.Errorf("No bytes written to file")
		}
		bytesToWrite -= bytesWritten
	}
	logicalOffset := fl.Index.AddEntry(fl.ByteOffset)
	fl.ByteOffset += int64(len(recordBytes))
	return logicalOffset, nil
}

func OpenFileLog(filePath string) (*FileLog, error) {
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	file, err := os.OpenFile(
		filePath,
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return &FileLog{
		File:       file,
		ByteOffset: info.Size(),
		Index:      NewIndex(),
	}, nil
}

func (fl *FileLog) Close() error {
	if fl.File != nil {
		return fl.File.Close()
	}
	return nil
}

func (fl *FileLog) ReadAt(logicalOffset int64) (*Record, error) {
	byteOffset, ok := fl.Index.Get(logicalOffset)
	if !ok {
		return nil, fmt.Errorf("invalid logical offset: %d", logicalOffset)
	}
	if fl.File == nil {
		return nil, fmt.Errorf("file is not open")
	}
	if byteOffset < 0 {
		return nil, fmt.Errorf("invalid byte offset: %d", byteOffset)
	}
	header := make([]byte, recordHeaderSize)
	_, err := fl.File.ReadAt(header, byteOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read record header: %w", err)
	}
	messageSize := uint64(binary.BigEndian.Uint32(header[8:12]))

	const maxMessageSize = 10 * 1024 * 1024
	if messageSize > maxMessageSize {
		return nil, fmt.Errorf("record message too large: %d bytes", messageSize)
	}
	recordBuffer := make([]byte, recordHeaderSize+int(messageSize))
	if messageSize > 0 {
		_, err := fl.File.ReadAt(recordBuffer, byteOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read record payload: %w", err)
		}
	}
	record, err := DecodeRecord(recordBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %w", err)
	}
	return record, nil
}
