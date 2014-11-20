package main

import (
	"os"
	"testing"
)

type levelDBTestCase struct {
	Setup    func(db Datastore)
	Action   func(db Datastore)
	Teardown func(db Datastore)
}

func (c levelDBTestCase) execute(t *testing.T) {
	// setup test
	dbPath := testHelpers.DBPath(t)
	defer os.Remove(dbPath)
	leveldb := testHelpers.LevelDB(t, dbPath)
	defer leveldb.Close()

	db := &levelDBDatastore{db: leveldb}

	// perform setup
	if c.Setup != nil {
		c.Setup(db)
	}

	// perform action
	c.Action(db)

	// perform teardown
	if c.Teardown != nil {
		c.Teardown(db)
	}
}

var ldbFTData = FrontendTestData{}
var ldbBTData = BackendTestData{}

// ----------------------------------------------
// levelDBFrontend.GetAll TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBFrontend.GetAll() function.
func Test_levelDBFrontend_GetAll(t *testing.T) {
	f1 := ldbFTData.OneFrontend()
	f2 := ldbFTData.OtherFrontend()

	setup := func(db Datastore) {
		// add frontend
		derr := db.SaveFrontend(f1)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)
		derr = db.SaveFrontend(f2)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// retrieve all frontends
		frontends, derr := db.GetAllFrontends()
		assert.EnsureNil(t, derr, "levelDBFrontend.GetAll() returned an unexpected error: %v", derr)

		// assert returned frontends
		assert.EnsureEqual(t, len(frontends), 2, "levelDBFrontend.GetAll() returned an unexpected number of values")
		assert.Equal(t, frontends[0], f1, "levelDBFrontend.GetAll() returned an unexpected value")
		assert.Equal(t, frontends[1], f2, "levelDBFrontend.GetAll() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBFrontend.Get TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBFrontend.Get() function.
func Test_levelDBFrontend_Get(t *testing.T) {
	f := ldbFTData.OneFrontend()

	setup := func(db Datastore) {
		// add frontend
		derr := db.SaveFrontend(f)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// retrieve frontend
		returnedFrontend, derr := db.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "levelDBFrontend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "levelDBFrontend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBFrontend.Save TESTS
// ----------------------------------------------

// Tests the "happy path" for creating a frontend with the levelDBFrontend.Save() function.
func Test_levelDBFrontend_Save_Create(t *testing.T) {
	f := ldbFTData.OneFrontend()

	testAction := func(db Datastore) {
		// add frontend
		derr := db.SaveFrontend(f)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedFrontend, derr := db.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "levelDBFrontend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "levelDBFrontend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// Tests the "happy path" for the levelDBFrontend.Save() function.
func Test_levelDBFrontend_Save_Update(t *testing.T) {
	f := ldbFTData.OneFrontend()

	setup := func(db Datastore) {
		// add frontend
		derr := db.SaveFrontend(f)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// update frontend
		f.Mode = "default"
		derr := db.SaveFrontend(f)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedFrontend, derr := db.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "levelDBFrontend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "levelDBFrontend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBFrontend.Delete TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBFrontend.Delete() function.
func Test_levelDBFrontend_Delete(t *testing.T) {
	f := ldbFTData.OneFrontend()

	setup := func(db Datastore) {
		// add frontend
		derr := db.SaveFrontend(f)
		assert.EnsureNil(t, derr, "levelDBFrontend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// delete frontend
		derr := db.DeleteFrontend(f.Name)
		assert.EnsureNil(t, derr, "levelDBFrontend.Delete() returned an unexpected error: %v", derr)

		// validate added data
		returnedFrontend, derr := db.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "levelDBFrontend.Get() returned an unexpected error: %v", derr)
		assert.Nil(t, returnedFrontend, "levelDBFrontend.Delete() failed to delete the frontend")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// Tests that the levelDBFrontend.Delete() function does not return an error if the frontend to
// delete does not exist.
func Test_levelDBFrontend_Delete_NonExistentFrontend(t *testing.T) {
	testAction := func(db Datastore) {
		// delete frontend
		derr := db.DeleteFrontend("does-not-exist")
		assert.Nil(t, derr, "levelDBFrontend.Delete() returned an unexpected error: %v", derr)
	}

	testCase := levelDBTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBBackend.GetAll TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBBackend.GetAll() function.
func Test_levelDBBackend_GetAll(t *testing.T) {
	b1 := ldbBTData.OneBackend()
	b2 := ldbBTData.OtherBackend()

	setup := func(db Datastore) {
		// add backend
		derr := db.SaveBackend(b1)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)
		derr = db.SaveBackend(b2)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// retrieve all backends
		backends, derr := db.GetAllBackends()
		assert.EnsureNil(t, derr, "levelDBBackend.GetAll() returned an unexpected error: %v", derr)

		// assert returned backends
		assert.EnsureEqual(t, len(backends), 2, "levelDBBackend.GetAll() returned an unexpected number of values")
		assert.Equal(t, backends[0], b1, "levelDBBackend.GetAll() returned an unexpected value")
		assert.Equal(t, backends[1], b2, "levelDBBackend.GetAll() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBBackend.Get TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBBackend.Get() function.
func Test_levelDBBackend_Get(t *testing.T) {
	b := ldbBTData.OneBackend()

	setup := func(db Datastore) {
		// add backend
		derr := db.SaveBackend(b)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// retrieve backend
		returnedBackend, derr := db.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "levelDBBackend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "levelDBBackend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBBackend.Save TESTS
// ----------------------------------------------

// Tests the "happy path" for creating a backend with the levelDBBackend.Save() function.
func Test_levelDBBackend_Save_Create(t *testing.T) {
	b := ldbBTData.OneBackend()

	testAction := func(db Datastore) {
		// add backend
		derr := db.SaveBackend(b)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedBackend, derr := db.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "levelDBBackend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "levelDBBackend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// Tests the "happy path" for updating a backend with the levelDBBackend.Save() function.
func Test_levelDBBackend_Save_Update(t *testing.T) {
	b := ldbBTData.OneBackend()

	setup := func(db Datastore) {
		// add backend
		derr := db.SaveBackend(b)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// update backend
		b.Mode = "default"
		derr := db.SaveBackend(b)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedBackend, derr := db.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "levelDBBackend.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "levelDBBackend.Get() returned an unexpected value")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// ----------------------------------------------
// levelDBBackend.Delete TESTS
// ----------------------------------------------

// Tests the "happy path" for the levelDBBackend.Delete() function.
func Test_levelDBBackend_Delete(t *testing.T) {
	b := ldbBTData.OneBackend()

	setup := func(db Datastore) {
		// add backend
		derr := db.SaveBackend(b)
		assert.EnsureNil(t, derr, "levelDBBackend.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(db Datastore) {
		// delete backend
		derr := db.DeleteBackend(b.Name)
		assert.EnsureNil(t, derr, "levelDBBackend.Delete() returned an unexpected error: %v", derr)

		// validate added data
		returnedBackend, derr := db.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "levelDBBackend.Get() returned an unexpected error: %v", derr)
		assert.Nil(t, returnedBackend, "levelDBBackend.Delete() failed to delete the backend")
	}

	testCase := levelDBTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}

// Tests that the levelDBBackend.Delete() function does not return an error if the backend to
// delete does not exist.
func Test_levelDBBackend_Delete_NonExistentBackend(t *testing.T) {
	testAction := func(db Datastore) {
		// delete backend
		derr := db.DeleteBackend("does-not-exist")
		assert.Nil(t, derr, "levelDBBackend.Delete() returned an unexpected error: %v", derr)
	}

	testCase := levelDBTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
	}
	testCase.execute(t)
}
