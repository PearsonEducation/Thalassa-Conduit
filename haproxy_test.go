package main

import (
	"os"
	"reflect"
	"testing"
	"text/template"
)

const (
	testConfigPath = "test-fixtures/haproxy.cfg"
	testTemplate   = `global
  maxconn 128

  defaults
    timeout connect 1000ms

{{range .Frontends}}
  frontend {{.Name}}{{if .Bind}}
    bind {{.Bind}}{{end}}{{if .Mode}}
    mode {{.Mode}}{{end}}{{if .DefaultBackend}}
    default_backend {{.DefaultBackend}}{{end}}{{if .Option}}
    option {{.Option}}{{end}}
{{end}}
{{range .Backends}}
  backend {{.Name}}{{if .Mode}}
    mode {{.Mode}}{{end}}{{if .Balance}}
    balance {{.Balance}}{{end}}{{range .Members}}
    server {{.Name}} {{.Host}}:{{.Port}} check inter 2000{{end}}
{{end}}
`
)

// ----------------------------------------------
// NewHAProxy TESTS
// ----------------------------------------------

// Tests the "happy path" for the NewHAProxy() function.
func Test_NewHAProxy(t *testing.T) {
	tmpl, _ := template.New("test").Parse(testTemplate)
	h := NewHAProxy(testConfigPath, tmpl, "")
	assert.EnsureNotNil(t, h, "NewHAProxy() returned a nil value")
	assert.Equal(t, reflect.TypeOf(h), reflect.TypeOf(&haProxyImpl{}), "NewHAProxy() returned an unexpected object type")
}

// Tests that the NewHAProxy() function assigns a default template if none is passed it.
func Test_NewHAProxy_NilTemplate(t *testing.T) {
	h := NewHAProxy(testConfigPath, nil, "")
	assert.EnsureNotNil(t, h, "NewHAProxy() returned a nil value")
	assert.Equal(t, reflect.TypeOf(h), reflect.TypeOf(&haProxyImpl{}), "NewHAProxy() returned an unexpected object type")

	expTmpl, _ := template.New("test").Parse(string(defaultTemplate))
	assert.Equal(t, h.Template(), expTmpl, "NewHAProxy() did not use the default template when none was provided")
}

// ----------------------------------------------
// haProxyImpl.Template TESTS
// ----------------------------------------------

func Test_haProxyImpl_Template(t *testing.T) {
	tmpl, _ := template.New("test").Parse("testing")
	h := &haProxyImpl{
		template: tmpl,
	}
	assert.Equal(t, h.Template(), tmpl, "haProxyImpl.Template() returned an unexpected value")
}

// ----------------------------------------------
// haProxyImpl.SetTemplate TESTS
// ----------------------------------------------

func Test_haProxyImpl_SetTemplate(t *testing.T) {
	h := &haProxyImpl{}
	tmpl, _ := template.New("test").Parse("testing")
	h.SetTemplate(tmpl)
	assert.Equal(t, h.template, tmpl, "haProxyImpl.SetTemplate() did not assign the template to the expected value")
}

// ----------------------------------------------
// haProxyImpl.GetConfig TESTS
// ----------------------------------------------

// Tests the "happy path" for the haProxyImpl.GetConfig() function.
func Test_haProxyImpl_GetConfig(t *testing.T) {
	tmpl, _ := template.New("test").Parse(testTemplate)
	h := &haProxyImpl{
		configPath: testConfigPath,
		template:   tmpl,
	}
	s, err := h.GetConfig()
	assert.EnsureNil(t, err, "haProxyImpl.GetConfig() returned an unexpected error: %v", err)
	assert.NotEmpty(t, s, "haProxyImpl.GetConfig() returned no data; expected contents of text-fixtures/haproxy.cfg")
}

// ----------------------------------------------
// haProxyImpl.GetFrontends TESTS
// ----------------------------------------------

// Tests the "happy path" for the haProxyImpl.GetFrontends() function.
func Test_haProxyImpl_GetFrontends(t *testing.T) {
	tmpl, _ := template.New("test").Parse(testTemplate)
	h := &haProxyImpl{
		configPath: testConfigPath,
		template:   tmpl,
	}
	f, err := h.GetFrontends()
	assert.EnsureNil(t, err, "haProxyImpl.GetFrontends() returned an unexpected error: %v", err)
	assert.EnsureEqual(t, len(f), 1, "haProxyImpl.GetFrontends() returned unexpected number of objects")

	expFrontend := &Frontend{
		Name:           "app",
		Bind:           "*:80",
		Mode:           "http",
		DefaultBackend: "app-1",
		Option:         "httplog",
	}
	assert.Equal(t, f[0], expFrontend, "haProxyImpl.GetFrontends() returned unexpected object")
}

// ----------------------------------------------
// haProxyImpl.GetBackends TESTS
// ----------------------------------------------

// Tests the "happy path" for the haProxyImpl.GetBackends() function.
func Test_haProxyImpl_GetBackends(t *testing.T) {
	tmpl, _ := template.New("test").Parse(testTemplate)
	h := &haProxyImpl{
		configPath: testConfigPath,
		template:   tmpl,
	}
	b, err := h.GetBackends()
	assert.EnsureNil(t, err, "haProxyImpl.GetBackends() returned an unexpected error: %v", err)
	assert.EnsureEqual(t, len(b), 2, "haProxyImpl.GetBackends() returned unexpected number of objects")

	expBackend := &Backend{
		Name: "app-2",
		Mode: "http",
		Members: BackendMembers{
			BackendMember{
				Name: "app2_node1",
				Host: "10.2.2.10",
				Port: 8080,
			},
			BackendMember{
				Name: "app2_node2",
				Host: "10.2.2.20",
				Port: 8080,
			},
		},
	}
	assert.Equal(t, b[1], expBackend, "haProxyImpl.GetBackends() returned unexpected object")
}

// ----------------------------------------------
// haProxyImpl.WriteConfig TESTS
// ----------------------------------------------

// Tests the "happy path" for the haProxyImpl.WriteConfig() function.
func Test_haProxyImpl_WriteConfig(t *testing.T) {
	testFile := "test-fixtures/test.cfg"
	defer os.Remove(testFile)

	tmpl, _ := template.New("test").Parse(testTemplate)
	h := &haProxyImpl{
		configPath: testFile,
		template:   tmpl,
	}
	frontends := Frontends{
		&Frontend{
			Name:           "test-app",
			Bind:           "*:80",
			Mode:           "http",
			DefaultBackend: "test-app-1",
		},
	}
	backends := Backends{
		&Backend{
			Name: "test-app-1",
			Mode: "http",
			Members: BackendMembers{
				BackendMember{
					Name: "testapp1_node1",
					Host: "10.2.2.10",
					Port: 8080,
				},
			},
		},
	}
	err := h.WriteConfig(frontends, backends)
	assert.EnsureNil(t, err, "haProxyImpl.WriteConfig() returned an unexpected error: %v", err)

	f, _ := h.GetFrontends()
	assert.EnsureEqual(t, len(f), 1, "haProxyImpl.GetFrontends() returned unexptected number of objects")

	b, _ := h.GetBackends()
	assert.EnsureEqual(t, len(b), 1, "haProxyImpl.GetBackends() returned unexptected number of objects")

	assert.Equal(t, f[0], frontends[0], "haProxyImpl.GetFrontends() returned unexpected object")
	assert.Equal(t, b[0], backends[0], "haProxyImpl.GetBackends() returned unexpected object")
}
