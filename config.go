package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config stores configuration information.
type Config struct {
	Port            string `json:"port"`
	HAConfigPath    string `json:"haconfig"`
	HATemplatePath  string `json:"hatemplate"`
	HAReloadCommand string `json:"hareload"`
	DBPath          string `json:"db-path"`
}

// GetConfig retrieves configuration information for the application.
func GetConfig() (*Config, []error) {
	config := &Config{
		Port:            "8080",
		HAConfigPath:    "/etc/haproxy/haproxy.cfg",
		HATemplatePath:  "haproxy.tmpl",
		HAReloadCommand: "service haproxy reload",
		DBPath:          "/var/db/conduit",
	}

	port := flag.String("port", "", "port the rest server will listen on")
	haconfig := flag.String("haconfig", "", "the path to the haproxy config file")
	hatemplate := flag.String("hatemplate", "", "the path to the haproxy config template file")
	hareload := flag.String("hareload", "", "the command to execute to reload HAProxy config")
	dbPath := flag.String("db-path", "", "Location to read or create database files")
	file := flag.String("f", "", "config file")
	flag.Parse()

	// if a config file was specified, load config from it
	if *file != "" {
		if err := readConfigFile(*file, config); err != nil {
			return nil, []error{err}
		}
	}

	// if flags are specified, overwrite any config file values with them
	if *port != "" {
		config.Port = *port
	}
	if *haconfig != "" {
		config.HAConfigPath = *haconfig
	}
	if *hatemplate != "" {
		config.HATemplatePath = *hatemplate
	}
	if *hareload != "" {
		config.HAReloadCommand = *hareload
	}
	if *dbPath != "" {
		config.DBPath = *dbPath
	}

	// validate the loaded config values
	if errs := validateConfig(config); errs != nil {
		return nil, errs
	}

	return config, nil
}

// readConfigFile reads config information into the given config instance from the specified configuration file.
func readConfigFile(path string, config *Config) error {
	file, err := os.Open(path)
	defer file.Close()
	if os.IsNotExist(err) {
		return fmt.Errorf("Config file '%s' does not exist", path)
	}

	if err == nil {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(config)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("Unable to load config file '%s': %v", path, err)
}

// validateConfig determines if the given Config instance contains valid values.
func validateConfig(config *Config) []error {
	// validate port
	errs := []error{}
	if config.Port == "" {
		errs = append(errs, fmt.Errorf("a port value is required"))
	} else {
		i, err := strconv.Atoi(config.Port)
		if err != nil || i < 1 || i > 65535 {
			errs = append(errs, fmt.Errorf("port value '%s' is invalid - must be an integer from 1-65535", config.Port))
		}
	}

	// validate haconfig
	if config.HAConfigPath == "" {
		errs = append(errs, fmt.Errorf("an haconfig value is required"))
	}

	// validate hatemplate
	if config.HATemplatePath == "" {
		errs = append(errs, fmt.Errorf("an hatemplate value is required"))
	}

	// validate reload
	if config.HAReloadCommand == "" {
		errs = append(errs, fmt.Errorf("a hareload value is required"))
	}

	// validate db-path
	if config.DBPath == "" {
		errs = append(errs, fmt.Errorf("a database path value is required"))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
