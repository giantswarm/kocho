PROJECT=kocho

BUILD_PATH := $(shell pwd)/.gobuild
GS_PATH := $(BUILD_PATH)/src/github.com/giantswarm
TEMPLATES=$(shell find default-templates -name '*.tmpl')

BIN := $(PROJECT)

SOURCE=$(shell find . -name '*.go')

VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

GOPATH := $(BUILD_PATH)
GOVERSION=1.5
ifndef GOOS
	GOOS := linux
endif
ifndef GOARCH
	GOARCH := amd64
endif

.PHONY: clean run-test get-deps deps update-deps fmt run-tests

all: get-deps $(BIN)

ci: clean all run-tests

clean:
	rm -rf $(BUILD_PATH) $(BIN) cli/templates_bindata.go

install: $(BIN)
	cp kocho /usr/local/bin/

get-deps: .gobuild .gobuild/bin/go-bindata

deps:
	@${MAKE} -B -s .gobuild/bin/go-bindata
	@${MAKE} -B -s .gobuild

.gobuild/bin/go-bindata:
	GOPATH=$(GOPATH) GOBIN=$(GOPATH)/bin go get github.com/jteeuwen/go-bindata/...

.gobuild:
	@mkdir -p $(GS_PATH)
	@rm -f $(GS_PATH)/$(PROJECT) && cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)
	#
	# Fetch private packages first (so `go get` skips them later)
	@GOPATH=$(GOPATH) builder go get github.com/aws/aws-sdk-go
	@GOPATH=$(GOPATH) builder go get github.com/go-ini/ini
	@GOPATH=$(GOPATH) builder go get github.com/jmespath/go-jmespath
	@GOPATH=$(GOPATH) builder go get github.com/crackcomm/cloudflare
	@GOPATH=$(GOPATH) builder go get github.com/juju/errgo
	@GOPATH=$(GOPATH) builder go get github.com/nlopes/slack
	@GOPATH=$(GOPATH) builder go get github.com/ryanuber/columnize
	@GOPATH=$(GOPATH) builder go get github.com/spf13/viper
	@GOPATH=$(GOPATH) builder go get github.com/spf13/pflag

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

$(BIN): $(SOURCE) VERSION cli/templates_bindata.go
	@echo Building for $(GOOS)/$(GOARCH)
	docker run \
		--rm \
		-v $(shell pwd):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-w /usr/code \
		golang:$(GOVERSION) \
		go build -a -ldflags \
		"-X github.com/giantswarm/kocho/cli.projectVersion=$(VERSION) -X github.com/giantswarm/kocho/cli.projectBuild=$(COMMIT)" \
		-o $(BIN)

cli/templates_bindata.go: .gobuild/bin/go-bindata $(TEMPLATES)
	.gobuild/bin/go-bindata -pkg cli -o cli/templates_bindata.go default-templates/

run-tests:
	GOPATH=$(GOPATH) go test ./... -cover

godoc: all
	@echo Opening godoc server at http://localhost:6060/pkg/github.com/$(ORGANIZATION)/$(PROJECT)/
	docker run \
	    --rm \
	    -v $(shell pwd):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOROOT=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -e GO15VENDOREXPERIMENT=1 \
	    -w /usr/code \
      -p 6060:6060 \
		golang:1.5 \
		godoc -http=:6060

fmt:
	gofmt -l -w .

bin-dist: all
	mkdir -p bin-dist/
	cp -f README.md bin-dist/
	cp -f LICENSE bin-dist/
	cp $(PROJECT) bin-dist/
	cd bin-dist/ && tar czf $(PROJECT).$(VERSION).tar.gz *
