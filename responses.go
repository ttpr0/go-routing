package main

type ErrorResponse struct {
	Request string `json:"request"`
	Error   any    `json:"error"`
}

func NewErrorResponse(request string, error any) ErrorResponse {
	return ErrorResponse{
		Request: request,
		Error:   error,
	}
}
