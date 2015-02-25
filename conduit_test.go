package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests that the GetStatus() handler behaves properly.
func Test_GetStatus(t *testing.T) {
	rw := httptest.NewRecorder()
	GetStatus(rw)

	expCode := http.StatusOK
	expBody := `{"status":"ok"}`
	assert.Equal(t, rw.Code, expCode, "GetStatus() returned unexpected status code")
	assert.Equal(t, rw.Body.String(), expBody, "GetStatus() returned unexpected body")
}
