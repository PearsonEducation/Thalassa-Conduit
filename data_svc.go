package main

import "strings"

// DataSvc represents a service that provided read/write access to HAProxy data.
type DataSvc interface {
	GetAllBackends() (Backends, *Error)
	GetBackend(key string) (*Backend, *Error)
	SaveBackend(f *Backend) *Error
	DeleteBackend(key string) *Error

	GetAllFrontends() (Frontends, *Error)
	GetFrontend(key string) (*Frontend, *Error)
	SaveFrontend(f *Frontend) *Error
	DeleteFrontend(key string) *Error
}

type dataSvcImpl struct {
	db Datastore
	ha HAProxy
}

// NewDataSvc retrieves a new BackendSvc instance.
func NewDataSvc(db Datastore, ha HAProxy) DataSvc {
	return &dataSvcImpl{db: db, ha: ha}
}

// GetAllBackends returns all the backends in the system, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) GetAllBackends() (Backends, *Error) {
	return ds.db.GetAllBackends()
}

// GetBackend returns the backend that has the specified name, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) GetBackend(name string) (*Backend, *Error) {
	return ds.db.GetBackend(name)
}

// SaveBackend persists a backend and returns an error if the operation failed.
// Potential error types:
//   ErrBadData: the backend Name is empty
//   ErrSync: HAProxy config sync failed and delete has been rolled back
//   ErrOutOfSync: HAProxy config and backend data store are out of sync
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) SaveBackend(b *Backend) *Error {
	// validate name
	if b.Name == "" {
		return NewErrorf(ErrBadData, "Name is required")
	}

	// get key value
	b.Name = ds.correctName(b.Name)

	// save record to update in case we need to rollback
	old, derr := ds.db.GetBackend(b.Name)
	if derr != nil {
		return derr
	}

	// execute save
	if derr = ds.db.SaveBackend(b); derr != nil {
		return derr
	}

	// sync HAProxy config
	var rollback func() *Error
	if old != nil {
		rollback = func() *Error { return ds.db.SaveBackend(old) }
	} else {
		rollback = func() *Error { return ds.db.DeleteBackend(b.Name) }
	}
	return ds.syncHAProxy(rollback)
}

// DeleteBackend removes the backend with the specified id; if the backend does not exist, no action is taken.
// Potential error types:
//   ErrNotFound: the backend to delete doesn't exist
//   ErrSync: HAProxy config sync failed and delete has been rolled back
//   ErrOutOfSync: HAProxy config and backend data store are out of sync
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) DeleteBackend(key string) *Error {
	// save record to delete in case we need to rollback
	old, derr := ds.db.GetBackend(key)
	if derr != nil {
		return derr
	}

	// check that backend to delete exists
	if old == nil {
		return NewErrorf(ErrNotFound, "the backend to delete does not exist")
	}

	// execute delete
	if derr = ds.db.DeleteBackend(key); derr != nil {
		return derr
	}

	// sync HAProxy config
	rollback := func() *Error { return ds.db.SaveBackend(old) }
	return ds.syncHAProxy(rollback)
}

// GetAllFrontends returns all the frontends in the system, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) GetAllFrontends() (Frontends, *Error) {
	return ds.db.GetAllFrontends()
}

// GetFrontend returns the frontend that has the specified id, or nil.
// Potential error types:
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) GetFrontend(key string) (*Frontend, *Error) {
	return ds.db.GetFrontend(key)
}

// SaveFrontend persists a frontend and returns an error if the operation failed.
// Potential error types:
//   ErrBadData: the frontend Name is empty
//   ErrSync: HAProxy config sync failed and update has been rolled back
//   ErrOutOfSync: HAProxy config and frontend data store are out of sync
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) SaveFrontend(f *Frontend) *Error {
	// validate name
	if f.Name == "" {
		return NewErrorf(ErrBadData, "Name is required")
	}

	// get key value
	f.Name = ds.correctName(f.Name)

	// save record to update in case we need to rollback
	old, derr := ds.db.GetFrontend(f.Name)
	if derr != nil {
		return derr
	}

	// execute save
	if derr = ds.db.SaveFrontend(f); derr != nil {
		return derr
	}

	// sync HAProxy config
	var rollback func() *Error
	if old != nil {
		rollback = func() *Error { return ds.db.SaveFrontend(old) }
	} else {
		rollback = func() *Error { return ds.db.DeleteFrontend(f.Name) }
	}
	return ds.syncHAProxy(rollback)
}

// DeleteFrontend removes the frontend with the specified id; if the frontend does not exist, no action is taken.
// Potential error types:
//   ErrNotFound: the frontend to delete doesn't exist
//   ErrSync: HAProxy config sync failed and delete has been rolled back
//   ErrOutOfSync: HAProxy config and frontend data store are out of sync
//   ErrDB: error reading/writing to the database
func (ds *dataSvcImpl) DeleteFrontend(key string) *Error {
	// save record to delete in case we need to rollback
	old, derr := ds.db.GetFrontend(key)
	if derr != nil {
		return derr
	}

	// check that frontend to delete exists
	if old == nil {
		return NewErrorf(ErrNotFound, "the frontend to delete does not exist")
	}

	// execute delete
	if derr = ds.db.DeleteFrontend(key); derr != nil {
		return derr
	}

	// sync HAProxy config
	rollback := func() *Error { return ds.db.SaveFrontend(old) }
	return ds.syncHAProxy(rollback)
}

// syncs the HAProxy config file with the backend data in the data store
func (ds *dataSvcImpl) syncHAProxy(rollback func() *Error) *Error {
	// function that syncs HAProxy config file
	sync := func() error {
		b, derr := ds.db.GetAllBackends()
		if derr != nil {
			return derr
		}
		f, err := ds.db.GetAllFrontends()
		if err != nil {
			return err
		}
		if err := ds.ha.WriteConfig(f.ToHAProxyFrontends(), b.ToHAProxyBackends()); err != nil {
			return err
		}
		return nil
	}

	// execute sync function - if sync fails, execute passed in rollback function
	if err := sync(); err != nil {
		if derr := rollback(); derr != nil {
			return NewError(ErrOutOfSync, derr)
		}
		return NewError(ErrSync, err)
	}

	// instruct HAProxy to reload it's config file
	if err := ds.ha.ReloadConfig(); err != nil {
		return NewError(ErrSync, err)
	}
	return nil
}

// formats and returns the key for the given backend
func (ds *dataSvcImpl) correctName(name string) string {
	// remove spaces in name and replace with underscores
	return strings.Replace(name, " ", "_", -1)
}
