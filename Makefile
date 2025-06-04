.DEFAULT_GOAL := mine

.PHONY: mine
mine: clean install loop

.PHONY: check
check: tailwind lint test

.PHONY: tailwind
tailwind:
	scripts/tailwind.sh

.PHONY: all
all: clean install test vet lint check-dl0 check-gofmt check-headers build

.PHONY: build
build:
	$(shell go env GOPATH)/bin/sorg build

.PHONY: check-gofmt
check-gofmt:
	scripts/check_gofmt.sh

.PHONY: check-headers
check-headers:
	scripts/check_headers.sh

.PHONY: clean
clean:
	mkdir -p public/
	rm -f -r public/*

.PHONY: compile
compile: install

# Long TTL (in seconds) to set on an object in S3. This is suitable for items
# that we expect to only have to invalidate very rarely like images. Although
# we set it for all assets, those that are expected to change more frequently
# like script or stylesheet files are versioned by a path that can be set at
# build time.
LONG_TTL := 86400

# Short TTL (in seconds) to set on an object in S3. This is suitable for items
# that are expected to change more frequently like any HTML file.
SHORT_TTL := 3600

.PHONY: install
install:
	go install .

.PHONY: killall
killall:
	killall sorg

.PHONY: lint
lint:
	golangci-lint run

.PHONY: loop
loop:
	$(shell go env GOPATH)/bin/sorg loop

.PHONY: sigusr2
sigusr2:
	killall -SIGUSR2 sorg

# sigusr2 aliases
.PHONY: reboot
reboot: sigusr2
.PHONY: restart
restart: sigusr2

.PHONY: test
test:
	go test ./...

.PHONY: test-nocache
test-nocache:
	go test -count=1 ./...

.PHONY: vet
vet:
	go vet ./...

# This is designed to be compromise between being explicit and readability. We
# can allow the find to discover everything in vendor/, but then the fswatch
# invocation becomes a huge unreadable wall of text that gets dumped into the
# shell. Instead, find all our own *.go files and then just tack the vendor/
# directory on separately (fswatch will watch it recursively).
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./vendor/*")

# Meant to be used in conjuction with `forego start`. When a Go file changes,
# this watch recompiles the project, then sends USR2 to the process which
# prompts Modulir to re-exec it.
.PHONY: watch-go
watch-go:
	fswatch -o $(GO_FILES) vendor/ | xargs -n1 -I{} make install sigusr2

#
# Helpers
#

.PHONY: check-target-dir
check-target-dir:
ifndef TARGET_DIR
	$(error TARGET_DIR is required)
endif
