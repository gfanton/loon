GO=go

install:
	$(GO) install .

test:
	$(GO) test -v ./...
