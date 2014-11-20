package main

import (
	"net/http"
	"os"
	"testing"
	"text/template"
	"time"
)

// Tests that the GetStatus() handler behaves properly.
func Test_GetStatus(t *testing.T) {
	actCode, actBody := GetStatus()

	expCode := http.StatusOK
	expBody := `{"status":"ok"}`
	assert.Equal(t, actCode, expCode, "GetStatus() returned unexpected status code")
	assert.Equal(t, actBody, expBody, "GetStatus() returned unexpected body")
}

// Tests that the initMartini() function behaves properly.
func Test_initMartini(t *testing.T) {
	c := &Config{
		HAConfigPath: "test-fixtures/haproxy.cfg",
		DBPath:       "/var/db/conduit",
	}

	d, _ := time.ParseDuration(defaultTimeout)
	signalChan := make(chan os.Signal, 1)
	server := NewServer(d, signalChan)

	db := testHelpers.NewDBManagerMock()
	tmpl, _ := template.New("test").Parse("")
	m := initMartini(server, c, db, tmpl)
	assert.NotNil(t, m, "initMartini() returned nil Martini instance")
	//TODO: add logic to properly test that martini was initialized correctly
}
