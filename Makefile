NAME			= loon
VERSION		= 1.0.0

GO			?= go
BIN			?= $(PWD)/build
LDFLAGS ?= -ldflags="-s -w -X main.Version=$(VERSION)"

install:
	$(GO) install -v .

build:
	@mkdir -p $(BIN)
	CGO_ENABLED=0 $(GO) build -o $(BIN)/$(NAME) -v $(LDFLAGS) .

test:
	$(GO) test -v ./...

lint:
	$(GO) vet .
