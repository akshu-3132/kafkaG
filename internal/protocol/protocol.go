package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BufferPool interface {
	GetBuffer(size int) []byte
	ReturnBuffer(buffer []byte)
}

// read the bytes from the conn object
func Read(r io.Reader, header []byte, bufferPool BufferPool) ([]byte, error) {
	_, err := io.ReadFull(r, header[:])
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("client disconnected")
		} else {
			return nil, fmt.Errorf("error reading frame")
		}
	}

	bufferSize := int(binary.BigEndian.Uint32(header[:]))
	buffer := bufferPool.GetBuffer(bufferSize)

	if buffer == nil {
		return nil, fmt.Errorf("Fail to get buffer")
	}

	bytesRead, err := io.ReadFull(r, buffer)
	if err != nil || bytesRead != len(buffer) {
		bufferPool.ReturnBuffer(buffer)
		return nil, fmt.Errorf("Payload size mismatch")
	}

	return buffer, nil

}

// write the bytes to the conn object
func Write(w io.Writer, buffer []byte) error {
	bytesWritten := 0
	sizeBuffer := len(buffer)
	totalBytesWritten := 0
	var err error
	for totalBytesWritten < sizeBuffer {
		bytesWritten, err = w.Write(buffer[totalBytesWritten:])
		if err != nil {
			return err
		}
		if bytesWritten == 0 {
			return io.ErrShortWrite
		}
		totalBytesWritten += bytesWritten
	}
	return nil
}
