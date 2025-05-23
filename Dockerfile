# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# Build Gzond in a stock Go builder container
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev linux-headers git

# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /go-zond/
COPY go.sum /go-zond/
RUN cd /go-zond && go mod download

ADD . /go-zond
RUN cd /go-zond && go run build/ci.go install -static ./cmd/gzond

# Pull Gzond into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-zond/build/bin/gzond /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["gzond"]

# Add some metadata labels to help programmatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
