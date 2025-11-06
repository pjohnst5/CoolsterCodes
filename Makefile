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
	@echo "Scanning images.."
	@temp_file=$$(mktemp); \
	find content/ -type f \( -iname "*.jpg" -o -iname "*.jpeg" -o -iname "*.png" \) -print0 2>/dev/null | while IFS= read -r -d '' img; do \
		if command -v identify >/dev/null 2>&1; then \
			dimensions=$$(identify -format "%wx%h" "$${img}[0]" 2>/dev/null | head -1); \
			if [ -n "$$dimensions" ]; then \
				width=$$(echo "$$dimensions" | cut -d'x' -f1); \
				height=$$(echo "$$dimensions" | cut -d'x' -f2); \
				if [ "$$width" -gt 1200 ] || [ "$$height" -gt 1200 ]; then \
					echo "❌ OVERSIZED: $$img ($$dimensions)"; \
					echo "OVERSIZED" >> "$$temp_file"; \
				fi; \
			else \
				echo "⚠️  SKIP: $$img (could not read dimensions)"; \
			fi; \
		else \
			echo "⚠️  ImageMagick 'identify' command not found. Install with:"; \
			echo "   macOS: brew install imagemagick"; \
			echo "   Ubuntu: sudo apt-get install imagemagick"; \
			rm -f "$$temp_file"; \
			exit 1; \
		fi; \
		echo "TOTAL" >> "$$temp_file"; \
	done; \
	if [ -f "$$temp_file" ]; then \
		total_images=$$(grep -c "TOTAL" "$$temp_file" 2>/dev/null); \
		oversized_found=$$(grep -c "OVERSIZED" "$$temp_file" 2>/dev/null); \
		[ -z "$$total_images" ] && total_images=0; \
		[ -z "$$oversized_found" ] && oversized_found=0; \
		rm -f "$$temp_file"; \
		echo "Total images scanned: $$total_images"; \
		echo "Oversized images: $$oversized_found"; \
		if [ "$$oversized_found" -gt 0 ]; then \
			echo "❌ FAILED: Found $$oversized_found oversized image(s)"; \
			echo "Run 'make build' to optimize images in-place"; \
			exit 1; \
		else \
			echo "✅ PASSED: All images are within size limits"; \
		fi; \
	else \
		echo "❌ ERROR: Could not create temporary file"; \
		exit 1; \
	fi
