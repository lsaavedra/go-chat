package api

import "fmt"

type APIError struct {
	HTTPStatusCode int    `json:"status_code"`
	Msg            string `json:"message"`
	Cause          error  `json:"cause,omitempty"`
}

type ErrorResponse struct {
	Msg   string `json:"message"`
	Cause string `json:"cause,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s, httpStatusCode %d: %v", e.Msg, e.HTTPStatusCode, e.Cause)
}

func (e *APIError) ErrorResponse() ErrorResponse {
	var cause string
	if e.Cause != nil {
		cause = e.Cause.Error()
	}

	msg := e.Msg
	if msg == "" && cause != "" {
		msg = cause
		cause = ""
	}

	return ErrorResponse{
		Msg:   msg,
		Cause: cause,
	}
}
