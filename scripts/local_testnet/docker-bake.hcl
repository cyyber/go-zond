variable "E2E_LOCAL_EL_IMAGE" {
  default = "local/go-qrl:e2e"
}

variable "E2E_LOCAL_CL_IMAGE" {
  default = "local/qrysm-beacon:8b80fa0c3f5a"
}

variable "E2E_LOCAL_VC_IMAGE" {
  default = "local/qrysm-validator:8b80fa0c3f5a"
}

variable "E2E_LOCAL_GENESIS_IMAGE" {
  default = "local/qrl-genesis-generator:360410c72353-8b80fa0c3f5a"
}

variable "GO_QRL_COMMIT" {
  default = ""
}

group "network" {
  targets = ["execution", "beacon", "validator", "genesis"]
}

group "support" {
  targets = ["beacon", "validator", "genesis"]
}

target "_local" {
  output = ["type=docker"]
}

target "execution" {
  inherits   = ["_local"]
  context    = "."
  dockerfile = "Dockerfile"
  tags       = [E2E_LOCAL_EL_IMAGE]
  args = {
    COMMIT               = GO_QRL_COMMIT
    GO_BUILDER_IMAGE     = "golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587"
    ALPINE_RUNTIME_IMAGE = "alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b"
  }
}

target "_support" {
  inherits   = ["_local"]
  context    = "scripts/local_testnet"
  dockerfile = "Dockerfile.support"
  args = {
    QRYSM_GO_BUILDER_IMAGE = "golang:1.25-bookworm@sha256:ea341baa9bd5ba6784f6d7161ace70544349a6242d54d34a0fbfd2c4d51c9d58"
    QRYSM_CL_BASE_IMAGE     = "qrledger/qrysm:beacon-chain-latest@sha256:52b6fbecfe442d0d451e1219652e464d69de8a09edd44d5c54bbbf5ebdb83000"
    QRYSM_VC_BASE_IMAGE     = "qrledger/qrysm:validator-latest@sha256:e830b41130a43211803fe3d17eeb0a66cd743f062d5407667ee3531bc5891ede"
    GENESIS_BASE_IMAGE      = "qrledger/qrysm:qrl-genesis-generator-latest@sha256:43d975e6b5e22e4de79d9027325cc05f996c2325705f2e199e012788e5faa0eb"
    QRYSM_GIT_REPO          = "https://github.com/rgeraldes24/qrysm.git"
    QRYSM_GIT_COMMIT        = "8b80fa0c3f5a98f2edc3fc8b7b9c67808373cafb"
    GENERATOR_GIT_REPO      = "https://github.com/cyyber/qrl-genesis-generator.git"
    GENERATOR_GIT_COMMIT    = "360410c72353c3a337f078018b36877dbbe40799"
  }
}

target "beacon" {
  inherits = ["_support"]
  target   = "beacon"
  tags     = [E2E_LOCAL_CL_IMAGE]
}

target "validator" {
  inherits = ["_support"]
  target   = "validator"
  tags     = [E2E_LOCAL_VC_IMAGE]
}

target "genesis" {
  inherits = ["_support"]
  target   = "genesis"
  tags     = [E2E_LOCAL_GENESIS_IMAGE]
}
