include .env
export

APP_NAME := export-komoot
REVISION := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n')
BUILD_CMD := go build -buildvcs -ldflags "-s -w -X $(PACKAGE).GitRevision=$(REVISION) -X $(PACKAGE).GitBranch=$(BRANCH)" -trimpath

init:
	@go mod download

build: init
	$(call build-binary)

test:
	@$(GO) test -cover `go list ./... | grep -v cmd`

run-incremental: build
	@DEBUG=0 ./$(APP_NAME) --email "$(KOMOOT_EMAIL)" --password "$(KOMOOT_PASSWD)" --userid "$(KOMOOT_USER_ID)" --to "export"

run-full: build
	@DEBUG=0 ./$(APP_NAME) --email "$(KOMOOT_EMAIL)" --password "$(KOMOOT_PASSWD)" --userid "$(KOMOOT_USER_ID)" --to "export" --fulldownload

run-filter: build
	@DEBUG=0 ./$(APP_NAME) --email "$(KOMOOT_EMAIL)" --password "$(KOMOOT_PASSWD)" --userid "$(KOMOOT_USER_ID)" --to "export" --fulldownload --filter "*KK*"

help: build
	@DEBUG=0 ./$(APP_NAME) --help

build-all:
	@rm -f ./$(APP_NAME)-*
	$(call build-binary-mac)
	$(call build-binary-linux)
	$(call build-binary-windows)

## Helper functions
define build-binary
	@GOOS=$(1) GOARCH=$(2) $(BUILD_CMD) -o $(APP_NAME)
endef

define build-binary-mac
	@echo "Building $(APP_NAME) for macos"
	$(call build-binary,darwin,arm64,arm64)
	@mv $(APP_NAME) $(APP_NAME)-arm
	$(call build-binary,darwin,amd64,x86_64)
	@mv $(APP_NAME) $(APP_NAME)-x86
	@lipo -create -output $(APP_NAME) $(APP_NAME)-x86 $(APP_NAME)-arm
	@tar czf ./$(APP_NAME)-$(REVISION)-macos.tar.gz $(APP_NAME)
	@rm -f $(APP_NAME)-x86 $(APP_NAME)-arm $(APP_NAME)
endef

define build-binary-linux
	@echo "Building $(APP_NAME) for linux"
	$(call build-binary,linux,arm64,arm64)
	@tar czf ./$(APP_NAME)-$(REVISION)-linux-arm64.tar.gz $(APP_NAME)
	$(call build-binary,linux,amd64,x86_64)
	@tar czf ./$(APP_NAME)-$(REVISION)-linux-x86_64.tar.gz $(APP_NAME)
	@rm -f $(APP_NAME)
endef

define build-binary-windows
	@echo "Building $(APP_NAME) for windows"
	@$(call build-binary,windows,arm64,arm64)
	@mv ./$(APP_NAME) ./$(APP_NAME).exe
	@zip -rq ./$(APP_NAME)-$(REVISION)-windows-arm64.zip $(APP_NAME).exe
	@$(call build-binary,windows,amd64,x86_64)
	@mv ./$(APP_NAME) ./$(APP_NAME).exe
	@zip -rq ./$(APP_NAME)-$(REVISION)-windows-x86_64.zip $(APP_NAME).exe
	@rm -f $(APP_NAME).exe $(APP_NAME)
endef
