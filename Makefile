.PHONY: build clean tidy test

BINARY_NAME := hostcontrol-mcp
DIST_DIR := dist

build:
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/hostcontrol-mcp

clean:
	rm -rf $(DIST_DIR)

tidy:
	go mod tidy

test:
	go test -v ./...
