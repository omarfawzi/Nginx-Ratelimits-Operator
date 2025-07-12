IMAGE ?= nginx-ratelimits-operator:latest

.PHONY: build lint fmt docker

build:
cd src && go build ./...

lint:
cd src && go vet ./...

fmt:
cd src && gofmt -w `find . -name '*.go'`

docker:
docker build -t $(IMAGE) src
