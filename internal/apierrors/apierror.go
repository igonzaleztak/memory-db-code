/*
The package apierrors contains the ApiError struct that represents an error that should be returned to the client.

The errors are defined following the next structure:
- Code: a known code used to recognize easily the error.
- Message: a human readable message that describes the error.
- SysMessage: the system message that describes the error. This field is only used for logging purposes, and it is not returned to the client.
- HTTPStatus: the HTTP status code that should be returned.

	type ApiError struct {
		Code       string `json:"code"`    // Code is a known code used to recognize easily the error.
		Message    string `json:"message"` // Message is a human readable message that describes the error.
		SysMessage string `json:"-"`       // SysMessage is the system message that describes the error. This field is only used for logging purposes, and it is not returned to the client.
		HTTPStatus int    `json:"-"`       // HTTPStatus is the HTTP status code that should be returned.
	}
*/
package apierrors

// ApiError is a struct that represents an error that should be returned to the client.
type ApiError struct {
	Code       string `json:"code"`    // Code is a known code used to recognize easily the error.
	Message    string `json:"message"` // Message is a human readable message that describes the error.
	SysMessage string `json:"-"`       // SysMessage is the system message that describes the error. This field is only used for logging purposes, and it is not returned to the client.
	HTTPStatus int    `json:"-"`       // HTTPStatus is the HTTP status code that should be returned.
}

// NewAPIError creates a new ApiError instance.
func NewAPIError(code string, message string, httpStatus int) *ApiError {
	return &ApiError{
		Code:       code,
		Message:    message,
		SysMessage: message, // by default, the system message is the same as the message
		HTTPStatus: httpStatus,
	}
}

// Error returns the error message. This method is required to implement the error interface.
// This way we can return api errors as regular errors.
func (e *ApiError) Error() string {
	return e.Message
}
