package main

import "net/http"

type util struct{}

func (util) badRequest(enc Encoder, err string) (int, string) {
	return http.StatusBadRequest, enc.Encode(NewErrorResponse(http.StatusBadRequest, err))
}

func (util) notFound(enc Encoder, err string) (int, string) {
	return http.StatusNotFound, enc.Encode(NewErrorResponse(http.StatusNotFound, err))
}

func (util) conflict(enc Encoder, err string) (int, string) {
	return http.StatusConflict, enc.Encode(NewErrorResponse(http.StatusConflict, err))
}
