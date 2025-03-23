package response

// ErrorResponse provides a standardized error format
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, code string) Response {
	return Response{
		Success: false,
		Error:   err,
		Code:    code,
	}
}
