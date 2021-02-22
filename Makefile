all: deps build

ENVVAR = GOOS=linux GOARCH=amd64
TAG = v0.0.2
APP_NAME = oomie

clean:
	rm -f ./bin/$(APP_NAME)

deps:
	go mod vendor

fmt:
	find . -path ./vendor -prune -o -name '*.go' -print | xargs -L 1 -I % gofmt -s -w %

build: clean fmt
	$(ENVVAR) CGO_ENABLED=0 go build -mod vendor -o ./bin/$(APP_NAME)

container:
	docker build -f Dockerfile -t $(APP_NAME):latest .
	docker tag $(APP_NAME):latest ${APP_NAME}:$(TAG)

.PHONY: all clean deps fmt build container
