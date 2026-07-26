# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gqrl qrvm all test lint fmt clean devtools \
	e2e-unit network-image network-start network-status live-test network-stop help

GOBIN = ./build/bin
GORUN = go run
E2E_GO = go -C testing/endtoend
override E2E_NETWORK_DIR := $(abspath $(or $(strip $(E2E_NETWORK_DIR)),/tmp/go-qrl-e2e-network))
E2E_NETWORK_TIMEOUT ?= 150m
E2E_SUITE_TIMEOUT ?= 25m
E2E_EXECUTION_IMAGE ?= local/go-qrl:e2e

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

#? e2e-unit: Run unit tests and vet for the isolated E2E module.
e2e-unit:
	$(E2E_GO) test -count=1 ./...
	$(E2E_GO) vet ./...

#? network-image: Build the go-qrl execution image used by the E2E network.
network-image:
	@status="$$(git status --porcelain=v1 --untracked-files=all)" || exit 1; \
	revision="$$(git rev-parse HEAD)" || exit 1; \
	if [ -n "$$status" ]; then \
		if [ "$${E2E_REQUIRE_CLEAN:-0}" = "1" ]; then \
			echo "E2E_REQUIRE_CLEAN=1 refuses a dirty execution-image build." >&2; \
			echo "$$status" >&2; \
			exit 1; \
		fi; \
		revision="working-tree-$$(git rev-parse --short=12 HEAD)"; \
		echo "Building dirty execution image with revision $$revision." >&2; \
	fi; \
	docker build \
		--build-arg "COMMIT=$$revision" \
		--tag "$(E2E_EXECUTION_IMAGE)" \
		.

#? network-start: Start a standalone E2E test network without running suites.
network-start: network-image
	$(E2E_GO) run ./cmd/e2e start \
		--network-dir "$(E2E_NETWORK_DIR)" \
		--execution-image "$(E2E_EXECUTION_IMAGE)" \
		--timeout "$(E2E_NETWORK_TIMEOUT)"

#? network-status: Check whether the standalone E2E network is ready.
network-status:
	$(E2E_GO) run ./cmd/e2e status --network-dir "$(E2E_NETWORK_DIR)"

#? live-test: Run selected Ginkgo E2E suites against the already-running network.
live-test:
	@test -n "$(strip $(E2E_PACKAGES))" || { echo "E2E_PACKAGES must name at least one suite package" >&2; exit 2; }
	E2E_NETWORK_DIR="$(E2E_NETWORK_DIR)" \
	$(E2E_GO) tool ginkgo \
		--tags=e2e \
		--procs=1 \
		--require-suite \
		--fail-on-empty \
		--fail-on-pending \
		--timeout="$(E2E_SUITE_TIMEOUT)" \
		--poll-progress-after=30s \
		--poll-progress-interval=30s \
		$(strip $(E2E_PACKAGES)) \
		-- -test.run='^TestE2E$$'

#? network-stop: Stop the deterministic E2E network slot for E2E_NETWORK_DIR.
network-stop:
	$(E2E_GO) run ./cmd/e2e stop --network-dir "$(E2E_NETWORK_DIR)"

#? help: Get more info on make commands.
help: Makefile
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
