package storage

import "fmt"

type SegmentManager struct {
	DirPath  string
	Segments []*Segment
}

// create a new segment manager
func NewSegmentManager(dirPath string) (*SegmentManager, error) {
	//create segment with offset 0 for the first ime
	segment, err := NewSegment(dirPath, 0)
	if err != nil {
		return nil, err
	}
	return &SegmentManager{
		DirPath:  dirPath,
		Segments: []*Segment{segment},
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
func (sm *SegmentManager) writeRecord(record *Record) (int64, error) {
	// get the last segment
	lastSegment := sm.Segments[len(sm.Segments)-1]
	offset, err := lastSegment.writeRecord(record)
	if err != nil {
		return -1, err
	}
	return offset, nil
}

// readRecords for the given offset
func (sm *SegmentManager) readRecord(targetOffset int64) (*Record, error) {
	//binary search to find the segment which contains the offset
	low := 0
	high := len(sm.Segments) - 1

	for low <= high {
		mid := low + (high-low)/2
		segment := sm.Segments[mid]
		if targetOffset >= segment.FirstOffset && targetOffset <= segment.LastOffset-1 {
			return segment.readRecord(targetOffset)
		} else if targetOffset < segment.FirstOffset {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return nil, fmt.Errorf("record with offset %d not found in any segment", targetOffset)
}
