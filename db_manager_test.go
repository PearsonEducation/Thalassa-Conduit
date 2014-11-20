package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var dbManHelper = TestHelpers{}

func closeDB(dbMgr DBManager) {
	if dbMgr != nil {
		dbMgr.Close()
	}
}

// ----------------------------------------------
// NewDBManager TESTS
// ----------------------------------------------

func Test_NewDBManager(t *testing.T) {
	// create db file for testing
	dbPath := testHelpers.DBPath(t)
	defer os.Remove(dbPath)
	conf := &Config{DBPath: dbPath}

	// create DBManager
	dbMgr, err := NewDBManager(conf)
	defer closeDB(dbMgr)
	assert.EnsureNil(t, err, "NewDBManager() returned an unexpected error: %#v", err)

	// validate that expected DBManager was returned
	assert.Equal(t, reflect.TypeOf(dbMgr), reflect.TypeOf(&levelDBManager{}), "NewDBManager() returned an unexpected object type")
}

// ----------------------------------------------
// levelDBManager.NewDatastore TESTS
// ----------------------------------------------

func Test_levelDBManager_NewDatastore(t *testing.T) {
	// create db file for testing
	dbPath := testHelpers.DBPath(t)
	defer os.Remove(dbPath)
	conf := &Config{DBPath: dbPath}

	// create DBManager
	dbMgr, err := NewDBManager(conf)
	defer closeDB(dbMgr)
	assert.EnsureNil(t, err, "NewDBManager() returned an unexpected error: %#v", err)

	// create BackendDB
	db := dbMgr.NewDatastore()
	assert.EnsureNotNil(t, db, "levelDBManager.NewDatastore() returned a nil value")

	// validate that expected BackendDB was returned
	assert.Equal(t, reflect.TypeOf(db), reflect.TypeOf(&levelDBDatastore{}), "levelDBManager.NewDatastore() returned an unexpected object type")
}

// ----------------------------------------------
// openLevelDBFromFile TESTS
// ----------------------------------------------

func Test_openLevelDBFromFile(t *testing.T) {
	// create db file for testing
	dbPath := testHelpers.DBPath(t)
	defer os.Remove(dbPath)
	openDB, recoverDB := getMockRecoverableOpenerAndRecoverer()

	// validate the function
	_, err := openLevelDBFromFile(openDB, recoverDB, dbPath, nil)
	assert.EnsureNil(t, err, "openLevelDBFromFile() returned an unexpected error: %v", err)
}

func Test_openLevelDBFromFile_Concurrently(t *testing.T) {

	t.Skip("Skipping test. This test will deadlock.")

	type Entry struct {
		ID   int    `json:"id"`
		Data string `json:"data"`
	}

	max := 3
	ch := make(chan Entry, max)

	name, err := ioutil.TempDir("", "conduit_test_db")
	defer os.Remove(name)

	if err != nil {
		t.Fatalf("Unexpected error creating the temp database: %s", err.Error())
	}

	tester := func(dbPath string, index int) {

		log.Printf("[GoRoutine %d] Opening database at: %v\n", index, dbPath)

		db, err := openLevelDBFromFile(leveldb.OpenFile, leveldb.RecoverFile, name, nil)

		if err != nil {
			t.Fatalf("Unexpected error opening the database file: %v", err.Error())
		}

		log.Printf("[GoRoutine %d] Opened database at: %v\n", index, dbPath)

		data := strconv.FormatInt(rand.Int63(), 10)

		log.Printf("[GoRoutine %d] Setting data field to: %v\n", index, data)

		dbEntry, err := json.Marshal(Entry{ID: index, Data: data})

		log.Printf("[GoRoutine %d] Generated data: %s\n", index, string(dbEntry))

		if err != nil {
			t.Fatalf("Unexpected error converting entry to a []byte: %v", err.Error)
		}

		err = db.Put([]byte(string(index)), dbEntry, nil)
		if err != nil {
			t.Fatalf("Unexpected error writing to the database: %v", err.Error())
		}

		waitSeconds := rand.Intn(10)

		log.Printf("[GoRoutine %d] Sleeping for %d seconds...\n", index, waitSeconds)

		select {
		case <-time.After(time.Duration(waitSeconds) * time.Second):
			found, err := db.Get([]byte(string(index)), nil)
			if err != nil {
				t.Fatalf("Unexpected error reading from the database: %v", err.Error())
			}

			log.Printf("[GoRoutine %d] Retrieved data: %v\n", index, string(found))

			result := &Entry{}
			err = json.Unmarshal(found, result)
			if err != nil {
				t.Fatalf("Unexpected error converting retrieved result to an &entry{}: %v", err.Error())
			}

			db.Close()

			ch <- *result
		}
	}

	for i := 0; i < max; i++ {
		go tester(name, i)
	}

	count := 0
	for count < max {
		r := <-ch
		log.Printf("[GoRoutine %d] Exited with data: %v\n", r.ID, r.Data)
		count++
	}
}

func Test_openLevelDBFromFile_AsSingleton(t *testing.T) {

	if testing.Short() {
		t.Skip("This test always works, but it can take up to five seconds to finish. Skipping for speed.")
	}

	type Entry struct {
		ID   int    `json:"id"`
		Data string `json:"data"`
	}

	// Vary the random data a little
	rand.Seed(time.Now().UnixNano())

	ch := make(chan Entry)

	// Create a temporary location for the leveldb data files
	name, err := ioutil.TempDir("", "conduit_test_db")
	defer os.Remove(name)

	if err != nil {
		t.Fatalf("Unexpected error creating the temp database: %s", err.Error())
	}

	// Open the database; The application should keep the DB open until the app shuts down
	log.Printf("Opening database at: %v\n", name)
	db, err := openLevelDBFromFile(leveldb.OpenFile, leveldb.RecoverFile, name, nil)
	defer db.Close()

	if err != nil {
		t.Fatalf("Unexpected error opening the database file: %v", err.Error())
	}

	log.Printf("Opened database at: %v\n", name)

	// A helper function to interact with the data
	tester := func(db leveldb.DB, index, waitSeconds int) {

		data := strconv.FormatInt(rand.Int63(), 10)

		log.Printf("[GoRoutine %d] Setting data field to: %v\n", index, data)

		dbEntry, err := json.Marshal(Entry{ID: index, Data: data})

		log.Printf("[GoRoutine %d] Generated data: %s\n", index, string(dbEntry))

		if err != nil {
			t.Fatalf("Unexpected error converting entry to a []byte: %v", err.Error)
		}

		err = db.Put([]byte(string(index)), dbEntry, nil)
		if err != nil {
			t.Fatalf("Unexpected error writing to the database: %v", err.Error())
		}

		log.Printf("[GoRoutine %d] Sleeping for %d seconds...\n", index, waitSeconds)

		select {
		case <-time.After(time.Duration(waitSeconds) * time.Second):
			found, err := db.Get([]byte(string(index)), nil)
			if err != nil {
				t.Fatalf("Unexpected error reading from the database: %v", err.Error())
			}

			log.Printf("[GoRoutine %d] Retrieved data: %v\n", index, string(found))

			result := &Entry{}
			err = json.Unmarshal(found, result)
			if err != nil {
				t.Fatalf("Unexpected error converting retrieved result to an &entry{}: %v", err.Error())
			}

			ch <- *result
		}
	}

	// Open separate goroutines
	max := 3
	for i := 0; i < max; i++ {
		go tester(*db, i, rand.Intn(6))
	}

	// Process results
	count := 0
	for count < max {
		r := <-ch
		log.Printf("[GoRoutine %d] Exited with data: %v\n", r.ID, r.Data)
		count++
	}
}

func getMockOpenerAndRecoverer(allowRecovery bool) (dbOpener, dbRecoverer) {
	openDB := func(dbPath string, o *opt.Options) (*leveldb.DB, error) {
		corrupted := leveldb.ErrCorrupted{Type: leveldb.CorruptedManifest, Err: errors.New("leveldb: manifest file missing")}
		return nil, corrupted
	}

	recoverDB := func(dbPath string, o *opt.Options) (*leveldb.DB, error) {
		if allowRecovery {
			return leveldb.OpenFile(dbPath, o)
		}

		corrupted := leveldb.ErrCorrupted{Type: leveldb.CorruptedManifest, Err: errors.New("leveldb: manifest file missing")}
		return nil, corrupted
	}
	return openDB, recoverDB
}

func getMockCorruptOpenerAndRecoverer() (dbOpener, dbRecoverer) {
	return getMockOpenerAndRecoverer(false)
}

func getMockRecoverableOpenerAndRecoverer() (dbOpener, dbRecoverer) {
	return getMockOpenerAndRecoverer(true)
}
