package main

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
)

// ----------------------------------------------
// GetHAProxyConfig TESTS
// ----------------------------------------------

// Tests the "happy path" for the GetHAProxyConfig() handler.
func Test_GetHAProxyConfig(t *testing.T) {
	// setup objects and mocks
	configStr := "test config"
	w := httptest.NewRecorder()
	enc := JSONEncoder{}
	h := testHelpers.NewHAProxyMock()
	h.config = configStr

	// execute function to test
	GetHAProxyConfig(w, enc, h)
	actCT := w.Header().Get("Content-Type")

	// assert return values
	assert.Equal(t, actCT, "text/plain", "GetHAProxyConfig() response has unexpected content type")
	assert.Equal(t, w.Code, 200, "GetHAProxyConfig() returned unexpected status code")
	assert.Equal(t, w.Body.String(), configStr, "GetHAProxyConfig() returned unexpected body")
}

func Test_GetHAProxyConfig_ErrorReadingConfig(t *testing.T) {
	// setup objects and mocks
	w := httptest.NewRecorder()
	enc := JSONEncoder{}
	h := testHelpers.NewHAProxyMock()
	h.getConfigAction = func() (string, error) { return "", errors.New("error") }

	// execute function to test
	GetHAProxyConfig(w, enc, h)

	// assert return values
	expCode := 500
	expBody := fmt.Sprintf(`"code":%d`, expCode)
	assert.Equal(t, w.Code, expCode, "GetHAProxyConfig() returned unexpected status code")
	assert.StringContains(t, w.Body.String(), expBody, "GetHAProxyConfig() returned unexpected body")
}

// ----------------------------------------------
// ReloadHAProxy TESTS
// ----------------------------------------------

// Tests the "happy path" for the RestartHAProxy() handler.
func Test_ReloadHAProxy(t *testing.T) {
	// setup objects and mocks
	w := httptest.NewRecorder()
	enc := JSONEncoder{}
	h := testHelpers.NewHAProxyMock()

	// execute function to test
	ReloadHAProxy(w, enc, h)
	actCT := w.Header().Get("Content-Type")

	// assert return values
	assert.Equal(t, actCT, "text/plain", "RestartHAProxy() response has unexpected content type")
	assert.Equal(t, w.Code, 200, "RestartHAProxy() returned unexpected status code")
	assert.StringContains(t, w.Body.String(), "success", "RestartHAProxy() returned unexpected body")
}
