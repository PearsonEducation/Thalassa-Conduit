package main

import (
	"io/ioutil"
	"testing"
	"text/template"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestHelpers provides a centralized place to access testing helper functions and mocks.
type TestHelpers struct{}

var testHelpers = TestHelpers{}

// NewFrontendTestData returns a struct that provides Frontend test data.
func (TestHelpers) NewFrontendTestData() FrontendTestData {
	return FrontendTestData{}
}

// NewBackendTestData returns a struct that provides Backend test data.
func (TestHelpers) NewBackendTestData() BackendTestData {
	return BackendTestData{}
}

// NewDatastoreMock returns a mock Datastore instance.
func (TestHelpers) NewDatastoreMock() *DatastoreMock {
	return &DatastoreMock{}
}

// NewDataSvcMock returns a mock DataSvc instance.
func (TestHelpers) NewDataSvcMock() *DataSvcMock {
	return &DataSvcMock{}
}

// NewHAProxyMock returns a mock HAProxy instance.
func (TestHelpers) NewHAProxyMock() *HAProxyMock {
	return &HAProxyMock{}
}

// NewEncoderMock returns a mock Encoder instance.
func (TestHelpers) NewEncoderMock() *EncoderMock {
	return &EncoderMock{}
}

// NewDBManagerMock returns a mock DBManager instance.
func (TestHelpers) NewDBManagerMock() *DBManagerMock {
	return &DBManagerMock{}
}

// DBPath retrieves a DBPath for data layer testing.
func (TestHelpers) DBPath(t *testing.T) string {
	dbPath, err := ioutil.TempDir("", "conduit_test_db")
	assert.EnsureNil(t, err, "Error while opening the temp database: %#v", err)
	return dbPath
}

// DBPath retrieves a LevelDB instance for data layer testing.
func (TestHelpers) LevelDB(t *testing.T, dbPath string) *leveldb.DB {
	db, err := openLevelDBFromFile(leveldb.OpenFile, leveldb.RecoverFile, dbPath, nil)
	assert.EnsureNil(t, err, "Error establishing connection to test db: %v", err)
	return db
}

// ----------------------------------------------
// FrontendTestData
// ----------------------------------------------

type FrontendTestData struct{}

func (f FrontendTestData) OneFrontend() *Frontend {
	return &Frontend{
		Name:           "first_frontend",
		Bind:           "*:80",
		DefaultBackend: "first test backend",
		Mode:           "http",
		KeepAlive:      "default",
	}
}

func (f FrontendTestData) OtherFrontend() *Frontend {
	return &Frontend{
		Name:           "second_frontend",
		Bind:           "*:8080",
		DefaultBackend: "second test backend",
		Mode:           "http",
		KeepAlive:      "default",
	}
}

// ----------------------------------------------
// BackendTestData
// ----------------------------------------------

type BackendTestData struct{}

func (b BackendTestData) OneBackend() *Backend {
	return &Backend{
		Name:    "test001-1.2.5",
		Version: "1.2.5",
		Balance: "roundrobin",
		Host:    "10.180.1.0",
		Mode:    "http",
		Members: BackendMembers{
			BackendMember{
				Name:      "backend/test001/10.180.1.1",
				Version:   "1.2.5",
				Host:      "10.180.1.1",
				Port:      8080,
				LastKnown: time.Now(),
			},
		},
	}
}

func (b BackendTestData) OneBackendMultiMembers() *Backend {
	return &Backend{
		Name:    "test002-1.2.5",
		Version: "1.2.5",
		Balance: "roundrobin",
		Host:    "10.180.2.0",
		Mode:    "http",
		Members: BackendMembers{
			BackendMember{
				Name:      "backend/test002/10.180.2.1",
				Version:   "1.2.5",
				Host:      "10.180.2.1",
				Port:      8080,
				LastKnown: time.Now(),
			},
			BackendMember{
				Name:      "backend/test002/10.180.2.2",
				Version:   "1.2.5",
				Host:      "10.180.2.2",
				Port:      8080,
				LastKnown: time.Now(),
			},
		},
	}
}

func (b BackendTestData) OtherBackend() *Backend {
	return &Backend{
		Name:    "test003-1.2.5",
		Version: "1.2.5",
		Balance: "roundrobin",
		Host:    "10.180.3.0",
		Mode:    "http",
		Members: BackendMembers{
			BackendMember{
				Name:      "backend/test003/10.180.3.1",
				Version:   "1.2.5",
				Host:      "10.180.3.1",
				Port:      8080,
				LastKnown: time.Now(),
			},
		},
	}
}

func (b BackendTestData) OtherBackendMultiMembers() *Backend {
	return &Backend{
		Name:    "test004-1.2.5",
		Version: "1.2.5",
		Balance: "roundrobin",
		Host:    "10.180.4.0",
		Mode:    "http",
		Members: BackendMembers{
			BackendMember{
				Name:      "backend/test004/10.180.4.1",
				Version:   "1.2.5",
				Host:      "10.180.4.1",
				Port:      8080,
				LastKnown: time.Now(),
			},
			BackendMember{
				Name:      "backend/test004/10.180.4.2",
				Version:   "1.2.5",
				Host:      "10.180.4.2",
				Port:      8080,
				LastKnown: time.Now(),
			},
		},
	}
}

// ----------------------------------------------
// DatastoreMock
// ----------------------------------------------

type DatastoreMock struct {
	Backends  Backends
	Frontends Frontends
}

func (db *DatastoreMock) GetAllBackends() (Backends, *Error) {
	b := make(Backends, len(db.Backends))
	for i, x := range db.Backends {
		val := *x
		b[i] = &val
	}
	return b, nil
}
func (db *DatastoreMock) GetBackend(key string) (*Backend, *Error) {
	var val *Backend
	for _, x := range db.Backends {
		if x.Name == key {
			val = x
		}
	}
	if val == nil {
		return nil, nil
	}
	b := *val
	return &b, nil
}
func (db *DatastoreMock) SaveBackend(b *Backend) *Error {
	for _, x := range db.Backends {
		if x.Name == b.Name {
			*x = *b
			return nil
		}
	}
	val := &Backend{}
	*val = *b
	db.Backends = append(db.Backends, val)
	return nil
}
func (db *DatastoreMock) DeleteBackend(key string) *Error {
	index := -1
	for i, x := range db.Backends {
		if x.Name == key {
			index = i
			break
		}
	}
	if index >= 0 {
		db.Backends = append(db.Backends[:index], db.Backends[index+1:]...)
	}
	return nil
}

func (db *DatastoreMock) GetAllFrontends() (Frontends, *Error) {
	f := make(Frontends, len(db.Frontends))
	for i, x := range db.Frontends {
		val := *x
		f[i] = &val
	}
	return f, nil
}
func (db *DatastoreMock) GetFrontend(key string) (*Frontend, *Error) {
	var val *Frontend
	for _, x := range db.Frontends {
		if x.Name == key {
			val = x
			break
		}
	}
	if val == nil {
		return nil, nil
	}
	f := *val
	return &f, nil
}
func (db *DatastoreMock) SaveFrontend(f *Frontend) *Error {
	for _, x := range db.Frontends {
		if x.Name == f.Name {
			*x = *f
			return nil
		}
	}
	val := &Frontend{}
	*val = *f
	db.Frontends = append(db.Frontends, val)
	return nil
}
func (db *DatastoreMock) DeleteFrontend(key string) *Error {
	index := -1
	for i, x := range db.Frontends {
		if x.Name == key {
			index = i
			break
		}
	}
	if index >= 0 {
		db.Frontends = append(db.Frontends[:index], db.Frontends[index+1:]...)
	}
	return nil
}

// ----------------------------------------------
// DataSvcMock
// ----------------------------------------------

type DataSvcMock struct {
	Frontends   Frontends
	Backends    Backends
	GetAllError *Error
	GetError    *Error
	SaveError   *Error
	DeleteError *Error
}

func (svc *DataSvcMock) GetAllFrontends() (Frontends, *Error) {
	if svc.GetAllError != nil {
		return nil, svc.GetAllError
	}
	b := make(Frontends, len(svc.Frontends))
	for i, x := range svc.Frontends {
		val := *x
		b[i] = &val
	}
	return b, nil
}
func (svc *DataSvcMock) GetFrontend(key string) (*Frontend, *Error) {
	if svc.GetError != nil {
		return nil, svc.GetError
	}
	var val *Frontend
	for _, x := range svc.Frontends {
		if x.Name == key {
			val = x
		}
	}
	if val == nil {
		return nil, nil
	}
	b := *val
	return &b, nil
}
func (svc *DataSvcMock) SaveFrontend(b *Frontend) *Error {
	if svc.SaveError != nil {
		return svc.SaveError
	}
	for _, x := range svc.Frontends {
		if x.Name == b.Name {
			*x = *b
			return nil
		}
	}
	val := &Frontend{}
	*val = *b
	svc.Frontends = append(svc.Frontends, val)
	return nil
}
func (svc *DataSvcMock) DeleteFrontend(key string) *Error {
	if svc.DeleteError != nil {
		return svc.DeleteError
	}
	index := -1
	for i, x := range svc.Frontends {
		if x.Name == key {
			index = i
			break
		}
	}
	if index >= 0 {
		svc.Frontends = append(svc.Frontends[:index], svc.Frontends[index+1:]...)
	}
	return nil
}

func (svc *DataSvcMock) GetAllBackends() (Backends, *Error) {
	if svc.GetAllError != nil {
		return nil, svc.GetAllError
	}
	b := make(Backends, len(svc.Backends))
	for i, x := range svc.Backends {
		val := *x
		b[i] = &val
	}
	return b, nil
}
func (svc *DataSvcMock) GetBackend(key string) (*Backend, *Error) {
	if svc.GetError != nil {
		return nil, svc.GetError
	}
	var val *Backend
	for _, x := range svc.Backends {
		if x.Name == key {
			val = x
		}
	}
	if val == nil {
		return nil, nil
	}
	b := *val
	return &b, nil
}
func (svc *DataSvcMock) SaveBackend(b *Backend) *Error {
	if svc.SaveError != nil {
		return svc.SaveError
	}
	for _, x := range svc.Backends {
		if x.Name == b.Name {
			*x = *b
			return nil
		}
	}
	val := &Backend{}
	*val = *b
	svc.Backends = append(svc.Backends, val)
	return nil
}
func (svc *DataSvcMock) DeleteBackend(key string) *Error {
	if svc.DeleteError != nil {
		return svc.DeleteError
	}
	index := -1
	for i, x := range svc.Backends {
		if x.Name == key {
			index = i
			break
		}
	}
	if index >= 0 {
		svc.Backends = append(svc.Backends[:index], svc.Backends[index+1:]...)
	}
	return nil
}

// ----------------------------------------------
// HAProxyMock
// ----------------------------------------------

type HAProxyMock struct {
	config             string
	template           *template.Template
	getConfigAction    func() (string, error)
	getFrontendsAction func() (Frontends, error)
	getBackendsAction  func() (Backends, error)
	writeConfigAction  func(frontends Frontends, backends Backends) error
	reloadConfigAction func() error
}

func (h *HAProxyMock) Template() *template.Template {
	return h.template
}

func (h *HAProxyMock) SetTemplate(t *template.Template) {
	h.template = t
}

func (h *HAProxyMock) GetConfig() (string, error) {
	if h.getConfigAction != nil {
		return h.getConfigAction()
	}
	return h.config, nil
}

func (h *HAProxyMock) GetFrontends() (Frontends, error) {
	if h.getFrontendsAction != nil {
		return h.getFrontendsAction()
	}
	return Frontends{}, nil
}

func (h *HAProxyMock) GetBackends() (Backends, error) {
	if h.getBackendsAction != nil {
		return h.getBackendsAction()
	}
	return Backends{}, nil
}

func (h *HAProxyMock) WriteConfig(frontends Frontends, backends Backends) error {
	if h.writeConfigAction != nil {
		return h.writeConfigAction(frontends, backends)
	}
	return nil
}

func (h *HAProxyMock) ReloadConfig() error {
	if h.reloadConfigAction != nil {
		return h.reloadConfigAction()
	}
	return nil
}

// ----------------------------------------------
// EncoderMock
// ----------------------------------------------

type EncoderMock struct {
	DecodeError error
}

func (e *EncoderMock) Encode(v interface{}) string {
	err := v.(*Error)
	return err.Error()
}

func (e *EncoderMock) EncodeColl(v ...interface{}) string {
	return ""
}

func (e *EncoderMock) Decode(b []byte, i interface{}) error {
	if e.DecodeError != nil {
		return e.DecodeError
	}
	return nil
}

// ----------------------------------------------
// DBManagerMock
// ----------------------------------------------

type DBManagerMock struct{}

func (m *DBManagerMock) Close() error {
	return nil
}
func (m *DBManagerMock) NewDatastore() Datastore {
	return nil
}
