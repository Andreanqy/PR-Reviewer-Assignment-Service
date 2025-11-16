BINARY_NAME=pr-server

PKG=.

GO=go

build:
	$(GO) build -o $(BINARY_NAME) $(PKG)

run:
	$(GO) run $(PKG)

build-run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

.PHONY: build run build-run clean
