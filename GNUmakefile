HOSTNAME = registry.terraform.io
NAMESPACE = ECCOShoes
NAME = utils
BINARY = terraform-provider-$(NAME)
VERSION = 1.0.0
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)

default: fmt lint build

build:
	go build -v ./...

vet:
	go vet ./...

install:
	go build -o $(BINARY)
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	mv $(BINARY) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build vet install generate
