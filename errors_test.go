package main

import (
	"errors"
	"fmt"
	"testing"
)

// ----------------------------------------------
// ErrorType.String TESTS
// ----------------------------------------------

func Test_ErrorType_String(t *testing.T) {
	assert.Equal(t, ErrConflict.String(), "ErrConflict", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrNotFound.String(), "ErrNotFound", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrBadData.String(), "ErrBadData", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrSync.String(), "ErrSync", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrOutOfSync.String(), "ErrOutOfSync", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrDB.String(), "ErrDB", "ErrorType.String() returned an unexpected value")
	assert.Equal(t, ErrUnknown.String(), "ErrUnknown", "ErrorType.String() returned an unexpected value")

	var ErrTest ErrorType = 99
	assert.Equal(t, ErrTest.String(), "", "ErrorType.String() returned an unexpected value")
}

// ----------------------------------------------
// NewError TESTS
// ----------------------------------------------

func Test_NewError(t *testing.T) {
	err := errors.New("test error")
	derr := NewError(ErrUnknown, err)

	assert.Equal(t, derr.Type, ErrUnknown, "Error.error returned an unexpected value")
	assert.Equal(t, derr.error, err, "Error.error returned an unexpected value")
}

// ----------------------------------------------
// NewErrorf TESTS
// ----------------------------------------------

func Test_NewErrorf(t *testing.T) {
	err := errors.New("test error")
	derr := NewErrorf(ErrUnknown, "test error")

	assert.Equal(t, derr.Type, ErrUnknown, "Error.error returned an unexpected value")
	assert.Equal(t, derr.error, err, "Error.error returned an unexpected value")
}

func Test_NewErrorf_WithPattern(t *testing.T) {
	pat := "test error: %s"
	s := "test"
	err := fmt.Errorf(pat, s)
	derr := NewErrorf(ErrUnknown, pat, s)

	assert.Equal(t, derr.Type, ErrUnknown, "Error.error returned an unexpected value")
	assert.Equal(t, derr.error, err, "Error.error returned an unexpected value")
}

// Tests that the NewErrorResponse() function behaves properly.
func Test_NewErrorResponse(t *testing.T) {
	c := 500
	m := "an error occurred"
	expected := &ErrorResponse{Code: c, Message: m}
	actual := NewErrorResponse(c, m)
	assert.Equal(t, *actual, *expected, "NewErrorResponse() return unexpected value")
}

// Tests that the ErrorResponse.String() function behaves properly.
func Test_ErrorResponse_String(t *testing.T) {
	c := 500
	m := "an error occurred"
	e := NewErrorResponse(c, m)
	expected := fmt.Sprintf("[%d] %s", c, m)
	actual := e.String()
	assert.Equal(t, actual, expected, "ErrorResponse.String() return unexpected value")
}
