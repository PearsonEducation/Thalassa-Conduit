package main

import (
	"fmt"
	"time"
)

// Frontend is the HAProxy frontend data structure, serializable to JSON.
type Frontend struct {
	ID             string            `json:"-"`
	ProxyType      string            `json:"-"`
	Name           string            `json:"name"`
	Bind           string            `json:"bind"`
	DefaultBackend string            `json:"defaultBackend"` // TODO: validation of defaults?
	Mode           string            `json:"mode"`           // http
	KeepAlive      string            `json:"keepalive"`      // default|close|server-close
	Option         string            `json:"option"`         // httplog
	Rules          []string          `json:"rules"`
	Meta           map[string]string `json:"meta"`
}

// String returns the string representation of a frontend.
func (f *Frontend) String() string {
	return f.Name
}

// ToHAProxyFrontend will convert this instance to an haproxy-client.Frontend object.
func (f *Frontend) ToHAProxyFrontend() *Frontend {
	return &Frontend{
		Name:           f.Name,
		Bind:           f.Bind,
		DefaultBackend: f.DefaultBackend,
		Mode:           f.Mode,
		KeepAlive:      f.KeepAlive,
		Option:         f.Option,
		Rules:          f.Rules,
	}
}

// Frontends represents an array of Frontend instances.
type Frontends []*Frontend

// ToInterfaces converts a Frontends instance to an array of empty interfaces.
func (f Frontends) ToInterfaces() []interface{} {
	if len(f) == 0 {
		return nil
	}
	ifs := make([]interface{}, len(f))
	for i, v := range f {
		ifs[i] = v
	}
	return ifs
}

// ToHAProxyFrontends will convert this instance to an haproxy-client.Frontends object.
func (f Frontends) ToHAProxyFrontends() Frontends {
	x := []*Frontend{}
	for _, frontend := range f {
		x = append(x, frontend.ToHAProxyFrontend())
	}
	values := Frontends(x)
	return values
}

// Backend is the HAProxy backend data structure, serializable to JSON.
type Backend struct {
	ID        string            `json:"-"`
	ProxyType string            `json:"-"`
	Name      string            `json:"name"`
	Version   string            `json:"version"` // TODO: proper formatting?
	Balance   string            `json:"balance"`
	Host      string            `json:"host"`
	Mode      string            `json:"mode"`
	Members   BackendMembers    `json:"members"`
	Meta      map[string]string `json:"meta"`
}

// String returns the string representation of a backend.
func (b *Backend) String() string {
	return b.Name
}

// ToHAProxyBackend will convert this instance to an haproxy-client.Backend object.
func (b *Backend) ToHAProxyBackend() *Backend {
	return &Backend{
		Name:    b.Name,
		Balance: b.Balance,
		Host:    b.Host,
		Mode:    b.Mode,
		Members: b.Members.ToHAProxyBackendMembers(),
	}
}

// Backends represents an array of Backend instances.
type Backends []*Backend

// ToInterfaces converts a Backends instance to an array of empty interfaces.
func (b Backends) ToInterfaces() []interface{} {
	if len(b) == 0 {
		return nil
	}
	ifs := make([]interface{}, len(b))
	for i, v := range b {
		ifs[i] = v
	}
	return ifs
}

// ToHAProxyBackends will convert this instance to an haproxy-client.Backends object.
func (b Backends) ToHAProxyBackends() Backends {
	x := []*Backend{}
	for _, backend := range b {
		x = append(x, backend.ToHAProxyBackend())
	}
	values := Backends(x)
	return values
}

// BackendMember is a struct representing the individual member nodes of a backend.
type BackendMember struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Host      string            `json:"host"`
	Port      int               `json:"port"`
	LastKnown time.Time         `json:"lastKnown"`
	Meta      map[string]string `json:"meta"`
}

// String returns the string representation of a backend member.
func (m *BackendMember) String() string {
	return fmt.Sprintf("%s-%s", m.Name, m.Version)
}

// ToHAProxyBackendMember will convert this instance to an haproxy-client.BackendMember object.
func (m *BackendMember) ToHAProxyBackendMember() *BackendMember {
	return &BackendMember{
		Name:      m.Name,
		Host:      m.Host,
		Port:      m.Port,
		LastKnown: m.LastKnown,
	}
}

// BackendMembers represents an array of BackendMember instances.
type BackendMembers []BackendMember

// ToInterfaces converts a BackendMembers instance to an array of empty interfaces.
func (m BackendMembers) ToInterfaces() []interface{} {
	if len(m) == 0 {
		return nil
	}
	ifs := make([]interface{}, len(m))
	for i, v := range m {
		ifs[i] = v
	}
	return ifs
}

// ToHAProxyBackendMembers will convert this instance to an haproxy-client.BackendMembers object.
func (m BackendMembers) ToHAProxyBackendMembers() BackendMembers {
	x := []BackendMember{}
	for _, member := range m {
		x = append(x, *member.ToHAProxyBackendMember())
	}
	values := BackendMembers(x)
	return values
}

// Datastore interface defines methods to manipulate backend data.
type Datastore interface {
	GetAllFrontends() (Frontends, *Error)
	GetFrontend(key string) (*Frontend, *Error)
	SaveFrontend(f *Frontend) *Error
	DeleteFrontend(key string) *Error

	GetAllBackends() (Backends, *Error)
	GetBackend(key string) (*Backend, *Error)
	SaveBackend(b *Backend) *Error
	DeleteBackend(key string) *Error
}
