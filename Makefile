.PHONY: build

GOBIN := $(shell go env GOPATH)/bin

build:
	go build -o clocker cmd/clocker/main.go \
	&& cp -r clocker $(GOBIN)/clocker