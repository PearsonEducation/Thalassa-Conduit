package main

import (
	"fmt"
	"testing"
)

// ----------------------------------------------
// Backend TESTS
// ----------------------------------------------

// Tests that the Backend.String() function behaves correctly.
func Test_Backend_String(t *testing.T) {
	b := Backend{Name: "Test Backend"}
	assert.Equal(t, b.String(), b.Name, "Backend.String() returned unexpected result")
}

// Tests that the Backends.ToInterfaces() function behaves correctly.
func Test_Backends_ToInterfaces(t *testing.T) {
	backends := Backends{&Backend{Name: "first"}, &Backend{Name: "second"}}
	result := backends.ToInterfaces()
	assert.Equal(t, len(result), len(backends), "Backends.ToInterfaces() returns an array of an unexpected length")
}

// Tests that the Backends.ToInterfaces() function returns the expected value when the array is empty.
func Test_Backends_ToInterfaces_Empty(t *testing.T) {
	backends := make(Backends, 0)
	result := backends.ToInterfaces()
	assert.Nil(t, result, "Backends.ToInterfaces() should return nil if there are no backends")
}

// ----------------------------------------------
// BackendMember TESTS
// ----------------------------------------------

// Tests that the BackendMember.String() function behaves correctly.
func Test_BackendMember_String(t *testing.T) {
	m := BackendMember{Name: "Test Backend"}
	expected := fmt.Sprintf("%s-%s", m.Name, m.Version)
	assert.Equal(t, m.String(), expected, "BackendMember.String() returned unexpected result")
}

// Tests that the BackendMembers.ToInterfaces() function behaves correctly.
func Test_BackendMembers_ToInterfaces(t *testing.T) {
	members := BackendMembers{BackendMember{Name: "first"}, BackendMember{Name: "second"}}
	result := members.ToInterfaces()
	assert.Equal(t, len(result), len(members), "BackendMembers.ToInterfaces() returns an array of an unexpected length")
}

// Tests that the BackendMembers.ToInterfaces() function returns the expected value when the array is empty.
func Test_BackendMembers_ToInterfaces_Empty(t *testing.T) {
	members := make(BackendMembers, 0)
	result := members.ToInterfaces()
	assert.Nil(t, result, "BackendMembers.ToInterfaces() should return nil if there are no backends")
}

// ----------------------------------------------
// Frontend TESTS
// ----------------------------------------------

// Tests that the Frontend.String() func behaves correctly.
func Test_Frontend_String(t *testing.T) {
	f := Frontend{Name: "TestFrontend"}
	assert.EnsureEqual(t, f.String(), f.Name, "Frontend.String() returned an unexpected result")
}

// Tests that the Frontends.ToInterfaces() function behaves correctly.
func Test_Frontends_ToInterfaces(t *testing.T) {
	frontends := Frontends{&Frontend{Name: "first"}, &Frontend{Name: "second"}}
	result := frontends.ToInterfaces()
	assert.Equal(t, len(result), len(frontends), "Frontends.ToInterfaces() returns an array of an unexpected length")
}

// Tests that the Frontends.ToInterfaces() function returns the expected value when the array is empty.
func Test_Frontends_ToInterfaces_Empty(t *testing.T) {
	frontends := make(Frontends, 0)
	result := frontends.ToInterfaces()
	assert.Nil(t, result, "Frontends.ToInterfaces() should return nil if there are no frontends")
}
