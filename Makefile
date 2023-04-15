GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
LOCAL_BIN:=$(CURDIR)/bin
FILE_HASH := $(shell git rev-parse HEAD)
PROJECT_NAME:=retranslator

export PATH:=$(LOCAL_BIN):$(PATH)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps:
	$(info Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.8.0
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.8.0
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v0.6.7
	GOBIN=$(LOCAL_BIN) go install github.com/golang/mock/mockgen@v1.6.0
	GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@v1.9.0

install-lint: ## Installs golangci-lint tool which a go linter
ifndef GOLANGCI_LINT
	${info golangci-lint not found, installing golangci-lint@latest}
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif

lint: install-lint ## Runs linters
	@echo "-- linter running"
	golangci-lint run -c .golangci.yaml ./pkg...
	golangci-lint run -c .golangci.yaml ./cmd...

gogen: ## generate code
	${info generate code...}
	go generate ./internal...

test: ## Runs tests
	${info Running tests...}
	go test -v -race ./... -cover -coverprofile cover.out
	go tool cover -func cover.out | grep total

bench: ## Runs benchmarks
	${info Running benchmarks...}
	go test -bench=. -benchmem ./... -run=^#

proto: ## Generates protobuf files
	protoc -I api \
		--go_out=pkg/retranslator --go_opt=paths=source_relative \
 		--go-grpc_out=pkg/retranslator --go-grpc_opt=paths=source_relative \
		api/*.proto
	make gogen

build: ## Build binary
	${info Building binary...}
	go build -ldflags="-X 'main.dbPath=/var/lib/retranslator/storage.db'" -o ./bin/retranslator ./cmd

stop: ## Stops the local environment
	${info Stopping containers...}
	docker container ls -q --filter name=${PROJECT_NAME} ; true
	${info Dropping containers...}
	docker rm -f -v $(shell docker container ls -q --filter name=${PROJECT_NAME}) ; true

run_docker: stop ## Run docker with binary
	${info Run app containered}
	GIT_HASH=${FILE_HASH} docker compose -p ${PROJECT_NAME} up --build -d

run: build ## Run binary
	@for i in {1..100}; do \
		echo "Iteration $$i"; \
		./bin/retranslator; \
		sleep 60; \
	done

logs_d: ## Show container logs
	${info Show logs...}
	docker logs -f retranslator

.PHONY: help install-lint test gogen lint stop dev_up build run run_docker logs_d stop
.DEFAULT_GOAL := help