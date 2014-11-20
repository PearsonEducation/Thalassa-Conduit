package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

const (
	defaultTemplate = `global
  maxconn 256

  defaults
    timeout connect 5000ms

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

// HAProxy represents an HAProxy installation
type HAProxy interface {
	Template() *template.Template
	SetTemplate(*template.Template)
	GetConfig() (string, error)
	GetFrontends() (Frontends, error)
	GetBackends() (Backends, error)
	WriteConfig(frontends Frontends, backends Backends) error
	ReloadConfig() error
}

type haProxyImpl struct {
	configPath string
	template   *template.Template
	reloadCmd  string
}

// NewHAProxy returns a new populated instance of an HAProxy struct.
func NewHAProxy(configPath string, configTemplate *template.Template, reloadCmd string) HAProxy {
	if configTemplate == nil {
		configTemplate, _ = template.New("test").Parse(string(defaultTemplate))
	}
	return &haProxyImpl{
		configPath: configPath,
		template:   configTemplate,
		reloadCmd:  reloadCmd,
	}
}

// Template returns the HAProxy config template.
func (h *haProxyImpl) Template() *template.Template {
	return h.template
}

// SetTemplate sets the value of the HAProxy config template.
func (h *haProxyImpl) SetTemplate(t *template.Template) {
	h.template = t
}

// GetConfig returns the contents of the HAProxy config file associated with this HAProxy instance.
func (h *haProxyImpl) GetConfig() (string, error) {
	c, err := ioutil.ReadFile(h.configPath)
	if err != nil {
		return "", err
	}
	return string(c), nil
}

// GetFrontends returns the Frontends in the HAProxy config file associated with this HAProxy instance.
func (h *haProxyImpl) GetFrontends() (Frontends, error) {
	s, err := h.GetConfig()
	if err != nil {
		return nil, err
	}
	frontends := Frontends{}
	lines := h.parseConfigText(s)
	// iterate through all the lines of the haproxy config file
	for i, l := range lines {
		// if line begins with "frontend ", then this line and the following lines (until an empty line
		// is reached) all contain information about a single frontend
		if strings.HasPrefix(l, "frontend ") && len(l) > 9 {
			f := &Frontend{}
			f.Name = l[9:]
			index := i + 1
			subline := lines[index]
			// loop through lines until an empty line or the end of the file is reached
			for subline != "" {
				if strings.HasPrefix(subline, "bind ") && len(subline) > 5 {
					f.Bind = subline[5:]
				} else if strings.HasPrefix(subline, "mode ") && len(subline) > 5 {
					f.Mode = subline[5:]
				} else if strings.HasPrefix(subline, "default_backend ") && len(subline) > 16 {
					f.DefaultBackend = subline[16:]
				} else if strings.HasPrefix(subline, "option ") && len(subline) > 7 {
					f.Option = subline[7:]
				}
				index++
				// if no more lines then it's EOF, so break
				if index == len(lines) {
					break
				}
				subline = lines[index]
			}
			frontends = append(frontends, f)
		}
	}
	return frontends, nil
}

// GetBackends returns the Backends in the HAProxy config file associated with this HAProxy instance.
func (h *haProxyImpl) GetBackends() (Backends, error) {
	s, err := h.GetConfig()
	if err != nil {
		return nil, err
	}
	backends := Backends{}
	lines := h.parseConfigText(s)
	// iterate through all the lines of the haproxy config file
	for i, l := range lines {
		// if line begins with "frontend ", then this line and the following lines (until an empty line
		// is reached) all contain information about a single frontend
		if strings.HasPrefix(l, "backend ") && len(l) > 8 {
			b := &Backend{}
			m := BackendMembers{}
			b.Name = l[8:]
			index := i + 1
			subline := lines[index]
			// loop through lines until an empty line or the end of the file is reached
			for subline != "" {
				if strings.HasPrefix(subline, "balance ") && len(subline) > 8 {
					b.Balance = subline[8:]
				} else if strings.HasPrefix(subline, "mode ") && len(subline) > 5 {
					b.Mode = subline[5:]
				} else if strings.HasPrefix(subline, "server ") && len(subline) > 7 {
					// each backend member is on a single line - parse data from the line to populate
					// a BackendMember instance
					parts := strings.Split(subline[7:], " ")
					if len(parts) != 5 {
						return nil, fmt.Errorf("haproxy config file is invalid - could not read members for backend %s", b.Name)
					}
					j := strings.Index(parts[1], ":")
					if j < 0 {
						return nil, fmt.Errorf("haproxy config file is invalid - could not read members for backend %s", b.Name)
					}
					host := parts[1][:j]
					port, err := strconv.Atoi(parts[1][j+1:])
					if err != nil {
						return nil, fmt.Errorf("haproxy config file is invalid - could not read members for backend %s", b.Name)
					}
					m = append(m, BackendMember{
						Name: parts[0],
						Host: host,
						Port: port,
					})
				}
				index++
				// if no more lines then it's EOF, so break
				if index == len(lines) {
					break
				}
				subline = lines[index]
			}
			b.Members = m
			backends = append(backends, b)
		}
	}
	return backends, nil
}

// WriteConfig replaces the existing HAProxy config file with a new one created from the config template
// with the given frontends and backends.
func (h *haProxyImpl) WriteConfig(frontends Frontends, backends Backends) error {
	data := struct {
		Frontends Frontends
		Backends  Backends
	}{
		frontends,
		backends,
	}

	var buffer bytes.Buffer
	if err := h.template.Execute(&buffer, data); err != nil {
		return err
	}
	if err := ioutil.WriteFile(h.configPath, []byte(buffer.String()), 0644); err != nil {
		return err
	}
	return nil
}

// ReloadConfig executes a command to tell HAProxy to reload it's config file.
func (h *haProxyImpl) ReloadConfig() error {
	cmdStr := h.reloadCmd
	if cmdStr == "" {
		cmdStr = "service haproxy reload"
	}
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// parses the contents of the haproxy config file into a slice of string values, each of
// which is a line of text in the file that has been trimmed of whitespace
func (h *haProxyImpl) parseConfigText(s string) []string {
	lines := strings.Split(s, "\n")
	r := make([]string, len(lines))
	for i, l := range lines {
		r[i] = strings.TrimSpace(l)
	}
	return r
}
