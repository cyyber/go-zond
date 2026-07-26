# Pin both build stages while allowing explicit base-image overrides.
ARG GO_BUILDER_IMAGE="golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2"
ARG ALPINE_RUNTIME_IMAGE="alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b"

# Build Gqrl in a stock Go builder container
FROM ${GO_BUILDER_IMAGE} AS builder

ARG COMMIT=""

RUN apk add --no-cache gcc musl-dev linux-headers git

# Get dependencies - will also be cached if we won't change go.mod/go.sum
WORKDIR /go-qrl
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go run build/ci.go install -git-commit="$COMMIT" -static ./cmd/gqrl

# Pull Gqrl into a second stage deploy alpine container
FROM ${ALPINE_RUNTIME_IMAGE}

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-qrl/build/bin/gqrl /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["gqrl"]

# Add some metadata labels to help programmatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL org.opencontainers.image.revision="$COMMIT" \
      commit="$COMMIT" \
      version="$VERSION" \
      buildnum="$BUILDNUM"
