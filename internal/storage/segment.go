package storage

import (
	"fmt"
	"path/filepath"
)

type Segment struct {
	Log         *FileLog
	Index       *Index
	FirstOffset int64
	LastOffset  int64
}

func NewSegment(dirPath string, offset int64) (*Segment, error) {
	segmentName := fmt.Sprintf("%020d", offset)
	indexFilePath := filepath.Join(dirPath, segmentName+".index")
	logFilePath := filepath.Join(dirPath, segmentName+".log")

	log, err := OpenFileLog(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	index, err := OpenIndex(indexFilePath)
	if err != nil {
		log.File.Close() // Close the log file if index creation fails
		return nil, err
	}

	return &Segment{
		Log:         log,
		Index:       index,
		FirstOffset: offset,
		LastOffset:  offset,
	}, nil
}

func (s *Segment) writeRecord(record *Record) (int64, error) {
	err := s.Log.Append(record)
	if err != nil {
		return -1, fmt.Errorf("failed to write record: %w", err)
	}
	if s.Index.ShouldIndexEntry(s.Log.ByteOffset) {
		indexEntry := &IndexEntry{
			Offset:     s.LastOffset,
			ByteOffset: s.Log.ByteOffset,
		}
		if err := s.Index.AddEntry(indexEntry.encode()); err != nil {
			return -1, fmt.Errorf("failed to add index entry: %w", err)
		}
		s.Index.LastIndexedByteOffset = s.Log.ByteOffset
	}
	currentOffset := s.LastOffset
	s.LastOffset++
	return currentOffset, nil
}

// readrecord from the .log file
// uses .index file to get the exact or closest offset
func (s *Segment) readRecord(offset int64) (*Record, error) {
	indexEntry, isPresent := s.Index.GetEntry(offset)
	var (
		record *Record
		err    error
	)
	if isPresent {
		record, err = s.Log.read(indexEntry.ByteOffset)
	} else {
		record, err = s.Log.closestRead(indexEntry, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %w", err)
	}
	return record, nil
}

func (s *Segment) Close() error {
	if err := s.Log.File.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	if err := s.Index.File.Close(); err != nil {
		return fmt.Errorf("failed to close index file: %w", err)
	}
	return nil
}
