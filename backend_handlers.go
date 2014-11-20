package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-martini/martini"
)

// GetBackends returns a list of HAProxy backends.
func GetBackends(enc Encoder, svc DataSvc) string {
	b, err := svc.GetAllBackends()
	if err != nil {
		panic(err)
	}
	return enc.EncodeMulti(b.ToInterfaces()...)
}

// GetBackend returns the requested HAProxy backend.
func GetBackend(enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	data, err := svc.GetBackend(params["name"])
	if err != nil {
		panic(err)
	}
	if data == nil {
		return util{}.notFound(enc, fmt.Sprintf("the backend with name %s does not exist", params["name"]))
	}
	return http.StatusOK, enc.Encode(data)
}

// PutBackend creates or updates an HAProxy backend.
func PutBackend(r *http.Request, enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	b := &Backend{}
	e := loadBackendFromRequest(r, enc, b)
	if e != nil {
		return util{}.badRequest(enc, "the backend data is invalid")
	}

	// always use the name identified in the resource
	b.Name = params["name"]

	existing, err := svc.GetBackend(b.Name)
	if err != nil {
		panic(err)
	}
	status := http.StatusOK
	if existing == nil {
		status = http.StatusCreated
	}

	err = svc.SaveBackend(b)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			return util{}.badRequest(enc, err.Error())
		default:
			panic(err)
		}
	}

	return status, enc.Encode(b)
}

// PostBackend performs a partial update of an existing HAProxy backend.
func PostBackend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	name := params["name"]
	b, err := svc.GetBackend(name)
	if err != nil {
		panic(err)
	}
	if b == nil {
		return util{}.notFound(enc, fmt.Sprintf("the backend with name %s does not exist", name))
	}

	e := loadBackendFromRequest(r, enc, b)
	if e != nil {
		return util{}.badRequest(enc, "the backend data is invalid")
	}

	err = svc.SaveBackend(b)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			return util{}.badRequest(enc, err.Error())
		default:
			panic(err)
		}
	}

	return http.StatusOK, enc.Encode(b)
}

// DeleteBackend removes an HAProxy backend.
func DeleteBackend(enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	key := params["name"]
	err := svc.DeleteBackend(key)
	if err != nil {
		switch err.Type {
		case ErrNotFound:
			return util{}.notFound(enc, fmt.Sprintf("the backend with name %s does not exist", key))
		default:
			panic(err)
		}
	}
	return http.StatusNoContent, ""
}

// GetBackendMembers returns a list of all members in a backend.
func GetBackendMembers(enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	b, err := svc.GetBackend(params["name"])
	if err != nil {
		panic(err)
	}
	if b == nil {
		return util{}.notFound(enc, fmt.Sprintf("the backend with name %s does not exist", params["name"]))
	}
	return http.StatusOK, enc.EncodeMulti(b.Members.ToInterfaces()...)
}

// parse request body into a Backend instance
func loadBackendFromRequest(r *http.Request, enc Encoder, b *Backend) *ErrorResponse {
	//TODO: Don't use ReadAll()... reading a terabyte of data in one go would be bad
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	err = enc.Decode(body, b)
	if err != nil {
		return NewErrorResponse(http.StatusBadRequest, fmt.Sprintf("the backend data is not valid"))
	}
	return nil
}
