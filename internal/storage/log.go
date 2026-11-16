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
	Index      *Index
	Offset     int64
}

func (fl *FileLog) Append(record *Record) (int64, error) {
	recordBytes, err := record.Encode()
	if err != nil {
		return -1, fmt.Errorf("Error encoding record: %v\n", err)
	}
	recordOffset := fl.Offset
	recordByteOffset := fl.ByteOffset
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
	fl.ByteOffset += int64(len(recordBytes))
	fl.Offset++
	if fl.Index.ShouldIndexEntry(recordByteOffset) {
		indexEntry := &IndexEntry{
			Offset:     recordOffset,
			ByteOffset: recordByteOffset,
		}
		err := fl.Index.AddEntry(indexEntry.encode())
		if err != nil {
			return -1, fmt.Errorf("Error adding index entry: %v\n", err)
		}
		fl.Index.LastIndexedByteOffset = recordByteOffset
	}
	return recordOffset, nil
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

	index, err := NewIndex(filePath + ".idx")
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create log index: %w", err)
	}

	return &FileLog{
		File:       file,
		ByteOffset: info.Size(),
		Index:      index,
		Offset:     0,
	}, nil
}

// closes the file and index file
func (fl *FileLog) Close() error {
	if fl.File != nil {
		err := fl.File.Close()
		if err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		err = fl.Index.File.Close()
		if err != nil {
			return fmt.Errorf("failed to close index file: %w", err)
		}
	}
	return nil
}

// Read the record from any given byteoffset in the log file
func (fl *FileLog) Read(byteOffset int64) (*Record, error) {
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

	const maxMessageSize = 10 * 1024 * 1024
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

// Finds the record for the given offset
// if its not present in index, scan the file from the closest index entry
func (fl *FileLog) ReadAt(offset int64) (*Record, error) {
	if offset < 0 {
		return nil, fmt.Errorf("invalid offset: %d", offset)
	}

	byteOffset := int64(0)
	currentOffset := int64(0)

	entry, exact := fl.Index.GetEntry(offset)
	if entry != nil {
		byteOffset = entry.ByteOffset
		currentOffset = entry.Offset
	}
	if exact {
		return fl.Read(byteOffset)
	}

	for {
		record, err := fl.Read(byteOffset)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("offset %d not found", offset)
			}
			return nil, err
		}
		if currentOffset == offset {
			return record, nil
		}
		byteOffset += recordHeaderSize + int64(len(record.Message))
		currentOffset++
	}
}
