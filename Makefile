BIN      := kube-janitor
IMG      ?= ghcr.io/niksecops-crypto/kube-janitor
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: build test lint docker push clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BIN) ./cmd/janitor

test:
	go test -v -race -coverprofile=coverage.out ./...

test-short:
	go test -short -race ./...

lint:
	golangci-lint run ./...

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMG):$(VERSION) -t $(IMG):latest .

push: docker
	docker push $(IMG):$(VERSION)
	docker push $(IMG):latest

coverage:
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ coverage.out coverage.html

.DEFAULT_GOAL := build
