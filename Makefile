.PHONY: run build test test-cover lint reuse-lint clean setup-hooks check-licenses docker-build docker-scan docker-tag

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0-dev")
IMAGE   ?= epochgate

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
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-scan: docker-build
	trivy image --severity HIGH,CRITICAL --exit-code 1 $(IMAGE):$(VERSION)

docker-tag: docker-build
	@echo "$(IMAGE):$(VERSION)"
	@echo "$(IMAGE):latest"

clean:
	rm -rf bin/
