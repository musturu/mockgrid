package objects

// ErrorResponse represents the structure of an error response.
type ErrorResponse struct {
	Errors []struct {
		Message string      `json:"message"`
		Field   interface{} `json:"field"`
		Help    interface{} `json:"help"`
	} `json:"errors"`
}

// GetErrorResponse constructs an ErrorResponse object.
func GetErrorResponse(message string, field interface{}, help interface{}) ErrorResponse {
	errorJSON := ErrorResponse{}
	e := struct {
		Message string      `json:"message"`
		Field   interface{} `json:"field"`
		Help    interface{} `json:"help"`
	}{
		message,
		field,
		help,
	}
	errorJSON.Errors = append(errorJSON.Errors, e)
	return errorJSON
}
