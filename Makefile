.PHONY: build vet test cover tidy clean

override GOCACHE := $(CURDIR)/.cache/go-build
override GOPATH := $(CURDIR)/.cache/go
GO ?= go

GO_ENV = GOCACHE=$(GOCACHE) GOPATH=$(GOPATH)

build:
	@mkdir -p $(GOCACHE) $(GOPATH)
	@$(GO_ENV) $(GO) build ./...

vet:
	@mkdir -p $(GOCACHE) $(GOPATH)
	@$(GO_ENV) $(GO) vet ./...

test:
	@mkdir -p $(GOCACHE) $(GOPATH)
	@$(GO_ENV) $(GO) test ./... -count=1 -timeout 120s

cover:
	@mkdir -p $(GOCACHE) $(GOPATH)
	@$(GO_ENV) $(GO) test -cover ./...

tidy:
	@mkdir -p $(GOCACHE) $(GOPATH)
	@$(GO_ENV) $(GO) mod tidy

clean:
	@rm -rf $(CURDIR)/.cache $(CURDIR)/.gocache $(CURDIR)/.gomodcache $(CURDIR)/.gopath
