package main

import "fmt"

// ErrorType represents the type of an error.
type ErrorType int

const (
	// ErrConflict indicates that a record to create already exists.
	ErrConflict ErrorType = iota
	// ErrNotFound indicates that a record to update does not exist.
	ErrNotFound
	// ErrBadData indicates that the data to save is incomplete or invalid.
	ErrBadData
	// ErrSync indicates that a problem occurred when syncing the haproxy config file and the
	// requested action has been rolled back.
	ErrSync
	// ErrOutOfSync indicates that the haproxy config file is out of sync with the database.
	ErrOutOfSync
	// ErrDB indicates that a problem occurred reading or writing to the database.
	ErrDB
	// ErrUnknown indicates that an unknown error has occurred.
	ErrUnknown
)

// String returns the string representation of an ErrorType.
func (t ErrorType) String() string {
	switch t {
	case ErrConflict:
		return "ErrConflict"
	case ErrNotFound:
		return "ErrNotFound"
	case ErrBadData:
		return "ErrBadData"
	case ErrSync:
		return "ErrSync"
	case ErrOutOfSync:
		return "ErrOutOfSync"
	case ErrDB:
		return "ErrDB"
	case ErrUnknown:
		return "ErrUnknown"
	}
	return ""
}

// Error wraps an error with a type.
type Error struct {
	Type ErrorType
	error
}

// NewError returns a new Error instance that wraps the given error.
func NewError(t ErrorType, err error) *Error {
	return &Error{
		Type:  t,
		error: err,
	}
}

// NewErrorf returns a new Error instance with a message formatted according to a format specifier.
func NewErrorf(t ErrorType, format string, a ...interface{}) *Error {
	return &Error{
		Type:  t,
		error: fmt.Errorf(format, a...),
	}
}

// ErrorResponse represents a serializable error structure.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// String returns the string representation of the error.
func (e *ErrorResponse) String() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewErrorResponse returns a new Error instance from the given data.
func NewErrorResponse(code int, msg string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: msg,
	}
}
