.DEFAULT_GOAL := dev

.PHONY: dev
dev: clean install loop

.PHONY: check
check: tailwind lint test

.PHONY: tailwind
tailwind:
	scripts/tailwind.sh

.PHONY: article
article:
	scripts/article.sh

.PHONY: all
all: clean install test vet lint build

.PHONY: build
build:
	$(shell go env GOPATH)/bin/coolstercodes build

.PHONY: clean
clean:
	mkdir -p public/
	rm -f -r public/*

.PHONY: install
install:
	go install .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: loop
loop:
	$(shell go env GOPATH)/bin/coolstercodes loop

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: vet
vet:
	go vet ./...
