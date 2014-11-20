package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

type backendHandlersTestCase struct {
	Setup    func(m *backendHandlersMocks)
	Action   func(m *backendHandlersMocks)
	Teardown func(m *backendHandlersMocks)
}

func (c backendHandlersTestCase) execute() {
	mocks := &backendHandlersMocks{
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

type backendHandlersMocks struct {
	Enc       Encoder
	Svc       *DataSvcMock
	Params    martini.Params
	Request   *http.Request
	ResWriter http.ResponseWriter
}

var bData = BackendTestData{}

// ----------------------------------------------
// GetBackends TESTS
// ----------------------------------------------

func Test_GetBackends(t *testing.T) {
	b1 := bData.OneBackend()
	b2 := bData.OtherBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b1)
		m.Svc.SaveBackend(b2)
	}

	testAction := func(m *backendHandlersMocks) {
		// retrieve and validate data
		actBody := GetBackends(m.Enc, m.Svc)
		expBody := m.Enc.EncodeMulti(b1, b2)
		assert.Equal(t, actBody, expBody, "GetBackends() returned an unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetBackends_SvcError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.GetAllError = NewErrorf(ErrUnknown, "")
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { GetBackends(m.Enc, m.Svc) }
		assert.Panic(t, b, "GetBackends() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// GetBackend TESTS
// ----------------------------------------------

func Test_GetBackend(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Params["name"] = b.Name
	}

	testAction := func(m *backendHandlersMocks) {
		// retrieve and validate data
		actCode, actBody := GetBackend(m.Enc, m.Svc, m.Params)
		assert.Equal(t, actCode, http.StatusOK, "GetBackend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(b), "GetBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetBackend_DoesNotExist(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// retrieve and validate data
		actCode, actBody := GetBackend(m.Enc, m.Svc, m.Params)
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "GetBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "GetBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetBackend_SvcError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.GetError = NewErrorf(ErrUnknown, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { GetBackend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "GetBackends() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// PutBackend TESTS
// ----------------------------------------------

func Test_PutBackend_Create(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutBackend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expBody := m.Enc.Encode(b)
		assert.Equal(t, actCode, http.StatusCreated, "PutBackend() returned unexpected status code")
		assert.Equal(t, actBody, expBody, "PutBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_Update(t *testing.T) {
	b1 := bData.OneBackend()
	b2 := *b1
	b2.Mode = "new mode"

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b1)
		m.Params["name"] = b2.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b2)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutBackend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		assert.Equal(t, actCode, http.StatusOK, "PutBackend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(b2), "PutBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_WithInvalidJSON(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Params["name"] = "12345"
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(`{"test:true}`))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutBackend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PutBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PutBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_SvcBadDataError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrBadData, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PutBackend(m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PutBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PutBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_SvcDBError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrDB, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PutBackend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PutBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_SvcSyncError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrSync, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PutBackend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PutBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PutBackend_SvcOutOfSyncError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("PUT", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PutBackend(m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PutBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// PostBackend TESTS
// ----------------------------------------------

func Test_PostBackend(t *testing.T) {
	b := bData.OneBackend()
	b2 := *b
	b2.Mode = "new mode"

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Params["name"] = b2.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(`{"mode":"new mode"}`))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		assert.Equal(t, actCode, http.StatusOK, "PostBackend() returned unexpected status code")
		assert.Equal(t, actBody, m.Enc.Encode(b2), "PostBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_WithInvalidJSON(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(`{"test:true}`))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_SvcNotFoundError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_SvcBadDataError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Svc.SaveError = NewErrorf(ErrBadData, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusBadRequest
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "PostBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "PostBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_SvcDBError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Svc.SaveError = NewErrorf(ErrDB, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PostBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_SvcSyncError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Svc.SaveError = NewErrorf(ErrSync, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PostBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_PostBackend_SvcOutOfSyncError(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Svc.SaveError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = b.Name
		m.Request, _ = http.NewRequest("POST", "/backends", strings.NewReader(m.Enc.Encode(b)))
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { PostBackend(m.ResWriter, m.Request, m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "PostBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// DeleteBackend TESTS
// ----------------------------------------------

func Test_DeleteBackend(t *testing.T) {
	b := bData.OneBackend()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Params["name"] = b.Name
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := DeleteBackend(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNoContent
		assert.Equal(t, actCode, expCode, "DeleteBackend() returned unexpected status code")
		assert.Empty(t, actBody, "DeleteBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteBackend_SvcNotFoundError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrNotFound, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := DeleteBackend(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "DeleteBackend() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "DeleteBackend() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteBackend_SvcDBError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrDB, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { DeleteBackend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "DeleteBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteBackend_SvcSyncError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrSync, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { DeleteBackend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "DeleteBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_DeleteBackend_SvcOutOfSyncError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.DeleteError = NewErrorf(ErrOutOfSync, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { DeleteBackend(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "DeleteBackend() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// GetBackendMembers TESTS
// ----------------------------------------------

func Test_GetBackendMembers(t *testing.T) {
	b := bData.OneBackendMultiMembers()

	setup := func(m *backendHandlersMocks) {
		m.Svc.SaveBackend(b)
		m.Params["name"] = b.Name
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := GetBackendMembers(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusOK
		expBody := m.Enc.Encode(b.Members)
		assert.Equal(t, expCode, actCode, "GetBackendMembers() returned unexpected status code")
		assert.Equal(t, expBody, actBody, "GetBackendMembers() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetBackendMembers_DoesNotExist(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test
		actCode, actBody := GetBackendMembers(m.Enc, m.Svc, m.Params)

		// assert return values
		expCode := http.StatusNotFound
		expBody := fmt.Sprintf(`"code":%d`, expCode)
		assert.Equal(t, actCode, expCode, "GetBackendMembers() returned unexpected status code")
		assert.StringContains(t, actBody, expBody, "GetBackendMembers() returned unexpected body")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

func Test_GetBackendMembers_SvcError(t *testing.T) {
	setup := func(m *backendHandlersMocks) {
		m.Svc.GetError = NewErrorf(ErrUnknown, "")
		m.Params["name"] = "12345"
	}

	testAction := func(m *backendHandlersMocks) {
		// execute function to test, check for panic
		b := func() { GetBackendMembers(m.Enc, m.Svc, m.Params) }
		assert.Panic(t, b, "GetBackendMembers() failed to panic when expected")
	}

	backendHandlersTestCase{
		Setup:    setup,
		Action:   testAction,
		Teardown: nil,
	}.execute()
}

// ----------------------------------------------
// loadBackendFromRequest TESTS
// ----------------------------------------------

func Test_loadBackendFromRequest(t *testing.T) {
	b1 := bData.OneBackend()
	enc := JSONEncoder{}
	r, _ := http.NewRequest("POST", "/backends", strings.NewReader(enc.Encode(b1)))

	// execute function to test
	b := &Backend{}
	err := loadBackendFromRequest(r, enc, b)

	// assert return values
	assert.EnsureNil(t, err, "loadBackendFromRequest() returned an expected error: %v", err)
	assert.Equal(t, b1, b, "loadBackendFromRequest() did not correctly populate the backend")
}

func Test_loadBackendFromRequest_BadData(t *testing.T) {
	enc := JSONEncoder{}
	r, _ := http.NewRequest("POST", "/backends", strings.NewReader(`{"test:true}`))

	// execute function to test
	b := &Backend{}
	err := loadBackendFromRequest(r, enc, b)

	// assert return values
	assert.EnsureNotNil(t, err, "loadBackendFromRequest() failed to return an expected error")
	assert.Equal(t, err.Code, http.StatusBadRequest, "loadBackendFromRequest() returned unexpected status code in error")
	assert.NotEmpty(t, err.Message, "loadBackendFromRequest() returned empty error message")
}
