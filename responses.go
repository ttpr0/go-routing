package main

type ErrorResponse struct {
	Request string `json:"request"`
	Error   string `json:"error"`
}

func NewErrorResponse(request, error string) ErrorResponse {
	return ErrorResponse{
		Request: request,
		Error:   error,
	}
}
