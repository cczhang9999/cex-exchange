GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: api
api:
	export PATH=$$PATH:$$(go env GOPATH)/bin; \
	protoc --proto_path=. \
	       --go_out=paths=source_relative:. \
	       --go-grpc_out=paths=source_relative:. \
	       api/proto/exchange.proto

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: wire
wire:
	export PATH=$$PATH:$$(go env GOPATH)/bin; \
	wire ./...

.PHONY: all
all: api wire build
