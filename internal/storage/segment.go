package storage

import (
	"fmt"
	"kafkaG/internal/config"
	"path/filepath"
)

type Segment struct {
	Log         *FileLog
	Index       *Index
	FirstOffset int64
	NextOffset  int64
}

func NewSegment(storageConfig *config.StorageConfig, offset int64) (*Segment, error) {
	segmentName := fmt.Sprintf("%020d", offset)
	indexFilePath := filepath.Join(storageConfig.DirPath, segmentName+".index")
	logFilePath := filepath.Join(storageConfig.DirPath, segmentName+".log")

	absPath, err := filepath.Abs(storageConfig.DirPath)
	println("Absolute path of the directory:", absPath)

	log, err := OpenFileLog(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	index, err := OpenIndex(indexFilePath, storageConfig.IndexInterval)
	if err != nil {
		log.File.Close() // Close the log file if index creation fails
		return nil, err
	}

	fmt.Println("new segment created with : ")
	fmt.Println("IndexInterval : ", index.IndexInterval)

	return &Segment{
		Log:         log,
		Index:       index,
		FirstOffset: offset,
		NextOffset:  offset,
	}, nil
}

func (s *Segment) writeRecord(record *Record) (int64, error) {
	byteOffset := s.Log.ByteOffset
	err := s.Log.Append(record)
	if err != nil {
		return -1, fmt.Errorf("failed to write record: %w", err)
	}
	if s.Index.ShouldIndexEntry(s.Log.ByteOffset) {
		indexEntry := &IndexEntry{
			Offset:     s.NextOffset,
			ByteOffset: byteOffset,
		}
		if err := s.Index.AddEntry(indexEntry.encode()); err != nil {
			return -1, fmt.Errorf("failed to add index entry: %w", err)
		}

		s.Index.LastIndexedByteOffset = s.Log.ByteOffset
		s.Index.EntryCount++
	}
	currentOffset := s.NextOffset
	s.NextOffset++
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

func RecoverSegment(filePath string, startingOffset int64, IndexInterval int64) (*Segment, error) {
	logFilePath := filePath + ".log"
	indexFilePath := filePath + ".index"
	log, err := OpenFileLog(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	index, err := OpenIndex(indexFilePath, IndexInterval)
	if err != nil {
		log.File.Close() // Close the log file if index creation fails
		return nil, err
	}

	// read the index file to get the last entry
	info, err := index.File.Stat()
	if err != nil {
		log.File.Close()
		index.File.Close()
		return nil, fmt.Errorf("failed to get index file info: %w", err)
	}

	entryCount := info.Size() / 16 // each index entry is 16 bytes

	nextOffset := startingOffset
	byteOffset := int64(0)

	if entryCount > 0 {
		lastEntry, err := index.ReadEntryAt(entryCount - 1)
		if err != nil {
			log.File.Close()
			index.File.Close()
			return nil, fmt.Errorf("failed to read last index entry: %w", err)
		}

		index.LastIndexedByteOffset = lastEntry.ByteOffset

		// read the last entry to move on
		record, err := log.read(lastEntry.ByteOffset)
		if err != nil {
			log.File.Close()
			index.File.Close()
			return nil, fmt.Errorf("failed to read last record: %w", err)
		}

		byteOffset = lastEntry.ByteOffset + recordHeaderSize + int64(len(record.Message))
		nextOffset = lastEntry.Offset + 1
	}

	logInfo, err := log.File.Stat()
	if err != nil {
		log.File.Close()
		index.File.Close()
		return nil, fmt.Errorf("failed to get log file info: %w", err)
	}
	logFileSize := logInfo.Size()

	//forward scan till EOF to get last offset and byte offset
	for byteOffset < logFileSize {
		record, err := log.read(byteOffset)
		if err != nil {
			log.File.Close()
			index.File.Close()
			return nil, fmt.Errorf("failed to read record: %w", err)
		}
		byteOffset += recordHeaderSize + int64(len(record.Message))
		nextOffset++
	}

	log.ByteOffset = byteOffset
	index.EntryCount = entryCount

	return &Segment{
		Log:         log,
		Index:       index,
		FirstOffset: startingOffset,
		NextOffset:  nextOffset,
	}, nil
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
