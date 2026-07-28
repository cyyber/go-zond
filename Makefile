# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gqrl qrvm all test lint fmt clean devtools \
	network-start network-stop e2e-test help

GOBIN = ./build/bin
GORUN = go run
DEVNET_GO = go -C testing/devnet
# Enclave name and start timeout default inside the devnet CLI; set these only
# to override.
DEVNET_ENCLAVE_NAME ?=
DEVNET_START_TIMEOUT ?=
DEVNET_EXECUTION_IMAGE ?= local/go-qrl:devnet
override DEVNET_PARAMS_FILE := $(if $(strip $(DEVNET_PARAMS_FILE)),$(abspath $(DEVNET_PARAMS_FILE)))
E2E_SUITE_TIMEOUT ?= 25m

#? gqrl: Build gqrl.
gqrl:
	$(GORUN) build/ci.go install ./cmd/gqrl
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gqrl\" to launch gqrl."

#? qrvm: Build qrvm.
qrvm:
	$(GORUN) build/ci.go install ./cmd/qrvm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/qrvm\" to launch qrvm."

#? all: Build all packages and executables.
all:
	$(GORUN) build/ci.go install

#? test: Run the tests.
test: all
	$(GORUN) build/ci.go test

#? lint: Run certain pre-selected linters.
lint: ## Run linters.
	$(GORUN) build/ci.go lint

#? fmt: Ensure consistent code formatting.
fmt:
	gofmt -s -w $(shell find . -name "*.go")

#? clean: Clean go cache, built executables, and the auto generated folder.
clean:
	go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

#? devtools: Install recommended developer tools.
devtools:
	env GOBIN= go install golang.org/x/tools/cmd/stringer@latest
	env GOBIN= go install github.com/fjl/gencodec@latest
	env GOBIN= go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	env GOBIN= go install ./cmd/abigen
	@type "hypc" 2> /dev/null || echo 'Please install hypc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

#? network-start: Start a standalone development network without running suites.
network-start:
	@docker info >/dev/null 2>&1 || { \
		echo "Docker is required and its daemon must be running" >&2; \
		exit 1; \
	}
	@kurtosis version 2>/dev/null | grep -Eq '^CLI Version:[[:space:]]+1\.20\.' || { \
		echo "Kurtosis CLI 1.20.x is required (https://docs.kurtosis.com/upgrade)" >&2; \
		exit 1; \
	}
	kurtosis engine start
	docker build --tag "$(DEVNET_EXECUTION_IMAGE)" .
	$(DEVNET_GO) run ./cmd/devnet start \
		--execution-image "$(DEVNET_EXECUTION_IMAGE)" \
		$(if $(DEVNET_ENCLAVE_NAME),--enclave-name "$(DEVNET_ENCLAVE_NAME)") \
		$(if $(DEVNET_START_TIMEOUT),--timeout "$(DEVNET_START_TIMEOUT)") \
		$(if $(DEVNET_PARAMS_FILE),--params-file "$(DEVNET_PARAMS_FILE)")

#? network-stop: Stop the development network.
network-stop:
	$(DEVNET_GO) run ./cmd/devnet stop $(if $(DEVNET_ENCLAVE_NAME),--enclave-name "$(DEVNET_ENCLAVE_NAME)")

#? e2e-test: Run selected Ginkgo E2E suites against the already-running network.
e2e-test:
	@test -n "$(strip $(E2E_PACKAGES))" || { echo "E2E_PACKAGES must name at least one suite package" >&2; exit 2; }
	DEVNET_ENCLAVE_NAME="$(DEVNET_ENCLAVE_NAME)" \
	$(DEVNET_GO) tool ginkgo \
		--tags=e2e \
		--procs=1 \
		--require-suite \
		--fail-on-empty \
		--fail-on-pending \
		--timeout="$(E2E_SUITE_TIMEOUT)" \
		$(strip $(E2E_PACKAGES)) \
		-- -test.run='^TestE2E$$'

#? help: Get more info on make commands.
help: Makefile
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
