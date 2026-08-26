.PHONY: all build install uninstall clean app test

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin

all: build

build:
	@mkdir -p bin
	@echo "Building sqldoc..."
	go build -o bin/sqldoc ./cmd/sqldoc
	@echo "Building sqldoc-serve..."
	go build -o bin/sqldoc-serve ./cmd/sqldoc-serve
	@echo "Building sqldoc-viewer..."
	CGO_ENABLED=1 go build -o bin/sqldoc-viewer ./cmd/sqldoc-viewer
	@echo "All binaries built in bin/"

install:
	@./install.sh

install-viewer:
	@./install.sh --viewer-only

app:
	@./install.sh

uninstall:
	@./install.sh --uninstall

clean:
	@rm -rf bin/ /tmp/sqldoc_* 2>/dev/null
	@echo "Cleaned build artifacts"

test:
	./bin/sqldoc info examples/demo.db
	./bin/sqldoc render examples/demo.db -o examples/output.html
