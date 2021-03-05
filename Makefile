APP_NAME := export-komoot
GO := go

REVISION := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n')
LDFLAGS := "-s -w -X $(PACKAGE).GitRevision=$(REVISION) -X $(PACKAGE).GitBranch=$(BRANCH)"

init:
	@go mod download

build: init
	@$(GO) build -v -ldflags $(LDFLAGS) -o $(APP_NAME)

test:
	@$(GO) test -cover `go list ./... | grep -v cmd`

run: build
	@DEBUG=0 ./$(APP_NAME)
