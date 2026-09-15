HOSTNAME    = registry.terraform.io
NAMESPACE   = Sonny00
NAME        = thehive
BINARY      = terraform-provider-$(NAME)
VERSION     = 0.1.0

# Auto-detect OS/arch for local install.
OS          ?= $(shell go env GOOS)
ARCH        ?= $(shell go env GOARCH)
PLUGIN_DIR  = $(HOME)/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS)_$(ARCH)

default: build

.PHONY: build install test testacc lint generate clean

## build: compile the provider binary
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)

## install: build and install into the local Terraform plugin cache
install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY) $(PLUGIN_DIR)/

## test: run unit tests
test:
	go test ./... -v -count=1

## testacc: run acceptance tests (requires a live TheHive instance)
testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

## lint: static analysis
lint:
	golangci-lint run ./...

## generate: regenerate docs (requires tfplugindocs)
generate:
	go generate ./...
	tfplugindocs generate

## clean: remove build artefacts
clean:
	rm -f $(BINARY)
