all: deps build

GOOS ?= linux
GOARCH ?= amd64
APP_NAME = oomie

clean:
	rm -f ./bin/$(APP_NAME)

deps:
	go mod vendor

fmt:
	find . -path ./vendor -prune -o -name '*.go' -print | xargs -L 1 -I % gofmt -s -w %

build: clean fmt deps
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -mod vendor -o ./bin/$(APP_NAME)

.PHONY: all clean deps fmt build container
