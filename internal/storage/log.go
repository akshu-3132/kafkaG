package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileLog struct {
	File       *os.File
	ByteOffset int64
}

func (fl *FileLog) Append(record *Record) {
	recordBytes, err := record.Encode()
	if err != nil {
		fmt.Printf("Error encoding record: %v\n", err)
		return
	}
	bytesToWrite := len(recordBytes)
	for bytesToWrite > 0 {
		bytesWritten, err := fl.File.Write(recordBytes)
		if err != nil {
			fmt.Printf("Error writing record to file: %v\n", err)
			return
		}
		if bytesWritten == 0 {
			fmt.Println("No bytes written to file")
			return
		}
		bytesToWrite -= bytesWritten
	}
	if err != nil {
		fmt.Printf("Error writing record to file: %v\n", err)
		return
	}
	fl.ByteOffset += int64(len(recordBytes))
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
	}, nil
}

func (fl *FileLog) Close() error {
	if fl.File != nil {
		return fl.File.Close()
	}
	return nil
}
