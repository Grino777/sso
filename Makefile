PROJECT_DIR = $(shell pwd)
PROJECT_BIN = $(PROJECT_DIR)/bin
$(shell [ -f bin ] || mkdir -p $(PROJECT_BIN))
PATH := $(PROJECT_BIN):$(PATH)

GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint

.PHONY: .install-linter
.install-linter:
	### INSTALL GOLANGCI-LINT ###
	[ -f $(PROJECT_BIN)/golangci-lint ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v2.1.6

.PHONY: lint
lint: .install-linter
	### RUN GOLANGCI-LINT ###
	$(GOLANGCI_LINT) run ./... --config=./.golangci.yml

.PHONY: lint-to-txt
lint-to-txt: .install-linter
	### RUN GOLANGCI-LINT TO TXT ###
	$(GOLANGCI_LINT) run ./... --config=./.golangci.yml > ./lint-log.log

.PHONY: lint-fast
lint-fast: .install-linter
	### RUN GOLANGCI-LINT FAST ###
	$(GOLANGCI_LINT) run ./... --fast --config=./.golangci.yml
