package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-martini/martini"
)

// GetFrontends returns a list of HAProxy frontends.
func GetFrontends(enc Encoder, svc DataSvc) string {
	f, err := svc.GetAllFrontends()
	if err != nil {
		panic(err)
	}
	return enc.EncodeMulti(f.ToInterfaces()...)
}

// GetFrontend returns the requested HAProxy frontend.
func GetFrontend(enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	data, err := svc.GetFrontend(params["name"])
	if err != nil {
		panic(err)
	}
	if data == nil {
		return util{}.notFound(enc, fmt.Sprintf("the frontend with name %s does not exist", params["name"]))
	}
	return http.StatusOK, enc.Encode(data)
}

// PutFrontend creates or updates an HAProxy frontend.
func PutFrontend(r *http.Request, enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	f := &Frontend{}
	e := loadFrontendFromRequest(r, enc, f)
	if e != nil {
		return util{}.badRequest(enc, "the frontend data is invalid")
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
			return util{}.badRequest(enc, err.Error())
		default:
			panic(err)
		}
	}

	return status, enc.Encode(f)
}

// PostFrontend performs a partial update of an existing HAProxy frontend.
func PostFrontend(w http.ResponseWriter, r *http.Request, enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	name := params["name"]
	f, err := svc.GetFrontend(name)
	if err != nil {
		panic(err)
	}
	if f == nil {
		return util{}.notFound(enc, fmt.Sprintf("the frontend with name %s does not exist", name))
	}

	e := loadFrontendFromRequest(r, enc, f)
	if e != nil {
		return util{}.badRequest(enc, "the frontend data is invalid")
	}

	err = svc.SaveFrontend(f)
	if err != nil {
		switch err.Type {
		case ErrBadData:
			return util{}.badRequest(enc, err.Error())
		default:
			panic(err)
		}
	}

	return http.StatusOK, enc.Encode(f)
}

// DeleteFrontend removes an HAProxy frontend.
func DeleteFrontend(enc Encoder, svc DataSvc, params martini.Params) (int, string) {
	key := params["name"]
	err := svc.DeleteFrontend(key)
	if err != nil {
		switch err.Type {
		case ErrNotFound:
			return util{}.notFound(enc, fmt.Sprintf("the frontend with name %s does not exist", key))
		default:
			panic(err)
		}
	}
	return http.StatusNoContent, ""
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
