package storage

import (
	"fmt"
	"path/filepath"
)

type Segment struct {
	Log   *FileLog
	Index *Index
}

func NewSegment(dirPath string, offset int64) (*Segment, error) {
	segmentName := fmt.Sprintf("%020d", offset)
	indexFilePath := filepath.Join(dirPath, segmentName+".index")
	logFilePath := filepath.Join(dirPath, segmentName+".log")

	log, err := OpenFileLog(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	index, err := NewIndex(indexFilePath)
	if err != nil {
		log.File.Close() // Close the log file if index creation fails
		return nil, err
	}

	return &Segment{
		Log:   log,
		Index: index,
	}, nil
}
