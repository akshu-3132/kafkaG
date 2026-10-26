package server

import (
	"fmt"
	"kafkaG/internal/protocol"
)

const (
	APIKeyProduce int16 = 1

	ErrorCodeNone           int16 = 0
	ErrorCodeInvalidRequest int16 = 1
	ErrorCodeUnsupportedAPI int16 = 2
)

func (s *Server) handleRequest(
	req *protocol.Request,
) (*protocol.Response, error) {
	resp := &protocol.Response{
		CorrelationID: req.CorrelationID,
		ErrorCode:     ErrorCodeNone,
	}

	switch req.APIKey {
	case APIKeyProduce:
		_, err := protocol.DecodeProduceRequest(req.Body)
		if err != nil {
			resp.ErrorCode = ErrorCodeInvalidRequest
			return resp, nil
		}
		// fmt.Println("Message : " + string(req.Body))
		return resp, nil

	default:
		resp.ErrorCode = ErrorCodeUnsupportedAPI
		return resp, fmt.Errorf("unsupported API key: %d", req.APIKey)
	}
}
