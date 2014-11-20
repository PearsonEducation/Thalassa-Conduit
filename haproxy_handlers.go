package main

import (
	"net/http"
)

// GetHAProxyConfig returns the contents of the haproxy.cfg file
func GetHAProxyConfig(w http.ResponseWriter, enc Encoder, h HAProxy) (int, string) {
	config, err := h.GetConfig()
	if err != nil {
		return http.StatusInternalServerError, enc.Encode(NewErrorResponse(http.StatusInternalServerError, "error loading haproxy.cfg file"))
	}
	w.Header().Set("Content-Type", "text/plain")
	return http.StatusOK, config
}

// ReloadHAProxy reloads the HAProxy service
func ReloadHAProxy(w http.ResponseWriter, enc Encoder, h HAProxy) (int, string) {
	if err := h.ReloadConfig(); err != nil {
		return http.StatusInternalServerError, enc.Encode(NewErrorResponse(http.StatusInternalServerError, "error reloading HAProxy"))
	}
	w.Header().Set("Content-Type", "text/plain")
	return http.StatusOK, "HAProxy successfully reloaded"
}
