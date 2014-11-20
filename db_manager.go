package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// DBManager interface defines methods for working with a database.
type DBManager interface {
	Close() error
	NewDatastore() Datastore
}

type dbOpener func(dbPath string, o *opt.Options) (*leveldb.DB, error)
type dbRecoverer func(dbPath string, o *opt.Options) (*leveldb.DB, error)

type levelDBManager struct {
	db *leveldb.DB
}

// NewDBManager will return a new DBManager instance.
func NewDBManager(config *Config) (DBManager, error) {
	db, err := openLevelDBFromFile(leveldb.OpenFile, leveldb.RecoverFile, config.DBPath, nil)
	if err != nil {
		return nil, err
	}
	return &levelDBManager{db: db}, nil
}

// Close will close the database.
func (m *levelDBManager) Close() error {
	return m.db.Close()
}

// NewDatastore will return a new Datastore instance.
func (m *levelDBManager) NewDatastore() Datastore {
	return &levelDBDatastore{db: m.db}
}

// OpenDBFromFile will attempt to open (or create) a connection to the database
// specified by dbPath using options o. If it detects that the database files
// are corrupt, this method will attempt to automatically recover them.
func openLevelDBFromFile(opener dbOpener, recoverer dbRecoverer, dbPath string, o *opt.Options) (*leveldb.DB, error) {

	db, err := opener(dbPath, o)

	if err != nil {
		if _, ok := err.(leveldb.ErrCorrupted); ok {
			db, err = attemptRecovery(recoverer, dbPath, o)
		}
	}

	return db, err
}

// AttemptRecovery attempts to recover from a corrupt database specified by
// dbPath using options o.
func attemptRecovery(recoverer dbRecoverer, dbPath string, o *opt.Options) (*leveldb.DB, error) {
	return recoverer(dbPath, o)
}
