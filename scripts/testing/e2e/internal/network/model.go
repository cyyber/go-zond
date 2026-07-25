// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns the lifecycle and immutable identity of separately
// managed E2E networks.
package network

type Environment struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	SeedFile     string
}

type StartRequest struct {
	NetworkDir     string
	ExecutionImage string
}
