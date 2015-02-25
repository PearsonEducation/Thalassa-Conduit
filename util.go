package main

import "net/http"

type util struct{}

func (u util) badRequest(w http.ResponseWriter, enc Encoder, err string) {
	u.writeResponse(w, http.StatusBadRequest, enc.Encode(NewErrorResponse(http.StatusBadRequest, err)))
}

func (u util) notFound(w http.ResponseWriter, enc Encoder, err string) {
	u.writeResponse(w, http.StatusNotFound, enc.Encode(NewErrorResponse(http.StatusNotFound, err)))
}

func (u util) conflict(w http.ResponseWriter, enc Encoder, err string) {
	u.writeResponse(w, http.StatusConflict, enc.Encode(NewErrorResponse(http.StatusConflict, err)))
}

func (util) writeResponse(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	w.Write([]byte(body))
}
