package main

import (
	"errors"
	"fmt"
	"testing"
)

type dataSvcTestCase struct {
	Setup    func(svc DataSvc)
	Action   func(svc DataSvc)
	Teardown func(svc DataSvc)
	Mocks    dataSvcMocks
}

type dataSvcMocks struct {
	DB Datastore
	HA HAProxy
}

func (c dataSvcTestCase) execute() {
	svc := NewDataSvc(c.Mocks.DB, c.Mocks.HA)

	// perform setup
	if c.Setup != nil {
		c.Setup(svc)
	}

	// perform action
	c.Action(svc)

	// perform teardown
	if c.Teardown != nil {
		c.Teardown(svc)
	}
}

func defaultMocks() dataSvcMocks {
	return dataSvcMocks{
		DB: testHelpers.NewDatastoreMock(),
		HA: testHelpers.NewHAProxyMock(),
	}
}

var bsData = BackendTestData{}
var fsData = FrontendTestData{}

// ----------------------------------------------
// frontendSvcImpl.GetAll TESTS
// ----------------------------------------------

// Tests the "happy path" for the frontendSvcImpl.GetAll() function.
func Test_frontendSvcImpl_GetAll(t *testing.T) {
	f1 := fsData.OneFrontend()
	f2 := fsData.OtherFrontend()

	setup := func(svc DataSvc) {
		// add frontend
		derr := svc.SaveFrontend(f1)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
		derr = svc.SaveFrontend(f2)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// retrieve and validate data
		frontends, derr := svc.GetAllFrontends()
		assert.EnsureNil(t, derr, "frontendSvcImpl.GetAll() returned an unexpected error: %v", derr)
		assert.EnsureEqual(t, len(frontends), 2, "frontendSvcImpl.GetAll() returned an unexpected number of values")
		assert.Equal(t, frontends[0], f1, "frontendSvcImpl.GetAll() returned an unexpected value")
		assert.Equal(t, frontends[1], f2, "frontendSvcImpl.GetAll() returned an unexpected value")
	}

	testCase := dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}
	testCase.execute()
}

// ----------------------------------------------
// frontendSvcImpl.Get TESTS
// ----------------------------------------------

// Tests the "happy path" for the frontendSvcImpl.Get() function.
func Test_frontendSvcImpl_Get(t *testing.T) {
	f := fsData.OneFrontend()

	setup := func(svc DataSvc) {
		// add frontend
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// retrieve and validate data
		returnedFrontend, derr := svc.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "frontendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

// ----------------------------------------------
// frontendSvcImpl.Save TESTS
// ----------------------------------------------

// Tests the "happy path" for creating a frontend with the frontendSvcImpl.Save() function.
func Test_frontendSvcImpl_Save_Create(t *testing.T) {
	f := fsData.OneFrontend()

	testAction := func(svc DataSvc) {
		// add frontend
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedFrontend, derr := svc.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "frontendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

// Tests the "happy path" for updating a frontend with the frontendSvcImpl.Save() function.
func Test_frontendSvcImpl_Save_Update(t *testing.T) {
	f := fsData.OneFrontend()

	setup := func(svc DataSvc) {
		// add frontend
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// update frontend
		f.Mode = "default"
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)

		// validate updated data
		returnedFrontend, derr := svc.GetFrontend(f.Name)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedFrontend, f, "frontendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_frontendSvcImpl_Add_NoNameProvided(t *testing.T) {
	f := fsData.OneFrontend()
	f.Name = ""

	testAction := func(svc DataSvc) {
		// attempt to add frontend, validate errors
		derr := svc.SaveFrontend(f)
		assert.EnsureNotNil(t, derr, "frontendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrBadData, fmt.Sprintf("frontendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_frontendSvcImpl_Save_CreateSyncError(t *testing.T) {
	f := fsData.OneFrontend()

	setup := func(svc DataSvc) {
		derr := svc.SaveFrontend(f)
		assert.EnsureNotNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr.Error())
	}
	testAction := func(svc DataSvc) {
		// attempt to add frontend, validate errors
		derr := svc.SaveFrontend(f)
		assert.EnsureNotNil(t, derr, "frontendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("frontendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		return errors.New("test")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks: dataSvcMocks{
			DB: testHelpers.NewDatastoreMock(),
			HA: ha,
		},
	}.execute()
}

func Test_frontendSvcImpl_Save_UpdateSyncError(t *testing.T) {
	f := fsData.OneFrontend()
	original := struct {
		Name string
		Mode string
	}{
		Name: f.Name,
		Mode: f.Mode,
	}

	counter := 0
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		counter++
		if counter >= 2 {
			return errors.New("test")
		}
		return nil
	}
	db := testHelpers.NewDatastoreMock()
	setup := func(svc DataSvc) {
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: '%v'", derr)
	}
	testAction := func(svc DataSvc) {
		// attempt to update frontend, validate errors
		derr := svc.SaveFrontend(f)
		assert.EnsureNotNil(t, derr, "frontendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("frontendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	teardown := func(svc DataSvc) {
		// assert data was rolled back
		f, derr := svc.GetFrontend(original.Name)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Get() returned an unexpected error: '%v'", derr)
		assert.Equal(t, f.Mode, original.Mode, "frontendSbcImpl.Save() failed to rollback the delete action")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: teardown,
		Mocks:    dataSvcMocks{DB: db, HA: ha},
	}.execute()
}

// ----------------------------------------------
// frontendSvcImpl.Delete TESTS
// ----------------------------------------------

// Tests the "happy path" for the frontendSvcImpl.Delete() function.
func Test_frontendSvcImpl_Delete(t *testing.T) {
	f := fsData.OneFrontend()

	setup := func(svc DataSvc) {
		// add frontend
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// delete frontend
		derr := svc.DeleteFrontend(f.Name)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Delete() returned an unexpected error: %v", derr)

		// assert that frontend was deleted
		returnedFrontend, derr := svc.GetFrontend("f.Name")
		assert.Nil(t, derr, "frontendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Nil(t, returnedFrontend, "frontendSvcImpl.Delete() failed to delete the frontend")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_frontendSvcImpl_Delete_NonExistentFrontend(t *testing.T) {
	f := fsData.OneFrontend()

	testAction := func(svc DataSvc) {
		// attempt to update frontend, validate errors
		derr := svc.DeleteFrontend(f.Name)
		assert.EnsureNotNil(t, derr, fmt.Sprintf("frontendSvcImpl.Delete() failed to return an expected error"))
		assert.Equal(t, derr.Type, ErrNotFound, fmt.Sprintf("frontendSvcImpl.Delete() returned an unexpected error type: '%v'", derr.Type.String()))
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_frontendSvcImpl_Delete_SyncError(t *testing.T) {
	f := fsData.OneFrontend()

	setup := func(svc DataSvc) {
		derr := svc.SaveFrontend(f)
		assert.EnsureNil(t, derr, "frontendSvcImpl.Save() returned an unexpected error: %v", derr)
	}
	testAction := func(svc DataSvc) {
		// attempt to update frontend, validate errors
		derr := svc.DeleteFrontend(f.Name)
		assert.EnsureNotNil(t, derr, fmt.Sprintf("frontendSvcImpl.Delete() failed to return an expected error"))
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("frontendSvcImpl.Delete() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	pass := 0
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		pass++
		if pass == 2 {
			return errors.New("test")
		}
		return nil
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks: dataSvcMocks{
			DB: testHelpers.NewDatastoreMock(),
			HA: ha,
		},
	}.execute()
}

// // ----------------------------------------------
// // frontendSvcImpl.getFrontendKey TESTS
// // ----------------------------------------------

// // Tests the "happy path" for the frontendSvcImpl.getFrontendKey() function.
// func Test_frontendSvcImpl_getFrontendKey(t *testing.T) {
// 	// setup test
// 	svc := &dataSvcImpl{}

// 	testCases := []struct{ Name, Result string }{
// 		{
// 			Name:   "name-without_spaces",
// 			Result: "name-without_spaces",
// 		},
// 		{
// 			Name:   "",
// 			Result: "",
// 		},
// 		{
// 			Name:   "name with spaces",
// 			Result: "name_with_spaces",
// 		},
// 	}

// 	for _, testcase := range testCases {
// 		frontend := &Frontend{Name: testcase.Name}
// 		actual := svc.getFrontendKey(frontend)
// 		assert.Equal(t, actual, testcase.Result, fmt.Sprintf("frontendSvcImpl.getFrontendKey() return an unexpected value with input '%v'", testcase.Name))
// 	}
// }

// // This test generates 1000 random UTF-8 strings for the Frontend name field
// // and uses the generated strings to test the getFrontendKey function.
// func Test_frontendSvcImpl_getFrontendKey_QuickCheck(t *testing.T) {
// 	// setup test
// 	svc := &frontendSvcImpl{}
// 	count := 0

// 	f := func(name string) bool {
// 		count++

// 		t.Logf("Test %d: Testing getFrontendKey with name (%v)\n", count, name)

// 		frontend := &Frontend{Name: name}
// 		actual := svc.getFrontendKey(frontend)

// 		return strings.Replace(name, " ", "_", -1) == actual
// 	}
// 	config := &quick.Config{MaxCount: 1000}
// 	if err := quick.Check(f, config); err != nil {
// 		t.Error(err)
// 	}
// }

// ----------------------------------------------
// backendSvcImpl.GetAll TESTS
// ----------------------------------------------

// Tests the "happy path" for the backendSvcImpl.GetAll() function.
func Test_backendSvcImpl_GetAll(t *testing.T) {
	b1 := bsData.OneBackend()
	b2 := bsData.OtherBackend()

	setup := func(svc DataSvc) {
		// add backend
		derr := svc.SaveBackend(b1)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
		derr = svc.SaveBackend(b2)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// retrieve and validate data
		backends, derr := svc.GetAllBackends()
		assert.EnsureNil(t, derr, "backendSvcImpl.GetAll() returned an unexpected error: %v", derr)
		assert.EnsureEqual(t, len(backends), 2, "backendSvcImpl.GetAll() returned an unexpected number of values")
		assert.Equal(t, backends[0], b1, "backendSvcImpl.GetAll() returned an unexpected value")
		assert.Equal(t, backends[1], b2, "backendSvcImpl.GetAll() returned an unexpected value")
	}

	testCase := dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}
	testCase.execute()
}

// ----------------------------------------------
// backendSvcImpl.Get TESTS
// ----------------------------------------------

// Tests the "happy path" for the backendSvcImpl.Get() function.
func Test_backendSvcImpl_Get(t *testing.T) {
	b := bsData.OneBackend()

	setup := func(svc DataSvc) {
		// add backend
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// retrieve and validate data
		returnedBackend, derr := svc.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "backendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "backendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

// ----------------------------------------------
// backendSvcImpl.Save TESTS
// ----------------------------------------------

// Tests the "happy path" for creating a backend with the backendSvcImpl.Save() function.
func Test_backendSvcImpl_Save_Create(t *testing.T) {
	b := bsData.OneBackend()

	testAction := func(svc DataSvc) {
		// add backend
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)

		// validate added data
		returnedBackend, derr := svc.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "backendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "backendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

// Tests the "happy path" for updating a backend with the backendSvcImpl.Save() function.
func Test_backendSvcImpl_Save_Update(t *testing.T) {
	b := bsData.OneBackend()

	setup := func(svc DataSvc) {
		// add backend
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// update backend
		b.Mode = "default"
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)

		// validate updated data
		returnedBackend, derr := svc.GetBackend(b.Name)
		assert.EnsureNil(t, derr, "backendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Equal(t, returnedBackend, b, "backendSvcImpl.Get() returned an unexpected value")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_backendSvcImpl_Save_NoNameProvided(t *testing.T) {
	b := bsData.OneBackend()
	b.Name = ""

	testAction := func(svc DataSvc) {
		// attempt to add backend, validate errors
		derr := svc.SaveBackend(b)
		assert.EnsureNotNil(t, derr, "backendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrBadData, fmt.Sprintf("backendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_backendSvcImpl_Save_CreateSyncError(t *testing.T) {
	b := bsData.OneBackend()

	setup := func(svc DataSvc) {
		derr := svc.SaveBackend(b)
		assert.EnsureNotNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr.Error())
	}
	testAction := func(svc DataSvc) {
		// attempt to add backend, validate errors
		derr := svc.SaveBackend(b)
		assert.EnsureNotNil(t, derr, "backendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("backendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		return errors.New("test")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks: dataSvcMocks{
			DB: testHelpers.NewDatastoreMock(),
			HA: ha,
		},
	}.execute()
}

func Test_backendSvcImpl_Save_UpdateSyncError(t *testing.T) {
	b := bsData.OneBackend()
	original := struct {
		Name string
		Mode string
	}{
		Name: b.Name,
		Mode: b.Mode,
	}

	counter := 0
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		counter++
		if counter >= 2 {
			return errors.New("test")
		}
		return nil
	}
	db := testHelpers.NewDatastoreMock()
	setup := func(svc DataSvc) {
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: '%v'", derr)
	}
	testAction := func(svc DataSvc) {
		// attempt to update backend, validate errors
		derr := svc.SaveBackend(b)
		assert.EnsureNotNil(t, derr, "backendSvcImpl.Save() failed to return an expected error")
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("backendSvcImpl.Save() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	teardown := func(svc DataSvc) {
		// assert data was rolled back
		f, derr := svc.GetBackend(original.Name)
		assert.EnsureNil(t, derr, "backendSvcImpl.Get() returned an unexpected error: '%v'", derr)
		assert.Equal(t, f.Mode, original.Mode, "backendSbcImpl.Save() failed to rollback the delete action")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: teardown,
		Mocks:    dataSvcMocks{DB: db, HA: ha},
	}.execute()
}

// ----------------------------------------------
// backendSvcImpl.Delete TESTS
// ----------------------------------------------

// Tests the "happy path" for the backendSvcImpl.Delete() function.
func Test_backendSvcImpl_Delete(t *testing.T) {
	b := bsData.OneBackend()

	setup := func(svc DataSvc) {
		// add backend
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
	}

	testAction := func(svc DataSvc) {
		// delete backend
		derr := svc.DeleteBackend(b.Name)
		assert.EnsureNil(t, derr, "backendSvcImpl.Delete() returned an unexpected error: %v", derr)

		// assert that backend was deleted
		returnedBackend, derr := svc.GetBackend("test001-1.2.5")
		assert.Nil(t, derr, "backendSvcImpl.Get() returned an unexpected error: %v", derr)
		assert.Nil(t, returnedBackend, "backendSvcImpl.Delete() failed to delete the backend")
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_backendSvcImpl_Delete_NonExistentBackend(t *testing.T) {
	b := bsData.OneBackend()

	testAction := func(svc DataSvc) {
		// attempt to update backend, validate errors
		derr := svc.DeleteBackend(b.Name)
		assert.EnsureNotNil(t, derr, fmt.Sprintf("backendSvcImpl.Delete() failed to return an expected error"))
		assert.Equal(t, derr.Type, ErrNotFound, fmt.Sprintf("backendSvcImpl.Delete() returned an unexpected error type: '%v'", derr.Type.String()))
	}

	dataSvcTestCase{
		Setup:    nil,
		Action:   testAction,
		Teardown: nil,
		Mocks:    defaultMocks(),
	}.execute()
}

func Test_backendSvcImpl_Delete_SyncError(t *testing.T) {
	b := bsData.OneBackend()

	setup := func(svc DataSvc) {
		derr := svc.SaveBackend(b)
		assert.EnsureNil(t, derr, "backendSvcImpl.Save() returned an unexpected error: %v", derr)
	}
	testAction := func(svc DataSvc) {
		// attempt to update backend, validate errors
		derr := svc.DeleteBackend(b.Name)
		assert.EnsureNotNil(t, derr, fmt.Sprintf("backendSvcImpl.Delete() failed to return an expected error"))
		assert.Equal(t, derr.Type, ErrSync, fmt.Sprintf("backendSvcImpl.Delete() returned an unexpected error type: '%v'", derr.Type.String()))
	}
	pass := 0
	ha := testHelpers.NewHAProxyMock()
	ha.writeConfigAction = func(frontends Frontends, backends Backends) error {
		pass++
		if pass == 2 {
			return errors.New("test")
		}
		return nil
	}

	dataSvcTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
		Mocks: dataSvcMocks{
			DB: testHelpers.NewDatastoreMock(),
			HA: ha,
		},
	}.execute()
}

// // ----------------------------------------------
// // backendSvcImpl.getBackendKey TESTS
// // ----------------------------------------------

// func Test_backendSvcImpl_getBackendKey(t *testing.T) {
// 	// setup test
// 	svc := &dataSvcImpl{}

// 	testCases := []struct{ Name, Result string }{
// 		{
// 			Name:   "name-without_spaces",
// 			Result: "name-without_spaces",
// 		},
// 		{
// 			Name:   "",
// 			Result: "",
// 		},
// 		{
// 			Name:   "name with spaces",
// 			Result: "name_with_spaces",
// 		},
// 	}

// 	for _, testcase := range testCases {
// 		backend := &Backend{Name: testcase.Name}
// 		actual := svc.getBackendKey(backend)
// 		assert.Equal(t, actual, testcase.Result, fmt.Sprintf("backendSvcImpl.getBackendKey() return an unexpected value with input '%v'", testcase.Name))
// 	}
// }

// // This test generates 1000 random UTF-8 strings for the Backend name field
// // and uses the generated strings to test the getBackendKey function.
// func Test_backendSvcImpl_getBackendKey_QuickCheck(t *testing.T) {
// 	// setup test
// 	svc := &backendSvcImpl{}
// 	count := 0

// 	f := func(name string) bool {
// 		count++

// 		t.Logf("Test %d: Testing getBackendKey with name (%v)\n", count, name)

// 		backend := &Backend{Name: name}
// 		actual := svc.getBackendKey(backend)

// 		return strings.Replace(name, " ", "_", -1) == actual
// 	}
// 	config := &quick.Config{MaxCount: 1000}
// 	if err := quick.Check(f, config); err != nil {
// 		t.Error(err)
// 	}
// }
