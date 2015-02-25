package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetBackends returns a list of HAProxy backends.
func GetBackends(w http.ResponseWriter, enc Encoder, svc DataSvc) {
	b, err := svc.GetAllBackends()
	if err != nil {
		panic(err)
	}
	util{}.writeResponse(w, http.StatusOK, enc.EncodeMulti(b.ToInterfaces()...))
}

// GetBackend returns the requested HAProxy backend.
func GetBackend(w http.ResponseWriter, enc Encoder, svc DataSvc, params Params) {
	data, err := svc.GetBackend(params["name"])
	if err != nil {
		panic(err)
	}
	if data == nil {
		util{}.notFound(w, enc, fmt.Sprintf("the backend with name %s does not exist", params["name"]))
		return
	}
	util{}.writeResponse(w, http.StatusOK, enc.Encode(data))
}

// PutBackend creates or updates an HAProxy backend.
func PutBackend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params Params) {
	b := &Backend{}
	e := loadBackendFromRequest(r, enc, b)
	if e != nil {
		util{}.badRequest(w, enc, "the backend data is invalid")
		return
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
			util{}.badRequest(w, enc, err.Error())
			return
		default:
			panic(err)
		}
	}

	util{}.writeResponse(w, status, enc.Encode(b))
}

// PostBackend performs a partial update of an existing HAProxy backend.
func PostBackend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params Params) {
	name := params["name"]
	b, err := svc.GetBackend(name)
	if err != nil {
		panic(err)
	}
	if b == nil {
		util{}.notFound(w, enc, fmt.Sprintf("the backend with name %s does not exist", name))
		return
	}

	e := loadBackendFromRequest(r, enc, b)
	if e != nil {
		util{}.badRequest(w, enc, "the backend data is invalid")
		return
	}

	err = svc.SaveBackend(b)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			util{}.badRequest(w, enc, err.Error())
			return
		default:
			panic(err)
		}
	}

	util{}.writeResponse(w, http.StatusOK, enc.Encode(b))
}

// DeleteBackend removes an HAProxy backend.
func DeleteBackend(w http.ResponseWriter, enc Encoder, svc DataSvc, params Params) {
	key := params["name"]
	err := svc.DeleteBackend(key)
	if err != nil {
		switch err.Type {
		case ErrNotFound:
			util{}.notFound(w, enc, fmt.Sprintf("the backend with name %s does not exist", key))
			return
		default:
			panic(err)
		}
	}
	util{}.writeResponse(w, http.StatusNoContent, "")
}

// GetBackendMembers returns a list of all members in a backend.
func GetBackendMembers(w http.ResponseWriter, enc Encoder, svc DataSvc, params Params) {
	b, err := svc.GetBackend(params["name"])
	if err != nil {
		panic(err)
	}
	if b == nil {
		util{}.notFound(w, enc, fmt.Sprintf("the backend with name %s does not exist", params["name"]))
		return
	}
	util{}.writeResponse(w, http.StatusOK, enc.EncodeMulti(b.Members.ToInterfaces()...))
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
