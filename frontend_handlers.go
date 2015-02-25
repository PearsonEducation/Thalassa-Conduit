package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetFrontends returns a list of HAProxy frontends.
func GetFrontends(w http.ResponseWriter, enc Encoder, svc DataSvc) {
	f, err := svc.GetAllFrontends()
	if err != nil {
		panic(err)
	}
	util{}.writeResponse(w, http.StatusOK, enc.EncodeMulti(f.ToInterfaces()...))
}

// GetFrontend returns the requested HAProxy frontend.
func GetFrontend(w http.ResponseWriter, enc Encoder, svc DataSvc, params Params) {
	data, err := svc.GetFrontend(params["name"])
	if err != nil {
		panic(err)
	}
	if data == nil {
		util{}.notFound(w, enc, fmt.Sprintf("the frontend with name %s does not exist", params["name"]))
		return
	}
	util{}.writeResponse(w, http.StatusOK, enc.Encode(data))
}

// PutFrontend creates or updates an HAProxy frontend.
func PutFrontend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params Params) {
	f := &Frontend{}
	e := loadFrontendFromRequest(r, enc, f)
	if e != nil {
		util{}.badRequest(w, enc, "the frontend data is invalid")
		return
	}

	// always use the name identified in the resource URL
	f.Name = params["name"]

	existing, err := svc.GetFrontend(f.Name)
	if err != nil {
		panic(err)
	}
	status := http.StatusOK
	if existing == nil {
		status = http.StatusCreated
	}

	err = svc.SaveFrontend(f)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			util{}.badRequest(w, enc, err.Error())
			return
		default:
			panic(err)
		}
	}

	util{}.writeResponse(w, status, enc.Encode(f))
}

// PostFrontend performs a partial update of an existing HAProxy frontend.
func PostFrontend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params Params) {
	name := params["name"]
	f, err := svc.GetFrontend(name)
	if err != nil {
		panic(err)
	}
	if f == nil {
		util{}.notFound(w, enc, fmt.Sprintf("the frontend with name %s does not exist", name))
		return
	}

	e := loadFrontendFromRequest(r, enc, f)
	if e != nil {
		util{}.badRequest(w, enc, "the frontend data is invalid")
		return
	}

	err = svc.SaveFrontend(f)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			util{}.badRequest(w, enc, err.Error())
			return
		default:
			panic(err)
		}
	}

	util{}.writeResponse(w, http.StatusOK, enc.Encode(f))
}

// DeleteFrontend removes an HAProxy frontend.
func DeleteFrontend(w http.ResponseWriter, enc Encoder, svc DataSvc, params Params) {
	key := params["name"]
	err := svc.DeleteFrontend(key)
	if err != nil {
		switch err.Type {
		case ErrNotFound:
			util{}.notFound(w, enc, fmt.Sprintf("the frontend with name %s does not exist", key))
			return
		default:
			panic(err)
		}
	}
	util{}.writeResponse(w, http.StatusNoContent, "")
}

// parse request body into a Frontend instance
func loadFrontendFromRequest(r *http.Request, enc Encoder, f *Frontend) *ErrorResponse {
	// TODO: Don't use ReadAll()... reading a terabyte of data in one go would be bad
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	err = enc.Decode(body, f)
	if err != nil {
		return NewErrorResponse(http.StatusBadRequest, fmt.Sprintf("the frontend data is not valid"))
	}
	return nil
}
