package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileLog struct {
	File       *os.File
	ByteOffset int64
}

// appends a record to the end fo the log file

func (fl *FileLog) Append(record *Record) error {
	// TODO update this method to handle the partial failure write, needs to rollback i guess
	bbuffer, err := record.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode record: %w", err)
	}
	bytesToBeWritten := len(bbuffer)
	for bytesToBeWritten > 0 {
		bytesWritten, err := fl.File.Write(bbuffer[len(bbuffer)-bytesToBeWritten:])
		if err != nil {
			return fmt.Errorf("failed to write record to log file: %w", err)
		}
		if bytesWritten == 0 {
			return fmt.Errorf("no bytes written to log file")
		}
		bytesToBeWritten -= bytesWritten
	}
	fl.ByteOffset += int64(len(bbuffer))
	return nil
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

	return &FileLog{
		File:       file,
		ByteOffset: 0,
	}, nil
}

// closes the file and index file
func (fl *FileLog) Close() error {
	if fl.File != nil {
		err := fl.File.Close()
		if err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}
	return nil
}

// Read the record from any given byteoffset in the log file
func (fl *FileLog) read(byteOffset int64) (*Record, error) {
	if fl.File == nil {
		return nil, fmt.Errorf("file is not open")
	}
	if byteOffset < 0 {
		return nil, fmt.Errorf("invalid byte offset: %d", byteOffset)
	}

	header := make([]byte, recordHeaderSize)
	if _, err := fl.File.ReadAt(header, byteOffset); err != nil {
		return nil, fmt.Errorf("failed to read record header: %w", err)
	}
	messageSize := uint64(binary.BigEndian.Uint32(header[8:12]))

	const maxMessageSize = 10 * 1024 * 1024 //TODO add message size from config
	if messageSize > maxMessageSize {
		return nil, fmt.Errorf("record message too large: %d bytes", messageSize)
	}

	recordBuffer := make([]byte, recordHeaderSize+int(messageSize))
	if _, err := fl.File.ReadAt(recordBuffer, byteOffset); err != nil {
		return nil, fmt.Errorf("failed to read record payload: %w", err)
	}

	record, err := DecodeRecord(recordBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to decode record: %w", err)
	}

	return record, nil
}

// closestRead reads the record with the closest offset to the targetOffset.
func (fl *FileLog) closestRead(closestIndexEntry *IndexEntry, targetOffset int64) (*Record, error) {
	if targetOffset < 0 {
		return nil, fmt.Errorf("invalid offset: %d", targetOffset)
	}
	byteOffset := closestIndexEntry.ByteOffset
	currentOffset := closestIndexEntry.Offset
	if closestIndexEntry == nil {
		byteOffset = 0
		currentOffset = 0
	}
	for {
		record, err := fl.read(byteOffset)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("offset %d not found", targetOffset)
			}
			return nil, err
		}
		if currentOffset == targetOffset {
			return record, nil
		}
		byteOffset += recordHeaderSize + int64(len(record.Message))
		currentOffset++
	}
}
