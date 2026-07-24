# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gqrl qrvm all test lint fmt clean devtools \
	network-start live-test network-stop help

GOBIN = ./build/bin
GO ?= latest
GORUN = go run
E2E_DIR = scripts/testing/e2e
E2E_RUNNER = go -C $(E2E_DIR) run ./cmd/e2e
E2E_GINKGO = go -C $(E2E_DIR) tool ginkgo
E2E_SUITES ?=
E2E_NETWORK_DIR ?= /tmp/go-qrl-e2e-network
E2E_NETWORK_DIR_ABS = $(abspath $(E2E_NETWORK_DIR))
E2E_TIMEOUT ?= 25m
empty :=
space := $(empty) $(empty)
comma := ,
E2E_SUITE_LIST = $(strip $(subst $(comma),$(space),$(E2E_SUITES)))
E2E_SUITE_PACKAGES = $(addprefix ./suites/,$(E2E_SUITE_LIST))
E2E_SUITE_LABELS = $(subst $(space), || ,$(E2E_SUITE_LIST))
E2E_DOCKER_BIN ?= docker

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

#? network-start: Start a standalone E2E test network without running suites.
network-start:
	$(E2E_RUNNER) start \
		--repo-root "$(CURDIR)" \
		--network-dir "$(E2E_NETWORK_DIR_ABS)" \
		--docker-bin "$(E2E_DOCKER_BIN)"

#? live-test: Run selected Ginkgo E2E suites against the already-running network.
live-test:
	@test -n "$(E2E_SUITE_LIST)" || { echo "E2E_SUITES must name at least one suite" >&2; exit 2; }
	E2E_SUITES="$(E2E_SUITES)" \
	E2E_NETWORK_DIR="$(E2E_NETWORK_DIR_ABS)" \
	E2E_REPO_ROOT="$(CURDIR)" \
	$(E2E_GINKGO) \
		--tags=e2e \
		--ldflags='-s -w' \
		--procs=1 \
		--require-suite \
		--fail-on-empty \
		--fail-on-pending \
		--label-filter='e2e && live && ($(E2E_SUITE_LABELS))' \
		--timeout="$(E2E_TIMEOUT)" \
		--poll-progress-after=30s \
		--poll-progress-interval=30s \
		$(E2E_SUITE_PACKAGES) \
		-- -test.run='^TestE2E$$'

#? network-stop: Stop only the exact E2E network recorded in E2E_NETWORK_DIR.
network-stop:
	$(E2E_RUNNER) stop --network-dir "$(E2E_NETWORK_DIR_ABS)"

#? help: Get more info on make commands.
help: Makefile
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'
