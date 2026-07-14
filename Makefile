.PHONY: run build test test-cover lint reuse-lint clean setup-hooks check-licenses docker-build docker-scan docker-tag

VERSION    ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0-dev")
REGISTRY   ?= docker.io
NAMESPACE  ?= satrill
IMAGE      ?= epochgate
FULL_IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(IMAGE)

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

reuse-lint:
	reuse lint

setup-hooks:
	git config core.hooksPath .githooks

check-licenses:
	.githooks/check-licenses

docker-build:
	docker build -t $(FULL_IMAGE):$(VERSION) -t $(FULL_IMAGE):latest .

docker-scan: docker-build
	trivy image --severity HIGH,CRITICAL --exit-code 1 $(FULL_IMAGE):$(VERSION)

docker-tag: docker-build
	@echo "$(FULL_IMAGE):$(VERSION)"
	@echo "$(FULL_IMAGE):latest"

clean:
	rm -rf bin/
