package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TODO segment recovery
type SegmentManager struct {
	DirPath        string
	Segments       []*Segment
	maxSegmentSize int64
}

// create a new segment manager
func NewSegmentManager(dirPath string) (*SegmentManager, error) {
	//create segment with offset 0 for the first ime
	segment, err := NewSegment(dirPath, 0)
	if err != nil {
		return nil, err
	}
	return &SegmentManager{
		DirPath:        dirPath,
		Segments:       []*Segment{segment},
		maxSegmentSize: 100, // TODO get segment size from config
	}, nil
}

// create a new segmment
func (sm *SegmentManager) CreateSegment(offset int64) (*Segment, error) {
	segment, err := NewSegment(sm.DirPath, offset)
	if err != nil {
		return nil, err
	}
	sm.Segments = append(sm.Segments, segment)
	return segment, nil
}

// writeRecord
func (sm *SegmentManager) WriteRecord(record *Record) (int64, error) {
	// get the last segment
	lastSegment := sm.Segments[len(sm.Segments)-1]

	if lastSegment.Log.ByteOffset+recordHeaderSize+int64(len(record.Message)) >= sm.maxSegmentSize {
		newSegment, err := sm.CreateSegment(lastSegment.NextOffset)
		if err != nil {
			return -1, err
		}
		sm.Segments = append(sm.Segments, newSegment)
		lastSegment = newSegment
	}

	offset, err := lastSegment.writeRecord(record)
	if err != nil {
		return -1, err
	}

	return offset, nil
}

// readRecords for the given offset
func (sm *SegmentManager) ReadRecord(targetOffset int64) (*Record, error) {
	//binary search to find the segment which contains the offset
	low := 0
	high := len(sm.Segments) - 1

	for low <= high {
		mid := low + (high-low)/2
		segment := sm.Segments[mid]
		if targetOffset >= segment.FirstOffset && targetOffset <= segment.NextOffset-1 {
			return segment.readRecord(targetOffset)
		} else if targetOffset < segment.FirstOffset {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return nil, fmt.Errorf("record with offset %d not found in any segment", targetOffset)
}

// when restart it will called
// look at the files in the directory and create the segments for each file
// also read the index and log file to get the first and last offset
func (sm *SegmentManager) RecoverSegments() error {
	entries, err := os.ReadDir(sm.DirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".log" {

			offsetStr := strings.TrimSuffix(entry.Name(), ".log")
			offset, err := strconv.ParseInt(offsetStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse offset from log file name: %w", err)
			}
			startingOffset := offset
			segment, err := RecoverSegment(filepath.Join(sm.DirPath, offsetStr), startingOffset)

			if err != nil {
				return fmt.Errorf("failed to create segment: %w", err)
			}

			sm.Segments = append(sm.Segments, segment)
		}
	}
	return nil
}
