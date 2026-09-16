#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell git describe --tags --long --always --dirty 2> /dev/null || echo v0.0.0-dev)
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
APP_DIR = ./app
MOCKS_DIR = $(CURDIR)/tests/mocks
HTTPS_GIT := https://github.com/fetchai/fetchd.git
DOCKER_BUF := docker run -v $(shell pwd):/workspace --workdir /workspace bufbuild/buf
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)

export GO111MODULE = on
export GOPRIVATE=github.com/fetchai/priv_wasmd_sec,github.com/fetchai/priv_wasmvm_sec

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

empty :=
space := $(empty) $(empty)
comma := ,
build_tags_comma_sep := $(subst $(space),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=fetch \
		  -X github.com/cosmos/cosmos-sdk/version.ServerName=fetchd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif

# PIE is enabled for all builds.
#
# PIE and static linking are *independent* properties:
#   -buildmode=pie     -> position-independent executable (platform *independent*)
#   -static-pie        -> *Linux* only *static* linking of the PIE executable
buildmode_flags += -buildmode=pie

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags $(build_tags_comma_sep) -ldflags '$(ldflags)' -trimpath $(buildmode_flags)

# The below include contains the tools target.
#include contrib/devtools/Makefile

all: install test

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/fetchd.exe ./cmd/fetchd
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/fetchd ./cmd/fetchd
endif

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-contract-tests-hooks:
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests.exe ./cmd/contract_tests
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/contract_tests ./cmd/contract_tests
endif

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/fetchd

# Static (musl) builds: link fetchd against the static wasmvm archive
# (libwasmvm_muslc.<arch>.a) instead of the dynamic libwasmvm.<arch>.so, so the
# resulting binary is fully static and the .so libraries do not need to be
# distributed/shipped alongside it.
#
# Requirements:
#   - the build must run on a musl-based system (e.g. Alpine Linux) with
#     gcc, make and xz available; the static-pie link mode used below is
#     supported by musl but NOT by glibc
#   - write access to /lib (the extracted archive must live in a default
#     linker search path, because the wasmvm cgo bindings link it via
#     -lwasmvm_muslc.<arch>; in Docker/container builds this is a given)

UNAME_M               := $(shell uname -m)
STATIC_WASMVM_MODULE  := github.com/CosmWasm/wasmvm/v3
STATIC_WASMVM_VERSION := $(shell go list -mod=readonly -m -f '{{.Version}}' $(STATIC_WASMVM_MODULE))
STATIC_WASMVM_LIB     := /lib/libwasmvm_muslc.$(UNAME_M).a
STATIC_BUILD_TAGS     := muslc
STATIC_LDFLAGS        := -linkmode=external -extldflags=-static

# Extract the static (musl) wasmvm archive shipped inside the module that
# satisfies github.com/CosmWasm/wasmvm/v3 (following any `replace` directives
# in go.mod) into /lib, so that the external linker (gcc) finds it when the
# `muslc` build tag makes the wasmvm cgo bindings link it
# (-lwasmvm_muslc.<arch>).
#
# The module dir is resolved *inside the recipe shell* (not via a Make
# $(shell) expansion) so that the resolution provably happens after the
# `go mod download` that precedes it. GOPRIVATE is cleared for the go
# commands so the file:// GOPROXY cache is used for the private module
# instead of an authenticated VCS fetch.
.PHONY: static-wasmvm-lib
static-wasmvm-lib: go.sum
	@echo "--> Extracting static wasmvm library ($(STATIC_WASMVM_VERSION)) to $(STATIC_WASMVM_LIB)"
	@dir=$$(GOPRIVATE= go list -mod=readonly -m -f '{{.Dir}}' $(STATIC_WASMVM_MODULE)); \
	  if [ -z "$$dir" ]; then \
	    GOPRIVATE= go mod download $(STATIC_WASMVM_MODULE); \
	    dir=$$(GOPRIVATE= go list -mod=readonly -m -f '{{.Dir}}' $(STATIC_WASMVM_MODULE)); \
	  fi; \
	  if [ -z "$$dir" ]; then \
	    echo "ERROR: could not resolve module dir for $(STATIC_WASMVM_MODULE)" >&2; \
	    exit 1; \
	  fi; \
	  unxz -c "$$dir/internal/api/libwasmvm_muslc.$(UNAME_M).a.xz" > "$(STATIC_WASMVM_LIB)"
	@chmod 644 "$(STATIC_WASMVM_LIB)"

# Generic static build: no wasmvm archive checksum pinning. Use this for
# non-release builds (e.g. local Alpine builds, CI builds from master). For
# release builds use the version-pinned target below instead.
install-static: static-wasmvm-lib
	$(MAKE) install \
	  LEDGER_ENABLED=$(LEDGER_ENABLED) \
	  BUILD_TAGS=$(STATIC_BUILD_TAGS) \
	  LDFLAGS="$(STATIC_LDFLAGS)"

build-static: static-wasmvm-lib
	$(MAKE) build \
	  LEDGER_ENABLED=$(LEDGER_ENABLED) \
	  BUILD_TAGS=$(STATIC_BUILD_TAGS) \
	  LDFLAGS="$(STATIC_LDFLAGS)"

# Release-pinned static build for v0.15.1: asserts the exact wasmvm version
# resolved from go.mod and the SHA256 of the *decompressed* static archive
# before building. The checksums are for github.com/fetchai/priv_wasmvm_sec/v3
# v3.0.8-rc.2 (the go.mod replace target of github.com/CosmWasm/wasmvm/v3);
# update them whenever the wasmvm pin changes.
verify-static-wasmvm-lib-v0.15.1: WASMVM_VERSION := v3.0.8-rc.2
verify-static-wasmvm-lib-v0.15.1: WASMVM_SHA256_x86_64 := 6863af60cebf04d094bc3bcf22a2777e1e1b4f1295d54e3b9a1082a3b359de8a
verify-static-wasmvm-lib-v0.15.1: WASMVM_SHA256_aarch64 := 46f4d0913331096f2926f28d5d0f4405eb700f571be229cf8774d942619370a4
verify-static-wasmvm-lib-v0.15.1: static-wasmvm-lib
	@echo "--> Verifying wasmvm version pin for fetchd v0.15.1"
	@if [ "$(WASMVM_VERSION)" != "$(STATIC_WASMVM_VERSION)" ]; then \
	  echo "ERROR: wasmvm module version is '$(STATIC_WASMVM_VERSION)', but fetchd" >&2; \
	  echo "       v0.15.1 is pinned to '$(WASMVM_VERSION)'." >&2; \
	  exit 1; \
	fi
	@echo "--> Verifying static wasmvm library checksum"
	@if [ "$$(sha256sum $(STATIC_WASMVM_LIB) | cut -d' ' -f1)" != "$(WASMVM_SHA256_$(UNAME_M))" ]; then \
	  echo "ERROR: SHA256 mismatch for $(STATIC_WASMVM_LIB)" >&2; \
	  echo "       expected: $(WASMVM_SHA256_$(UNAME_M))" >&2; \
	  echo "       actual:   $$(sha256sum $(STATIC_WASMVM_LIB) | cut -d' ' -f1)" >&2; \
	  exit 1; \
	fi
	@echo "    OK"

install-static-v0.15.1: verify-static-wasmvm-lib-v0.15.1 install-static

.PHONY: install-static build-static verify-static-wasmvm-lib-v0.15.1 install-static-v0.15.1

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/fetchd -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit
test-all: test-unit test-ledger-mock test-race test-cover

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
UNIT_TEST_ARGS		= cgo ledger test_ledger_mock norace
AMINO_TEST_ARGS		= ledger test_ledger_mock test_amino norace
LEDGER_TEST_ARGS	= cgo ledger norace
LEDGER_MOCK_ARGS	= ledger test_ledger_mock norace
TEST_RACE_ARGS		= cgo ledger test_ledger_mock
ifeq ($(EXPERIMENTAL),true)
	UNIT_TEST_ARGS		+= experimental
	AMINO_TEST_ARGS		+= experimental
	LEDGER_TEST_ARGS	+= experimental
	LEDGER_MOCK_ARGS	+= experimental
	TEST_RACE_ARGS		+= experimental
endif

test-unit: ARGS=-count=1 -tags='$(UNIT_TEST_ARGS)'
test-unit-amino: ARGS=-tags='${AMINO_TEST_ARGS}'
test-ledger: ARGS=-tags='${LEDGER_TEST_ARGS}'
test-ledger-mock: ARGS=-tags='${LEDGER_MOCK_ARGS}'
test-race: ARGS=-race -tags='${TEST_RACE_ARGS}'
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)

$(TEST_TARGETS): run-tests

SUB_MODULES = $(shell find . -type f -name 'go.mod' -print0 | xargs -0 -n1 dirname | sort)
CURRENT_DIR = $(shell pwd)

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "Unit tests"; \
	for module in $(SUB_MODULES); do \
		cd ${CURRENT_DIR}/$$module; \
		go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) ./... | tparse; \
	done
else
	@echo "Unit tests"; \
	for module in $(SUB_MODULES); do \
		cd ${CURRENT_DIR}/$$module; \
		go test -mod=readonly $(ARGS) $(TEST_PACKAGES) ./... ; \
	done
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)

test-cover:
	@export VERSION=$(VERSION);
	@bash scripts/test_cover.sh
.PHONY: test-cover

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)
.PHONY: benchmark


########################################
### Local validator nodes using docker and docker-compose

build-docker-fetchdnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/fetchd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/fetchd:Z tendermint/fetchdnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up -d

# Stop local testnet
localnet-stop:
	docker-compose down

.PHONY: all build-linux install install-debug \
	go-mod-cache draw-deps clean build \
	test test-all test-cover test-unit test-race


###############################################################################
###                                Protobuf                                 ###
###############################################################################

containerProtoVer=v0.2
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=${PROJECT_NAME}-proto-gen-$(containerProtoVer)
containerProtoFmt=${PROJECT_NAME}-proto-fmt-$(containerProtoVer)
containerProtoGenSwagger=${PROJECT_NAME}-proto-gen-swagger-$(containerProtoVer)

proto-all: proto-gen proto-lint proto-check-breaking proto-format
.PHONY: proto-all proto-gen proto-gen-docker proto-lint proto-check-breaking proto-format

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) sh ./scripts/protocgen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \; ; fi

proto-format-direct:
	find ./ -not -path "./third_party/*" -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-lint-direct:
	@buf lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=master

proto-check-breaking-direct:
	@buf breaking --against '.git#branch=master'

GOGO_PROTO_URL   = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
REGEN_COSMOS_PROTO_URL = https://raw.githubusercontent.com/regen-network/cosmos-proto/master
COSMOS_PROTO_URL   = https://raw.githubusercontent.com/cosmos/cosmos-sdk/master/proto/cosmos

GOGO_PROTO_TYPES    = third_party/proto/gogoproto
REGEN_COSMOS_PROTO_TYPES  = third_party/proto/cosmos_proto
COSMOS_PROTO_TYPES    = third_party/proto/cosmos

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(REGEN_COSMOS_PROTO_TYPES)
	@curl -sSL $(REGEN_COSMOS_PROTO_URL)/cosmos.proto > $(REGEN_COSMOS_PROTO_TYPES)/cosmos.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)/base/query/v1beta1/
	@curl -sSL $(COSMOS_PROTO_URL)/base/query/v1beta1/pagination.proto > $(COSMOS_PROTO_TYPES)/base/query/v1beta1/pagination.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)/base/v1beta1/
	@curl -sSL $(COSMOS_PROTO_URL)/base/v1beta1/coin.proto > $(COSMOS_PROTO_TYPES)/base/v1beta1/coin.proto
