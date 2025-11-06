.DEFAULT_GOAL := dev

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  dev          - Clean, install, and start development loop"
	@echo "  build        - Build the site once"
	@echo "  clean        - Clean the public/ directory"
	@echo "  install      - Install the binary"
	@echo "  loop         - Start the development server loop"
	@echo "  check        - Run tailwind, lint, and test"
	@echo "  test         - Run all tests"
	@echo "  test-mphoto  - Run image optimization module tests"
	@echo "  images       - Scan for oversized images (>1200x1200)"
	@echo "  lint         - Run linter"
	@echo "  vet          - Run go vet"
	@echo "  tailwind     - Build Tailwind CSS"
	@echo "  article      - Create a new article"

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

.PHONY: images
images:
	@echo "Scanning images for size violations..."
	@oversized=0; total=0; \
	for img in $$(find content/ -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) 2>/dev/null); do \
		if [ -f "$$img" ]; then \
			total=$$((total + 1)); \
			if command -v identify >/dev/null 2>&1; then \
				dimensions=$$(identify -format "%wx%h" "$$img" 2>/dev/null); \
				if [ -n "$$dimensions" ]; then \
					width=$${dimensions%x*}; height=$${dimensions#*x}; \
					if [ "$$width" -gt 1200 ] || [ "$$height" -gt 1200 ]; then \
						echo "❌ OVERSIZED: $$img ($$dimensions)"; \
						oversized=$$((oversized + 1)); \
					fi; \
				fi; \
			else \
				echo "⚠️  ImageMagick not found"; \
				exit 1; \
			fi; \
		fi; \
	done; \
	echo "Total: $$total images, Oversized: $$oversized"; \
	if [ "$$oversized" -gt 0 ]; then \
		echo "❌ FAILED: Run 'make build' to optimize"; \
		exit 1; \
	else \
		echo "✅ All images are properly sized"; \
	fi
