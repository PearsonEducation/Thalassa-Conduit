package main

import (
	"net/http"
)

// GetHAProxyConfig returns the contents of the haproxy.cfg file
func GetHAProxyConfig(w http.ResponseWriter, enc Encoder, h HAProxy) {
	config, err := h.GetConfig()
	if err != nil {
		util{}.writeResponse(w, http.StatusInternalServerError,
			enc.Encode(NewErrorResponse(http.StatusInternalServerError, "error loading haproxy.cfg file")))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	util{}.writeResponse(w, http.StatusOK, config)
}

// ReloadHAProxy reloads the HAProxy service
func ReloadHAProxy(w http.ResponseWriter, enc Encoder, h HAProxy) {
	if err := h.ReloadConfig(); err != nil {
		util{}.writeResponse(w, http.StatusInternalServerError,
			enc.Encode(NewErrorResponse(http.StatusInternalServerError, "error reloading HAProxy")))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	util{}.writeResponse(w, http.StatusOK, "HAProxy successfully reloaded")
}
