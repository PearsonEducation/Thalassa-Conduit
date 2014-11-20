package main

import (
	"fmt"
	"net/http"
	"testing"
)

// Tests that the util.badRequest() function behaves properly.
func Test_util_badRequest(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}

	msg := "baaaaad request"
	expCode := http.StatusBadRequest
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	actCode, actBody := u.badRequest(enc, msg)
	assert.Equal(t, actCode, expCode, "badRequest() returned unexpected status code")
	assert.Equal(t, actBody, expBody, "badRequest() returned unexpected body")
}

// Tests that the util.notFound() function behaves properly.
func Test_util_notFound(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}

	msg := "missing!"
	expCode := http.StatusNotFound
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	actCode, actBody := u.notFound(enc, msg)
	assert.Equal(t, actCode, expCode, "notFound() returned unexpected status code")
	assert.Equal(t, actBody, expBody, "notFound() returned unexpected body")
}

// Tests that the util.conflict() function behaves properly.
func Test_util_conflict(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}

	msg := "already exists!"
	expCode := http.StatusConflict
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	actCode, actBody := u.conflict(enc, msg)
	assert.Equal(t, actCode, expCode, "conflict() returned unexpected status code")
	assert.Equal(t, actBody, expBody, "conflict() returned unexpected body")
}
