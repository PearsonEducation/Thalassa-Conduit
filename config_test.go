package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

type configTestCase struct {
	Name             string
	ConfigPath       string
	ExpPort          string
	ExpHAConfig      string
	ExpHATemplate    string
	ExpHAReload      string
	ExpDBPath        string
	CommandLineFlags []string
}

// ----------------------------------------------
// GetConfig TESTS
// ----------------------------------------------

func getConfigTestCases() []configTestCase {
	return []configTestCase{
		{
			Name:             "GetDefaultConfig",
			ConfigPath:       "",
			ExpPort:          "8080",
			ExpHAConfig:      "/etc/haproxy/haproxy.cfg",
			ExpHATemplate:    "haproxy.tmpl",
			ExpHAReload:      "service haproxy reload",
			ExpDBPath:        "/var/db/conduit",
			CommandLineFlags: []string{},
		},
		{
			Name:             "GetConfigFromFile",
			ConfigPath:       "test-fixtures/config.json",
			ExpPort:          "3000",
			ExpHAConfig:      "test-fixtures/haproxy.cfg",
			ExpHATemplate:    "text-fixtures/template.txt",
			ExpHAReload:      "reload haproxy",
			ExpDBPath:        "/var/db/conduit",
			CommandLineFlags: []string{},
		},
		{
			Name:          "GetConfigWithFlags",
			ConfigPath:    "",
			ExpPort:       "9000",
			ExpHAConfig:   "/var/lib/haproxy.cfg",
			ExpHATemplate: "haproxy.txt",
			ExpHAReload:   "reload",
			ExpDBPath:     "/var/lib/conduit",
			CommandLineFlags: []string{
				"conduit",
				"-port", "9000",
				"-haconfig", "/var/lib/haproxy.cfg",
				"-hatemplate", "haproxy.txt",
				"-hareload", "reload",
				"-db-path", "/var/lib/conduit",
			},
		},
		{
			Name:          "GetConfigWithFlagsFromFile",
			ConfigPath:    "",
			ExpPort:       "3000",
			ExpHAConfig:   "/var/lib/haproxy.cfg",
			ExpHATemplate: "haproxy.txt",
			ExpHAReload:   "reload haproxy",
			ExpDBPath:     "/var/lib/conduit",
			CommandLineFlags: []string{
				"conduit",
				//"-port", "9000", //Port is explicitly not specified
				"-haconfig", "/var/lib/haproxy.cfg",
				"-hatemplate", "haproxy.txt",
				//"-hareload", "reload", //Reload command is explicitly not specified
				"-db-path", "/var/lib/conduit",
				"-f", "./test-fixtures/config.json",
			},
		},
	}
}

// Tests that the GetConfig() function behaves properly.
func Test_GetConfig(t *testing.T) {
	for _, input := range getConfigTestCases() {

		if len(input.CommandLineFlags) > 0 {
			os.Args = input.CommandLineFlags
		}

		config := &Config{}
		if input.ConfigPath == "" {
			conf, err := GetConfig()
			if !assert.Nil(t, err, "Error getting config: '%v' in test '%v'", err, input.Name) {
				continue
			}
			config = conf
		} else {
			err := readConfigFile(input.ConfigPath, config)
			if !assert.Nil(t, err, "Error loading config: '%v' in test '%v'", err, input.Name) {
				continue
			}
		}

		assert.Equal(t, config.Port, input.ExpPort, fmt.Sprintf("Config.Port has incorrect value in test '%s'", input.Name))
		assert.Equal(t, config.HAConfigPath, input.ExpHAConfig, fmt.Sprintf("Config.HAConfigPath has incorrect value in test '%s'", input.Name))
		assert.Equal(t, config.HATemplatePath, input.ExpHATemplate, fmt.Sprintf("Config.HATemplatePath has incorrect value in test '%s'", input.Name))
		assert.Equal(t, config.HAReloadCommand, input.ExpHAReload, fmt.Sprintf("Config.ReloadHAConfig has incorrect value in test '%s'", input.Name))
		assert.Equal(t, config.DBPath, input.ExpDBPath, fmt.Sprintf("Config.DBPath has incorrect value in test '%s'", input.Name))

		// resets the command line arguments with a new Flag Set
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		flag.Usage = nil
	}
}

// ----------------------------------------------
// validateConfig TESTS
// ----------------------------------------------

// Tests that the validateConfig() function properly validates correct config values.
func Test_validateConfig(t *testing.T) {
	config := &Config{}
	err := readConfigFile("test-fixtures/config.json", config)
	assert.EnsureNil(t, err, "readConfigFile() returned an unexpected error: %v", err)

	errs := validateConfig(config)
	assert.Empty(t, errs, "validateConfig() returned non-empty slice of errors: %v", errs)
}

// Tests that the validateConfig() function properly invalidates bad config values.
func Test_validateConfig_InvalidValues(t *testing.T) {
	config := &Config{
		Port:            "1234567890", //invalid
		HAConfigPath:    "",           //invalid
		HAReloadCommand: "",           //invalid
		DBPath:          "",           //invalid
		//HATemplate field is missing    //invalid
	}

	errs := validateConfig(config)
	assert.EnsureEqual(t, len(errs), 5, "validateConfig() returned unexpected error count")

	config.Port = ""
	errs = validateConfig(config)
	assert.EnsureEqual(t, len(errs), 5, "validateConfig() returned unexpected error count")
}
