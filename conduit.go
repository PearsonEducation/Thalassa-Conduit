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

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
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

// Params is an alias for a map[string]string and represents route parameters.
type Params map[string]string

// NewServer returns a new Server instance.
func NewServer(timeout time.Duration, signalChan chan os.Signal) Server {
	return &serverImpl{
		stopChan:   make(chan int, 1),
		signalChan: signalChan,
		shutdown:   false,
		timeout:    timeout,
	}
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

	// start the web server in a new goroutine
	d, _ := time.ParseDuration(defaultTimeout)
	server := NewServer(d, signalChan)
	go server.Run(config, dbManager, template)

	// listen for a signal to restart or stop the web server
	return waitForSignal(signalChan, server)
}

// waits for a signal to reload config or shutdown the web server
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

//
func (s *serverImpl) Run(config *Config, dbMgr DBManager, tmpl *template.Template) {
	router := initRouter(s, config, dbMgr, tmpl)
	neg := initNegroni(router)

	server := &http.Server{Addr: ":" + config.Port, Handler: neg}

	// grab the listener so that we have control over it - allows us to manually close the
	// listener when we're ready to shut down the http server
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Printf("%v", err)
		s.signalChan <- syscall.SIGINT
		return
	}

	// a goroutine that will listen for a signal to stop the server
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

// func initRouter(server Server, config *Config, dbMgr DBManager, tmpl *template.Template) *mux.Router {
func initRouter(server Server, config *Config, dbMgr DBManager, tmpl *template.Template) *mux.Router {
	r := mux.NewRouter()

	// initialize values to inject into handlers
	enc := JSONEncoder{}
	ha := NewHAProxy(config.HAConfigPath, tmpl, config.HAReloadCommand)
	svc := NewDataSvc(dbMgr.NewDatastore(), ha)

	// admin routes
	r.HandleFunc(`/status`, func(w http.ResponseWriter, r *http.Request) {
		GetStatus(w)
	}).Methods("GET")

	r.HandleFunc(`/haproxy/config`, func(w http.ResponseWriter, r *http.Request) {
		GetHAProxyConfig(w, enc, ha)
	}).Methods("GET")

	r.HandleFunc(`/haproxy/reload`, func(w http.ResponseWriter, r *http.Request) {
		ReloadHAProxy(w, enc, ha)
	}).Methods("GET")

	r.HandleFunc(`/restart`, func(w http.ResponseWriter, r *http.Request) {
		GetRestart(w, server)
	}).Methods("GET")

	// frontend routes
	r.HandleFunc(`/frontends`, func(w http.ResponseWriter, r *http.Request) {
		GetFrontends(w, enc, svc)
	}).Methods("GET")

	r.HandleFunc(`/frontends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		GetFrontend(w, enc, svc, mux.Vars(r))
	}).Methods("GET")

	r.HandleFunc(`/frontends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		PutFrontend(w, r, enc, svc, mux.Vars(r))
	}).Methods("PUT")

	r.HandleFunc(`/frontends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		PostFrontend(w, r, enc, svc, mux.Vars(r))
	}).Methods("POST")

	r.HandleFunc(`/frontends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		DeleteFrontend(w, enc, svc, mux.Vars(r))
	}).Methods("DELETE")

	// backend routes
	r.HandleFunc(`/backends`, func(w http.ResponseWriter, r *http.Request) {
		GetBackends(w, enc, svc)
	}).Methods("GET")

	r.HandleFunc(`/backends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		GetBackend(w, enc, svc, mux.Vars(r))
	}).Methods("GET")

	r.HandleFunc(`/backends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		PutBackend(w, r, enc, svc, mux.Vars(r))
	}).Methods("PUT")

	r.HandleFunc(`/backends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		PostBackend(w, r, enc, svc, mux.Vars(r))
	}).Methods("POST")

	r.HandleFunc(`/backends/{name}`, func(w http.ResponseWriter, r *http.Request) {
		DeleteBackend(w, enc, svc, mux.Vars(r))
	}).Methods("DELETE")

	r.HandleFunc(`/backends/{name}/members`, func(w http.ResponseWriter, r *http.Request) {
		GetBackendMembers(w, enc, svc, mux.Vars(r))
	}).Methods("GET")

	return r
}

// initialize Negroni (middleware, handler)
func initNegroni(handler http.Handler) *negroni.Negroni {
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.Use(ContentTypeMiddleware())
	n.UseHandler(handler)
	return n
}

// ContentTypeMiddleware gets Negroni middleware that sets the Content-Type header for all responses.
func ContentTypeMiddleware() negroni.HandlerFunc {
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	})
}

// GetStatus is a REST handler that will return the application status.
func GetStatus(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// GetRestart is a REST handler that will stop the http server, reload configuration, and restart it.
func GetRestart(w http.ResponseWriter, server Server) {
	server.SignalRestart()
	w.WriteHeader(http.StatusOK)
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
