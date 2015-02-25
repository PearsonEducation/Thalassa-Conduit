package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_util_writeResponse(t *testing.T) {
	u := util{}
	rw := httptest.NewRecorder()

	expCode := http.StatusGone
	expBody := fmt.Sprintf(`{"code":%d,"message":"gone!!"}`, expCode)

	u.writeResponse(rw, expCode, expBody)
	assert.Equal(t, rw.Code, expCode, "badRequest() returned unexpected status code")
	assert.Equal(t, rw.Body.String(), expBody, "badRequest() returned unexpected body")
}

// Tests that the util.badRequest() function behaves properly.
func Test_util_badRequest(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}
	rw := httptest.NewRecorder()

	msg := "baaaaad request"
	expCode := http.StatusBadRequest
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	u.badRequest(rw, enc, msg)
	assert.Equal(t, rw.Code, expCode, "badRequest() returned unexpected status code")
	assert.Equal(t, rw.Body.String(), expBody, "badRequest() returned unexpected body")
}

// Tests that the util.notFound() function behaves properly.
func Test_util_notFound(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}
	rw := httptest.NewRecorder()

	msg := "missing!"
	expCode := http.StatusNotFound
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	u.notFound(rw, enc, msg)
	assert.Equal(t, rw.Code, expCode, "notFound() returned unexpected status code")
	assert.Equal(t, rw.Body.String(), expBody, "notFound() returned unexpected body")
}

// Tests that the util.conflict() function behaves properly.
func Test_util_conflict(t *testing.T) {
	u := util{}
	enc := JSONEncoder{}
	rw := httptest.NewRecorder()

	msg := "already exists!"
	expCode := http.StatusConflict
	expBody := fmt.Sprintf(`{"code":%d,"message":"%s"}`, expCode, msg)

	u.conflict(rw, enc, msg)
	assert.Equal(t, rw.Code, expCode, "conflict() returned unexpected status code")
	assert.Equal(t, rw.Body.String(), expBody, "conflict() returned unexpected body")
}
