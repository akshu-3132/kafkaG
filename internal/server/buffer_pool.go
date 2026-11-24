package server

import (
	"math/bits"
	"sync"
)

// It will contain buffer of different sizes in the power of 2
// it will have minimum baseline number of buffers for each size

type BufferPool struct {
	mu               [13]sync.Mutex
	bytesmu          sync.Mutex
	Buffers          [13][][]byte // 4KB = 0 , 8KB = 1.... upto 16MB
	currentBytesUsed int64
	maxBytesAllowed  int64
}

func NewBufferPool(maxBytesAllowed int64) *BufferPool {
	return &BufferPool{
		currentBytesUsed: 0,
		maxBytesAllowed:  maxBytesAllowed,
	}
}

func (bp *BufferPool) GetBuffer(size int) []byte {

	idx := bits.TrailingZeros(uint(size))
	idx = idx - 2

	if size < 4 || size > 16*1024 {
		return nil
	}

	bp.mu[idx].Lock()
	defer bp.mu[idx].Unlock()

	if len(bp.Buffers[idx]) == 0 {
		return bp.allocateBuffer(size)
	}

	lastIdx := len(bp.Buffers[idx]) - 1
	buffer := bp.Buffers[idx][lastIdx]

	bp.Buffers[idx][lastIdx] = nil
	bp.Buffers[idx] = bp.Buffers[idx][:lastIdx]

	return buffer
}

func (bp *BufferPool) allocateBuffer(size int) []byte {
	if size < 4 || size > 16*1024 {
		return nil
	}

	bp.bytesmu.Lock()
	defer bp.bytesmu.Unlock()

	if bp.currentBytesUsed+int64(size)*1024 > bp.maxBytesAllowed {
		return nil
	}

	bp.currentBytesUsed += int64(size) * 1024
	return make([]byte, size*1024)
}
