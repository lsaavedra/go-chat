package api

import "fmt"

// APIError represents any error occurred during request processing,
// it holds the HTTP status code to be replied by the server.
type APIError struct {
	HTTPStatusCode int `json:"status_code"`
	// optional if Cause is present.
	Msg string `json:"message"`
	// optional if Msg is present.
	Cause error `json:"cause,omitempty"`
}

// ErrorResponse is used as the error response API.
// swagger:model
type ErrorResponse struct {
	Msg   string `json:"message"`
	Cause string `json:"cause,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s, httpStatusCode %d: %v", e.Msg, e.HTTPStatusCode, e.Cause)
}

// ErrorResponse maps APIError to ErrorResponse.
// If Msg is empty, Cause is used as Msg.
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
