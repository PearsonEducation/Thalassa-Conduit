Conduit
=======

# Overview

Conduit is a wrapper for HAProxy that exposes a REST API for manipulating the HAProxy configuration file.  In addition, it adds the ability to associate metadata with Backends and Frontends, as well as associate backends and backend members with an application version number.  Conduit enables no-downtime deployments and supports running multiple versions of your application simultaneously as well as easy rollbacks.

Conduit is meant as a stand-alone REST wrapper, and is not part of the existing Thalassa suite of applications (i.e. Aqueduct, Server).  The Thalassa prefix is part of the name for legal reasons, but we will refer to this product as 'Conduit' throughout the documentation.

## Current Status

Conduit is currently at version 0.0.1 and is considered Alpha software.  Though the API is stable, there are still a few backlog items that we would like to address:

* address remaining TODOs in the code
* determine if any additional HAProxy configuration fields should be supported by the API
* since Martini is not idiomatic Go, explore using other REST routing and/or handler packages
* fill in any testing gaps
* analyze unit test coverage and fill in any gaps
* add integration tests that hit a running instance of the application

## HAProxy Fundamentals

Conduit does not try to obfuscate HAProxy, and it's important to know the fundamentals of how HAProxy works to understand Conduit. The API mirrors HAProxy's semantics. The [HAProxy documentation](http://cbonte.github.io/haproxy-dconv/configuration-1.4.html) contains a wealth of detailed information.

1. **Frontends** - A "frontend" section describes a set of listening sockets accepting client connections.

2. **Backends** - A "backend" section describes a set of servers to which the proxy will connect to forward incoming connections.

3. **Members/Servers** - Conduit calls the servers that *backends* route to "members". In other words, members of a backend pool of servers.

4. **Config file** - At startup, HAProxy loads a configuration file and never looks at that file again. Conduit manages this by re-templating a new config file and gracefully restarting the HAProxy process.

# Installation

    go get github.com/PearsonEducation/Thalassa-Conduit
    cd $GOPATH/src/github.com/PearsonEducation/Thalassa-Conduit

You can then use the commands defined in the makefile to perform different actions:

* `make build` - Builds the Conduit executable binary file. Performs `clean` first.
* `make clean` - Removes the Conduit executable binary file, the pkg directory for vendorized files and any coverage reports.
* `make install` - Installs the Conduit pkgs. Performs `build` first.
* `make run` - Runs the Conduit application. Performs `build` first.
* `make test` - Runs the Conduit unit tests. Performs `install` first. 
* `make cover` - Runs the Conduit unit tests and produces coverage reports. Performs `install` first.
* `make cover-report` - Opens the Conduit coverage reports html view. Performs `cover` first.
* `make dep-install` - Runs `go get`.

Conduit uses an embedded [LevelDB](http://en.wikipedia.org/wiki/LevelDB) data store written for Go ([goleveldb](https://github.com/syndtr/goleveldb)).  LevelDB stores data in a file, and so you will need to ensure that the Conduit has read/write permissions to the file location.  By default, the file is created at /var/db/conduit, but you can specify an alternative location in a Conduit config file or a command line flag (details in the next section).

# Running

The following commands line flags are supported when running Conduit from the command line:

    -help               show the help page
    -port=##            port the REST server will bind to          [default: "10000"]
    -haconfig=path      path to the HAProxy config file            [default: "/etc/haproxy/haproxy.cfg"]
    -hatemplate=path    path to the HAProxy config template file   [default: "haproxy.tmpl"]
    -hareload=cmd       shell command to reload HAProxy config     [default: "service haproxy reload"]
    -db-path=path       path to database file                      [default: "/var/db/conduit"]
    -f=path             path to a config file

Instead of passing in numerous flags, you can create a JSON config file with the values and use the '-f' flag to point Conduit to the file.  For example, a sample config file may look like this:

    {
        "port": "10000",
        "haconfig": "/etc/haproxy/haproxy.cfg",
        "hatemplate": "haproxy.tmpl",
        "hareload": "service haproxy reload",
        "db-path": "/var/db/conduit"
    }

# REST API

### GET `/frontends`

Returns an array of objects for all of the frontends configured for this Conduit server.

For example:

    [{
        "name": "myapp",
        "bind": "*:8080,*:80",
        "defaultBackend": "live",
        "mode": "http",
        "keepalive": "default",
        "option": "httplog"
    }]

### GET `/frontends/{name}`

Get a specific frontend by its name.  Expect a response status of `200`, or `404` if it doesn't exist.

### PUT `/frontends/{name}`

Create or update a frontend by its name.  Use a `Content-Type` of `application/json` and a body like:

    [{
        "name": "myapp",
        "bind": "*:8080,*:80",
        "defaultBackend": "live",
        "mode": "http",
        "keepalive": "default",
        "option": "httplog"
    }]

Expect a response status of `201` if a new frontend gets created or `200` if an existing frontend is updated.

#### Routing Rules

There are currently 3 types of rules that can be applied to frontends: `path`, `url`, and `header`.

Path rules support `path`, `path_beg`, and `path_reg` HAProxy operations

    {
        "type": "path"
      , "operation": "path|path_beg|path_reg"
      , "value": "favicon.ico|/ecxd/|^/article/[^/]*$"
      , "backend": "foo" // if rule is met, the backend to route the request to
    }


Url rules support `url`, `url_beg`, and `url_reg` HAProxy operations

    {
        "type": "url"
      , "operation": "url|url_beg|url_reg"
      , "value": "/bar" // value for the operation
      , "backend": "bar" // if rule is met, the backend to route the request to
    }

Header rules support `hdr_dom` with a entire value at this point

    {
        "type": "header"
      , "header": "host"            // the name of the HTTP header
      , "operation": "hdr_dom"
      , "value": "baz.com"
      , "backend": "baz" // if rule is met, the backend to route the request to
    }

#### Raw Config

The `rawConfig` property is an end around way to insert raw lines of config for frontends and backends. Use them sparingly but use them if you need them.

### POST `/frontends/{name}`

Perform an update of a frontend by its name, and can be used to update one or more fields of a frontend.  Use a `Content-Type` of `application/json` and expect a response status of `200`, or `404` if it doesn't exist.

### DELETE `/frontends/{name}`

Delete a specific frontend by its name.  Expect a response status of `200`, or `404` if it doesn't exist.

### GET `/backends`

Returns an array of objects for all of the backends configured for this Conduit server.

For example:

    [{
        "name": "live",
        "version": "1.0.0",
        "balance": "roundrobin",
        "host": "",
        "mode": "http",
        "members": [
            {
                "name": "myapp",
                "version": "1.0.0",
                "host": "10.10.240.121",
                "port": 8080,
                "lastKnown": ""
            },
            {
                "name": "myapp",
                "version": "1.0.0",
                "host": "10.10.240.80",
                "port": 8080,
                "lastKnown": ""
            }
        ]
    }]

### GET `/backends/{name}`

Get a specific backend by its name.  Expect a response status of `200`, or `404` if it doesn't exist.

### PUT `/backends/{name}`

Create or update a backend by its name.  Use a `Content-Type` of `application/json` and a body like:

    [{
        "name": "live",
        "version": "1.0.0",
        "balance": "roundrobin",
        "host": "",
        "mode": "http",
        "members": [
            {
                "name": "myapp",
                "version": "1.0.0",
                "host": "10.10.240.121",
                "port": 8080,
                "lastKnown": ""
            },
            {
                "name": "myapp",
                "version": "1.0.0",
                "host": "10.10.240.80",
                "port": 8080,
                "lastKnown": ""
            }
        ]
    }]

Expect a response status of `201` if a new backend gets created or `200` if an existing backend is updated.

### POST `/backends/{name}`

Perform an update of a backend by its name, and can be used to update one or more fields of a backend.  Use a `Content-Type` of `application/json` and expect a response status of `200`, or `404` if it doesn't exist.

### DELETE `/backends/{name}`

Delete a specific backend by its name.  Expect a response status of `200`, or `404` if it doesn't exist.

### GET `/backends/{name}/members`

Get the members of a specific backend by its name.  Expext a response status of `200`, or `404` if the backend doesn't exist.

### GET `/haproxy/config`

Return the current contents of the HAProxy config file.

    global
      log 127.0.0.1 local0
      log 127.0.0.1 local1 notice
      daemon
      maxconn 4096
      user haproxy 
      group haproxy 
      stats socket /tmp/haproxy.status.sock user appuser level admin

      defaults
        log global
        option dontlognull
        option redispatch
        retries 3
        maxconn 2000
        timeout connect 5000ms
        timeout client 50000ms
        timeout server 50000ms

      listen stats :1988
        mode http
        stats enable
        stats uri /
        stats refresh 2s
        stats realm Haproxy\ Stats
        stats auth showme:showme


      frontend myapp
        bind *:8080,*:80
        mode http
        default_backend live
        option httplog


      backend live
        mode http
        balance roundrobin
        server live_10.10.240.121:8080 10.10.240.121:8080 check inter 2000
        server live_10.10.240.80:8080 10.10.240.80:8080 check inter 2000

      backend staged
        mode http
        balance roundrobin
        server staged_10.10.240.174:8080 10.10.240.174:8080 check inter 2000
        server staged_10.10.240.206:8080 10.10.240.206:8080 check inter 2000

### GET `/haproxy/reload`

Signals the HAProxy process to reload its configuration file.

### GET `/restart`

Signals Conduit to reload it's configuration and restart its REST server.

# Known Limitations and Roadmap

Conduit currently doesn't implement any type of authentication or authorization and at this point expects to be running on a trusted private network. This will be addressed in the future. Ultimately auth should be extensible and customizable. Suggestions and pull requests welcome!

# License

Licensed under Apache 2.0. See [LICENSE](https://github.com/PearsonEducation/thalassa-conduit/blob/master/LICENSE) file.

# Authors

Conduit was created and is maintained by [Dave Laursen](https://github.com/davelaursen) and [Scott Engle](https://github.com/scottengle).