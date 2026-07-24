// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

const (
	packageLocator  = "github.com/rgeraldes24/qrl-package@1f31cd03dbe2061225701ea79d956cfeceaf91db"
	expectedChainID = "0x539"
	prefundBalance  = "2000000QRL"
	qrysmCommit     = "8b80fa0c3f5a98f2edc3fc8b7b9c67808373cafb"
	genesisCommit   = "3884e4228ef347bbfa4a80a90e819cd05140f122"

	executionServiceName = "el-1-gqrl-qrysm"
	consensusServiceName = "cl-1-qrysm-gqrl"
	validatorServiceName = "vc-1-gqrl-qrysm"
	rpcPortID            = "rpc"
	webSocketPortID      = "ws"
	graphQLPath          = "/graphql"
)

var (
	localImageRefs = map[string]string{
		"execution": "local/go-qrl:network",
		"consensus": "local/qrysm-beacon:8b80fa0c3f5a",
		"validator": "local/qrysm-validator:8b80fa0c3f5a",
		"genesis":   "local/qrl-genesis-generator:3884e4228ef3-8b80fa0c3f5a",
	}
	requiredServices = map[string]string{
		"execution": executionServiceName,
		"consensus": consensusServiceName,
		"validator": validatorServiceName,
	}
)

func pinnedBuildEnvironment() []string {
	return []string{
		"E2E_PINNED_GO_BUILDER_IMAGE=golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587",
		"E2E_PINNED_ALPINE_RUNTIME_IMAGE=alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b",
		"E2E_PINNED_QRYSM_GO_BUILDER_IMAGE=golang:1.25-bookworm@sha256:ea341baa9bd5ba6784f6d7161ace70544349a6242d54d34a0fbfd2c4d51c9d58",
		"E2E_PINNED_CL_BASE_IMAGE=qrledger/qrysm:beacon-chain-latest@sha256:52b6fbecfe442d0d451e1219652e464d69de8a09edd44d5c54bbbf5ebdb83000",
		"E2E_PINNED_VC_BASE_IMAGE=qrledger/qrysm:validator-latest@sha256:e830b41130a43211803fe3d17eeb0a66cd743f062d5407667ee3531bc5891ede",
		"E2E_PINNED_QRYSM_GIT_REPO=https://github.com/rgeraldes24/qrysm.git",
		"E2E_PINNED_QRYSM_GIT_COMMIT=" + qrysmCommit,
		"E2E_PINNED_GENESIS_GO_BUILDER_IMAGE=golang:1.25-bookworm@sha256:ea341baa9bd5ba6784f6d7161ace70544349a6242d54d34a0fbfd2c4d51c9d58",
		"E2E_PINNED_GENESIS_BASE_IMAGE=qrledger/qrysm:qrl-genesis-generator-latest@sha256:43d975e6b5e22e4de79d9027325cc05f996c2325705f2e199e012788e5faa0eb",
		"E2E_PINNED_GENERATOR_GIT_REPO=https://github.com/rgeraldes24/qrl-genesis-generator.git",
		"E2E_PINNED_GENERATOR_GIT_COMMIT=" + genesisCommit,
	}
}
