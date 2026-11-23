// Run the write benchmarks with:
//
//	go test -bench=BenchmarkWriteRecordSweep -benchtime=200ms -run='^$' ./internal/storage/
//
// Run a single size with:
//
//	go test -bench='BenchmarkWriteRecordSweep/1KB' -benchtime=2s -run='^$' ./internal/storage/
package storage

import (
	"testing"
	"time"

	"kafkaG/internal/config"
)

func newBenchSegmentManager(b *testing.B) *SegmentManager {
	b.Helper()

	cfg := &config.StorageConfig{
		DirPath:        b.TempDir(),
		MaxSegmentSize: 64 << 20,
		IndexInterval:  4096,
		MaxMessageSize: 1000000,
	}

	sm, err := NewSegmentManager(cfg.DirPath, cfg)
	if err != nil {
		b.Fatal(err)
	}
	return sm
}

func benchWrite(b *testing.B, size int) {
	sm := newBenchSegmentManager(b)

	msg := make([]byte, size)
	for i := range msg {
		msg[i] = 'x'
	}

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := &Record{Timestamp: time.Now().Unix(), Message: msg}
		if _, err := sm.WriteRecord(rec); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteRecord(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100B", 100},
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			benchWrite(b, s.size)
		})
	}
}

func BenchmarkWriteRecordSweep(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"100B", 100},
		{"200B", 200},
		{"500B", 500},
		{"1KB", 1024},
		{"2KB", 2 * 1024},
		{"5KB", 5 * 1024},
		{"10KB", 10 * 1024},
		{"20KB", 20 * 1024},
		{"50KB", 50 * 1024},
		{"100KB", 100 * 1024},
		{"200KB", 200 * 1024},
		{"500KB", 500 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			benchWrite(b, s.size)
		})
	}
}
