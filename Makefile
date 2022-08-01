IMAGE_REPO ?= ghcr.io/lodrem/echoserver
IMAGE_TAG ?= v0.0.1

.PHONY: build
build:
	go build -o bin/echosever

.PHONY: image
image:
	docker buildx build ../ \
		--output=type=docker \
		--no-cache \
		--force-rm \
		--tag $(IMAGE_REPO):$(IMAGE_TAG) \
		--file Dockerfile