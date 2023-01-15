IMAGE_REPO ?= ghcr.io/lodrem/echoserver
IMAGE_TAG ?= v0.0.1

.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/echoserver

.PHONY: unit-test
unit-test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint: golangci-lint
	$(LINTER) run -v

.PHONY: image
image:
	docker buildx build ./ \
		--output=type=docker \
		--no-cache \
		--force-rm \
		--tag $(IMAGE_REPO):$(IMAGE_TAG) \
		--file Dockerfile

LINTER = $(shell pwd)/bin/golangci-lint
LINTER_VERSION = v1.50.1
.PHONY: golangci-lint
golangci-lint:  ## Download golangci-lint locally if necessary.
	@echo "Installing golangci-lint"
	@test -s $(LINTER) || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(LINTER_VERSION)/install.sh | sh -s -- -b $(shell pwd)/bin $(LINTER_VERSION)
