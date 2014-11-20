package main

import (
	"encoding/json"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	ldbutil "github.com/syndtr/goleveldb/leveldb/util"
)

type levelDBDatastore struct {
	db *leveldb.DB
}

// GetAllFrontends returns all the frontends in the database, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) GetAllFrontends() (Frontends, *Error) {
	db := ldb.db
	results := Frontends{}

	iter := db.NewIterator(ldbutil.BytesPrefix([]byte("frontend")), nil)
	for iter.Next() {
		frontend := &Frontend{}
		if err := json.Unmarshal(iter.Value(), frontend); err != nil {
			return nil, NewError(ErrDB, err)
		}

		results = append(results, frontend)
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		return nil, NewError(ErrDB, err)
	}

	return results, nil
}

// GetFrontend returns the frontend that has the specified id, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) GetFrontend(key string) (*Frontend, *Error) {
	db := ldb.db
	result := &Frontend{}
	id := fmt.Sprintf("frontend/%s", key)
	resultBytes, err := db.Get([]byte(id), nil)

	if err != nil {
		// If an entity isn't found in the database, its reported as an error
		// from LevelDB, but this isn't an error to Conduit
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, NewError(ErrDB, err)
	}

	if err = json.Unmarshal(resultBytes, result); err != nil {
		return nil, NewError(ErrDB, err)
	}

	return result, nil
}

// SaveFrontend upserts a frontend and returns an error if the operation failed.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) SaveFrontend(f *Frontend) *Error {
	db := ldb.db

	f.ProxyType = "frontend"
	f.ID = fmt.Sprintf("%s/%s", f.ProxyType, f.Name)

	aBytes, err := json.Marshal(f)
	if err != nil {
		return NewError(ErrDB, err)
	}

	if err := db.Put([]byte(f.ID), aBytes, nil); err != nil {
		return NewError(ErrDB, err)
	}

	// clear out ID and ProxyType fields - if we don't, then this object won't be considered
	// equal to a Fackend instance returned by the Get function since these fields are
	// not populated in that call
	f.ID = ""
	f.ProxyType = ""
	return nil
}

// DeleteFrontend removes the frontend with the specified id; if the frontend does not exist, no action is taken.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) DeleteFrontend(key string) *Error {
	db := ldb.db
	fullkey := fmt.Sprintf("frontend/%s", key)
	err := db.Delete([]byte(fullkey), nil)

	if err != nil {
		return NewError(ErrDB, err)
	}

	return nil
}

// GetAllBackends returns all the backends in the database, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) GetAllBackends() (Backends, *Error) {
	db := ldb.db
	results := Backends{}

	iter := db.NewIterator(ldbutil.BytesPrefix([]byte("backend")), nil)
	for iter.Next() {
		backend := &Backend{}
		if err := json.Unmarshal(iter.Value(), backend); err != nil {
			return nil, NewError(ErrDB, err)
		}

		results = append(results, backend)
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		return nil, NewError(ErrDB, err)
	}

	return results, nil
}

// GetBackend returns the backend that has the specified id, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) GetBackend(key string) (*Backend, *Error) {
	db := ldb.db
	result := &Backend{}
	id := fmt.Sprintf("backend/%s", key)

	resultBytes, err := db.Get([]byte(id), nil)
	if err != nil {
		// If an entity isn't found in the database, its reported as an error
		// from LevelDB, but this isn't an error to Conduit
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, NewError(ErrDB, err)
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return nil, NewError(ErrDB, err)
	}

	return result, nil
}

// SaveBackend upserts a backend and returns an error if the operation failed.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) SaveBackend(b *Backend) *Error {
	db := ldb.db

	b.ProxyType = "backend"
	b.ID = fmt.Sprintf("%s/%s", b.ProxyType, b.Name)

	aBytes, err := json.Marshal(b)
	if err != nil {
		return NewError(ErrDB, err)
	}

	if err := db.Put([]byte(b.ID), aBytes, nil); err != nil {
		return NewError(ErrDB, err)
	}

	// clear out ID and ProxyType fields - if we don't, then this object won't be considered
	// equal to a haproxy.Backend instance returned by the Get function since these fields are
	// not populated in that call
	b.ID = ""
	b.ProxyType = ""
	return nil
}

// DeleteBackend removes the backend with the specified id; if the backend does not exist, no action is taken.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ldb *levelDBDatastore) DeleteBackend(key string) *Error {
	db := ldb.db
	id := fmt.Sprintf("backend/%s", key)

	if err := db.Delete([]byte(id), nil); err != nil {
		return NewError(ErrDB, err)
	}

	return nil
}
