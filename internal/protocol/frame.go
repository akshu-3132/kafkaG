package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Frame struct {
	Payload []byte
}

func (f *Frame) Encode() []byte {
	buffer := make([]byte, 4+len(f.Payload))
	binary.BigEndian.PutUint32(buffer[0:4], uint32(len(f.Payload)))
	copy(buffer[4:], f.Payload)
	return buffer
}

func (f *Frame) Decode(data []byte) (*Frame, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short to decode frame")
	}
	length := binary.BigEndian.Uint32(data[0:4])
	if len(data) < int(4+length) {
		return nil, fmt.Errorf("data too short for payload")
	}
	f.Payload = data[4 : 4+length]
	return f, nil

}

func (f *Frame) WriteTo(w io.Writer) error {
	data := f.Encode()
	_, err := w.Write(data)
	return err
}

func ReadFrame(r io.Reader) (*Frame, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header)
	payload := make([]byte, length)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, err
	}
	return &Frame{Payload: payload}, nil
}
