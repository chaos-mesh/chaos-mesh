package api_server

import "fmt"

const (
	statusOK         = 200
	statusOtherError = 1
)

// Response is the body part of HTTP Response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func errResponsef(format string, args ...interface{}) *Response {
	return &Response{
		Code:    statusOtherError,
		Message: fmt.Sprintf(format, args...),
	}
}

func successResponse(data interface{}) *Response {
	return &Response{
		Code: statusOK,
		Data: data,
	}
}
