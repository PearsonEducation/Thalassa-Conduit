package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

type frontendHandlersTestCase struct {
	Setup    func(m *frontendHandlersMocks)
	Action   func(m *frontendHandlersMocks)
	Teardown func(m *frontendHandlersMocks)
}

func (c frontendHandlersTestCase) execute() {
	mocks := &frontendHandlersMocks{
		Enc:       JSONEncoder{},
		Svc:       testHelpers.NewDataSvcMock(),
		Params:    make(map[string]string),
		ResWriter: httptest.NewRecorder(),
	}

	// perform setup
	if c.Setup != nil {
		c.Setup(mocks)
	}

	// perform action
	c.Action(mocks)

	// perform teardown
	if c.Teardown != nil {
		c.Teardown(mocks)
	}
}

type frontendHandlersMocks struct {
	Enc       Encoder
	Svc       *DataSvcMock
	Params    martini.Params
	Request   *http.Request
	ResWriter http.ResponseWriter
}

var fData = FrontendTestData{}

// ----------------------------------------------
// GetFrontends TESTS
// ----------------------------------------------

func Test_GetFrontends(t *testing.T) {
	f1 := fData.OneFrontend()
	f2 := fData.OtherFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f1)
		m.Svc.SaveFrontend(f2)
	}

	testAction := func(m *frontendHandlersMocks) {
		// retrieve and validate data
		actBody := GetFrontends(m.Enc, m.Svc)
		expBody := m.Enc.EncodeMulti(f1, f2)
		assert.Equal(t, actBody, expBody, "GetFrontends() returned an unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetFrontends_SvcError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.GetAllError = NewErrorf(ErrUnknown, "")
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { GetFrontends(m.Enc, m.Svc) }
		assert.Panic(t, f, "GetFrontends() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// GetFrontend TESTS
// ----------------------------------------------

func Test_GetFrontend(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Params["name"] = f.Name
	}

	testAction := func(m *frontendHandlersMocks) {
		// retrieve and validate data
		actCode, actBody := GetFrontend(m.Enc, m.Svc, m.Params)
		assert.Equal(t, actCode, http.StatusOK, "GetFrontend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(f), "GetFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetFrontend_DoesNotExist(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// retrieve and validate data
		actCode, actBody := GetFrontend(m.Enc, m.Svc, m.Params)
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "GetFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "GetFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetFrontend_SvcError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.GetError = NewErrorf(ErrUnknown, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { GetFrontend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "GetFrontends() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// PutFrontend TESTS
// ----------------------------------------------

func Test_PutFrontend_Create(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutFrontend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expBody := m.Enc.Encode(f)
		assert.Equal(t, actCode, http.StatusCreated, "PutFrontend() returned unexpected status code")
		assert.Equal(t, actBody, expBody, "PutFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_Update(t *testing.T) {
	f1 := fData.OneFrontend()
	f2 := *f1
	f2.Mode = "new mode"

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f1)
		m.Params["name"] = f2.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f2)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutFrontend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		assert.Equal(t, actCode, http.StatusOK, "PutFrontend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(f2), "PutFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_WithInvalidJSON(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Params["name"] = "12345"
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(`{"test:true}`))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutFrontend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PutFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PutFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_SvcBadDataError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrBadData, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutFrontend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PutFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PutFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_SvcDBError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrDB, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PutFrontend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PutFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_SvcSyncError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrSync, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PutFrontend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PutFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutFrontend_SvcOutOfSyncError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("PUT", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PutFrontend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PutFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// PostFrontend TESTS
// ----------------------------------------------

func Test_PostFrontend(t *testing.T) {
	f := fData.OneFrontend()
	f2 := *f
	f2.Mode = "new mode"

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Params["name"] = f2.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(`{"mode":"new mode"}`))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		assert.Equal(t, actCode, http.StatusOK, "PostFrontend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(f2), "PostFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_WithInvalidJSON(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(`{"test:true}`))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_SvcNotFoundError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_SvcBadDataError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Svc.SaveError = NewErrorf(ErrBadData, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_SvcDBError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Svc.SaveError = NewErrorf(ErrDB, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PostFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_SvcSyncError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Svc.SaveError = NewErrorf(ErrSync, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PostFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostFrontend_SvcOutOfSyncError(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Svc.SaveError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = f.Name
		m.Request, _ = http.NewRequest("POST", "/frontends", strings.NewReader(m.Enc.Encode(f)))
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { PostFrontend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "PostFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// DeleteFrontend TESTS
// ----------------------------------------------

func Test_DeleteFrontend(t *testing.T) {
	f := fData.OneFrontend()

	setup := func(m *frontendHandlersMocks) {
		m.Svc.SaveFrontend(f)
		m.Params["name"] = f.Name
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := DeleteFrontend(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNoContent
		assert.Equal(t, actCode, expCode, "DeleteFrontend() returned unexpected status code")
		assert.Empty(t, actBody, "DeleteFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteFrontend_SvcNotFoundError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrNotFound, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test
		actCode, actBody := DeleteFrontend(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "DeleteFrontend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "DeleteFrontend() returned unexpected body")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteFrontend_SvcDBError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrDB, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { DeleteFrontend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "DeleteFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteFrontend_SvcSyncError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrSync, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { DeleteFrontend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "DeleteFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteFrontend_SvcOutOfSyncError(t *testing.T) {
	setup := func(m *frontendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *frontendHandlersMocks) {
		// execute function to test, check for panic
		f := func() { DeleteFrontend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, f, "DeleteFrontend() failed to panic when expected")
	}

	frontendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// loadFrontendFromRequest TESTS
// ----------------------------------------------

func Test_loadFrontendFromRequest(t *testing.T) {
	f1 := fData.OneFrontend()
	enc := JSONEncoder{}
	r, _ := http.NewRequest("POST", "/frontends", strings.NewReader(enc.Encode(f1)))

	// execute function to test
	f := &Frontend{}
	err := loadFrontendFromRequest(r, enc, f)

	// assert return values
	assert.EnsureNil(t, err, "loadFrontendFromRequest() returned an expected error: %v", err)
	assert.Equal(t, f1, f, "loadFrontendFromRequest() did not correctly populate the frontend")
}

func Test_loadFrontendFromRequest_BadData(t *testing.T) {
	enc := JSONEncoder{}
	r, _ := http.NewRequest("POST", "/frontends", strings.NewReader(`{"test:true}`))

	// execute function to test
	f := &Frontend{}
	err := loadFrontendFromRequest(r, enc, f)

	// assert return values
	assert.EnsureNotNil(t, err, "loadFrontendFromRequest() failed to return an expected error")
	assert.Equal(t, err.Code, http.StatusBadRequest, "loadFrontendFromRequest() returned unexpected status code in error")
	assert.NotEmpty(t, err.Message, "loadFrontendFromRequest() returned empty error message")
}
