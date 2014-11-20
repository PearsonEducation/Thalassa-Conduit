GO ?= go
GOFILE ?= conduit.go
OUTFILE ?= conduit

GOBIN := $(GOPATH)/bin
GOPATH := $(CURDIR)/_vendor:$(GOPATH)

build: clean $(GOFILE)
	$(GO) build -v -o $(OUTFILE) *.go

clean:
	rm -f $(OUTFILE)
	rm -f ./coverage.out
	rm -rf ./_vendor/pkg

install: build
	$(GO) install

run: build
	./$(OUTFILE)

test: install
	$(GO) test -cover -short

cover: install
	$(GO) test -coverprofile=coverage.out -short

cover-report: cover
	$(GO) tool cover -html=./coverage.out

dep-install:
	$(GOBIN)/$(GO) get

