package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/go-martini/martini"
)

const (
	defaultTimeout = "5s"

	helpText = `
Thalassa Conduit

Conduit provides a REST API for interacting with HAProxy and making changes
to its configuration file.  Conduit also provides the ability to store
additional metadata that cannot be stored in the config file.

Usage: %s [options]

Options:
   -help               show this help page
   -port=8080          port the REST server will listen on
   -haconfig=path      path to the HAProxy config file
   -hatemplate=path    path to the HAProxy config template file
   -hareload=cmd       shell command to reload HAProxy config
   -f=path             path to a config file, overwrites CLI flags
   -db-path=path       path to location of database files

`
)

// Server represents an http server.
type Server interface {
	Run(config *Config, dbMgr DBManager, tmpl *template.Template)
	Stop()
	SignalShutdown()
	SignalRestart()
}

type serverImpl struct {
	stopChan   chan int
	signalChan chan os.Signal
	shutdown   bool
	timeout    time.Duration
}

// NewServer returns a new Server instance.
func NewServer(timeout time.Duration, signalChan chan os.Signal) Server {
	return &serverImpl{
		stopChan:   make(chan int, 1),
		signalChan: signalChan,
		shutdown:   false,
		timeout:    timeout,
	}
}

// Run will start the server.
func (s *serverImpl) Run(config *Config, dbMgr DBManager, tmpl *template.Template) {
	m := initMartini(s, config, dbMgr, tmpl)
	server := &http.Server{Addr: ":" + config.Port, Handler: m}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Printf("%v", err)
		s.signalChan <- syscall.SIGINT
		return
	}

	go func() {
		<-s.stopChan
		server.SetKeepAlivesEnabled(false)
		listener.Close()
	}()

	log.Printf("Running on port " + config.Port)
	if err = server.Serve(listener); err != nil {
		if !s.shutdown {
			log.Printf("%v", err)
			s.signalChan <- syscall.SIGINT
		}
	}
}

// Stop will stop the server.
func (s *serverImpl) Stop() {
	if !s.shutdown {
		s.shutdown = true
		s.stopChan <- 1
		<-time.After(s.timeout)
		close(s.stopChan)
	}
}

// SignalShutdown will stop the server and send a shutdown signal (SIGINT) on the signal channel.
func (s *serverImpl) SignalShutdown() {
	s.Stop()
	s.signalChan <- syscall.SIGINT
}

// SignalShutdown will stop the server and send a restart signal (SIGHUP) on the signal channel.
func (s *serverImpl) SignalRestart() {
	s.Stop()
	s.signalChan <- syscall.SIGHUP
}

func main() {
	// create channel for handling application reload and exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	for {
		if exit, code := start(signalChan); exit {
			os.Exit(code)
		}
		// resets the command line arguments with a new Flag Set
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		flag.Usage = nil
	}
}

// starts the application
func start(signalChan chan os.Signal) (bool, int) {
	// show help text if -help flag is specified
	if len(os.Args) == 2 && os.Args[1] == "-help" {
		fmt.Printf(helpText, os.Args[0])
		os.Exit(0)
	}

	// load config
	config, errs := GetConfig()
	if errs != nil {
		os.Stderr.WriteString("error loading config data:\n")
		for _, e := range errs {
			os.Stderr.WriteString(fmt.Sprintf("  %v\n", e))
		}
		os.Stderr.WriteString("\n")
		return true, 1
	}

	// initialize DB manager
	dbManager, err := NewDBManager(config)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error initializing DB manager: %s", err.Error()))
		return true, 1
	}
	defer dbManager.Close()

	// load haproxy config template
	t, err := ioutil.ReadFile(config.HATemplatePath)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error loading HAProxy config template file: %s", err.Error()))
		return true, 1
	}
	template, _ := template.New("test").Parse(string(t))

	// start the web server in a new goroutine and listen for
	d, _ := time.ParseDuration(defaultTimeout)
	server := NewServer(d, signalChan)
	go server.Run(config, dbManager, template)

	return waitForSignal(signalChan, server)
}

func waitForSignal(signalChan chan os.Signal, server Server) (bool, int) {
	select {
	case sig := <-signalChan:
		switch sig {
		case syscall.SIGHUP:
			log.Printf("[INFO] Received SIGHUP signal - reloading configuration")
			server.Stop()
			return false, 0
		default:
			log.Printf("[WARN] Received %v signal - shutting down", sig)
			server.Stop()
			return true, 0
		}
	}
}

func initMartini(server Server, config *Config, dbMgr DBManager, tmpl *template.Template) *martini.Martini {
	m := martini.New()

	// setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(SetContentTypeHeader)

	// setup routes
	r := martini.NewRouter()
	r.Get(`/status`, GetStatus)
	r.Get(`/haproxy/config`, GetHAProxyConfig)
	r.Get(`/haproxy/reload`, ReloadHAProxy)
	r.Get(`/restart`, GetRestart)

	r.Get(`/frontends`, GetFrontends)
	r.Get(`/frontends/:name`, GetFrontend)
	r.Put(`/frontends/:name`, PutFrontend)
	r.Post(`/frontends/:name`, PostFrontend)
	r.Delete(`/frontends/:name`, DeleteFrontend)

	r.Get(`/backends`, GetBackends)
	r.Get(`/backends/:name`, GetBackend)
	r.Put(`/backends/:name`, PutBackend)
	r.Post(`/backends/:name`, PostBackend)
	r.Delete(`/backends/:name`, DeleteBackend)
	r.Get(`/backends/:name/members`, GetBackendMembers)

	// dependency injection
	ha := NewHAProxy(config.HAConfigPath, tmpl, config.HAReloadCommand)
	m.MapTo(JSONEncoder{}, (*Encoder)(nil))
	m.MapTo(NewDataSvc(dbMgr.NewDatastore(), ha), (*DataSvc)(nil))
	m.MapTo(ha, (*HAProxy)(nil))
	m.MapTo(server, (*Server)(nil))

	// add router action
	m.Action(r.Handle)
	return m
}

// SetContentTypeHeader sets the Content-Type header for all responses.
func SetContentTypeHeader(c martini.Context, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// GetStatus is a REST handler that will return the application status.
func GetStatus() (int, string) {
	//TODO: add logic to determine if app is healthy
	return http.StatusOK, `{"status":"ok"}`
}

// GetRestart is a REST handler that will stop the http server, reload configuration, and restart it.
func GetRestart(server Server) (int, string) {
	server.SignalRestart()
	return http.StatusOK, ""
}
