.DEFAULT_GOAL := help
.PHONY: help run build test test-cover lint reuse-lint clean setup-hooks check-licenses docker-build docker-scan docker-tag docker-push

VERSION    ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0-dev")
REGISTRY   ?= docker.io
NAMESPACE  ?= satriller
IMAGE      ?= epochgate
FULL_IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(IMAGE)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Start the server
	go run ./cmd/server

build: ## Build binary to bin/server
	go build -o bin/server ./cmd/server

test: ## Run all tests
	go test ./...

test-cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Run golangci-lint
	golangci-lint run ./...

reuse-lint: ## Check REUSE compliance
	reuse lint

setup-hooks: ## Install git hooks
	git config core.hooksPath .githooks

check-licenses: ## Check dependency licenses with Trivy
	.githooks/check-licenses

docker-build: ## Build Docker image
	docker build -t $(FULL_IMAGE):$(VERSION) -t $(FULL_IMAGE):latest .

docker-scan: docker-build ## Build + scan for HIGH/CRITICAL CVEs
	trivy image --severity HIGH,CRITICAL --exit-code 1 $(FULL_IMAGE):$(VERSION)

docker-tag: docker-build ## Show Docker image tags
	@echo "$(FULL_IMAGE):$(VERSION)"
	@echo "$(FULL_IMAGE):latest"

docker-push: docker-build ## Push to Docker Hub + update repo description
	docker push $(FULL_IMAGE):$(VERSION)
	docker push $(FULL_IMAGE):latest
	@echo "Updating Docker Hub description..."
	@TOKEN=$$(curl -s https://hub.docker.com/v2/auth/token \
		-H 'Content-Type: application/json' \
		-d "{\"identifier\":\"$(NAMESPACE)\",\"secret\":\"$(DOCKER_PASSWORD)\"}" \
		| jq -r '.access_token') && \
	README_JSON=$$(cat README.md | jq -Rs .) && \
	curl -s -X PATCH \
		"https://hub.docker.com/v2/repositories/$(NAMESPACE)/$(IMAGE)/" \
		-H "Authorization: Bearer $$TOKEN" \
		-H "Content-Type: application/json" \
		-d "{\"description\": \"Reverse proxy gatekeeper for npm - blocks young packages to prevent supply-chain attacks\", \"full_description\": $$README_JSON}" \
		| jq -r '.description // .message'

clean: ## Remove build artifacts
	rm -rf bin/
