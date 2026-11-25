package server

import (
	"fmt"
	"math/bits"
	"sync"
)

// It will contain buffer of different sizes in the power of 2
// it will have minimum baseline number of buffers for each size

// TODO add some baseline buffers to avoid warmups
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

	if size > 16*1024*1024 || size <= 0 {
		return nil
	}

	var nearestPowerOfTwo, idx int

	if size <= 4*1024 {
		nearestPowerOfTwo = 4 * 1024
		idx = 0
	} else {
		nearestPowerOfTwo = 1 << (bits.Len(uint(size - 1)))
		idx = bits.TrailingZeros(uint(nearestPowerOfTwo))
		idx = idx - 12 // 4KB = 0
	}

	bp.mu[idx].Lock()
	defer bp.mu[idx].Unlock()

	//TODO add functionality to check whether buffer can be allocated or not
	//! ie deleting the buffer from memory
	if len(bp.Buffers[idx]) == 0 {
		buffer := bp.allocateBuffer(nearestPowerOfTwo)
		if buffer == nil {
			return nil
		}
		return buffer[:size] // Return a slice of the requested size
	}

	lastIdx := len(bp.Buffers[idx]) - 1
	buffer := bp.Buffers[idx][lastIdx]

	bp.Buffers[idx][lastIdx] = nil
	bp.Buffers[idx] = bp.Buffers[idx][:lastIdx]

	return buffer[:size] // Return a slice of the requested size
}

func (bp *BufferPool) allocateBuffer(size int) []byte {
	if size < 4*1024 || size > 16*1024*1024 {
		return nil
	}

	bp.bytesmu.Lock()
	defer bp.bytesmu.Unlock()

	if bp.currentBytesUsed+int64(size) > bp.maxBytesAllowed {
		return nil
	}

	fmt.Println("becfore bytes used in KB ", bp.currentBytesUsed/1024)
	bp.currentBytesUsed += int64(size)
	fmt.Println("after bytes used : ", bp.currentBytesUsed/1024)
	return make([]byte, size)
}

func (bp *BufferPool) ReturnBuffer(buffer []byte) {
	bufferCap := cap(buffer)

	if bufferCap < 4*1024 || bufferCap > 16*1024*1024 {
		return
	}

	if bufferCap&(bufferCap-1) != 0 {
		return
	}

	idx := bits.TrailingZeros(uint(bufferCap))
	idx = idx - 12 // 4KB = 0

	buffer = buffer[:bufferCap] // Reset the slice to its full capacity

	bp.mu[idx].Lock()
	defer bp.mu[idx].Unlock()
	bp.Buffers[idx] = append(bp.Buffers[idx], buffer)
}
