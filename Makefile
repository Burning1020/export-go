#
# Copyright (c) 2018 Cavium
#
# SPDX-License-Identifier: Apache-2.0
#


.PHONY: build clean test docker run


GO=CGO_ENABLED=0 go
GO_ARM64=CGO_ENABLED=0 GOARCH=arm64 go
GOCGO=CGO_ENABLED=1 go
GOCGO_ARM64=CGO_ENABLED=1 GOARCH=arm64 go
#CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++

DOCKERS=docker_export_client docker_export_distro
DOCKERS-arm64=docker_export_client-arm64 docker_export_distro-arm64

.PHONY: $(DOCKERS) $(DOCKERS-arm64)

MICROSERVICES=cmd/export-client/export-client cmd/export-distro/export-distro

.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/export-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	go build ./...

cmd/export-client/export-client:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/export-client

cmd/export-client/export-client-arm64:
	$(GO_ARM64) build $(GOFLAGS) -o $@ ./cmd/export-client

cmd/export-distro/export-distro:
	$(GOCGO) build $(GOFLAGS) -o $@ ./cmd/export-distro

cmd/export-distro/export-distro-arm64:
	$(GOCGO_ARM64) build $(GOFLAGS) -o $@ ./cmd/export-distro

clean:
	rm -f $(MICROSERVICES)

test:
	go test -cover ./...
	go vet ./...

prepare:
	glide install

run:
	cd bin && ./edgex-launch.sh

run_docker:
	cd bin && ./edgex-docker-launch.sh

docker: $(DOCKERS)

docker_export_client:
	docker build \
		-f docker/Dockerfile.export-client \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-export-client-go:$(GIT_SHA) \
		-t edgexfoundry/docker-export-client-go:$(VERSION)-dev \
		.

docker_export_distro:
	docker build \
		-f docker/Dockerfile.export-distro \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-export-distro-go:$(GIT_SHA) \
		-t edgexfoundry/docker-export-distro-go:$(VERSION)-dev \
		.

docker-arm64: $(DOCKERS-arm64)

docker_export_client-arm64:
	docker build \
		-f docker/Dockerfile.export-client_ARM64 \
		--label "git_sha=$(GIT_SHA)" \
		-t burning/docker-export-client-go-arm64:$(GIT_SHA) \
		-t burning/docker-export-client-go-arm64:$(VERSION)-dev \
		.

docker_export_distro-arm64:
	docker build \
		-f docker/Dockerfile.export-distro_ARM64 \
		--label "git_sha=$(GIT_SHA)" \
		-t burning/docker-export-distro-go-arm64:$(GIT_SHA) \
		-t burning/docker-export-distro-go-arm64:$(VERSION)-dev \
		.
