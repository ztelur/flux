package http

import (
	"errors"
	"github.com/bytepowered/flux"
	"net/http"
)

var (
	ErrUnknownHttpBackendResponse = errors.New("BACKEND:UNKNOWN_HTTP_RESPONSE")
)

func NewBackendResponseCodecFunc() flux.BackendResponseCodecFunc {
	return func(ctx flux.Context, value interface{}) (*flux.BackendResponse, error) {
		resp, ok := value.(*http.Response)
		if !ok {
			return &flux.BackendResponse{
				StatusCode: http.StatusBadGateway,
				Headers:    make(http.Header, 0),
				Body:       nil,
			}, ErrUnknownHttpBackendResponse
		}
		return &flux.BackendResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	}
}
