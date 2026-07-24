// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/qrlclient"
)

const chainAdvancementWindow = 30 * time.Second

type probeRequest struct {
	RPCURL, Address, ExpectedChainID string
}

func probeNetwork(ctx context.Context, request probeRequest) error {
	address, err := common.NewAddressFromString(request.Address)
	if err != nil {
		return errors.New("signer readiness requires a valid wallet address")
	}
	expectedChainID, ok := parseChainID(request.ExpectedChainID)
	if !ok {
		return errors.New("expected chain ID is invalid")
	}
	client, err := qrlclient.DialContext(ctx, request.RPCURL)
	if err != nil {
		return fmt.Errorf("dial RPC: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("read chain ID: %w", err)
	}
	if chainID.Cmp(expectedChainID) != 0 {
		return fmt.Errorf("chain ID %s differs from expected %s", chainID, expectedChainID)
	}
	firstBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("read block number: %w", err)
	}
	if firstBlock == 0 {
		return errors.New("chain has not produced a post-genesis block")
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(chainAdvancementWindow)
	defer deadline.Stop()
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-deadline.C:
			return fmt.Errorf("chain did not advance beyond block %d within %s", firstBlock, chainAdvancementWindow)
		case <-ticker.C:
			block, err := client.BlockNumber(ctx)
			if err != nil {
				return fmt.Errorf("read advancing block number: %w", err)
			}
			if block > firstBlock {
				balance, err := client.BalanceAt(ctx, address, nil)
				if err != nil {
					return fmt.Errorf("read E2E wallet balance: %w", err)
				}
				if balance.Sign() <= 0 {
					return fmt.Errorf("E2E wallet %s has no balance", request.Address)
				}
				return nil
			}
		}
	}
}

func parseChainID(value string) (*big.Int, bool) {
	value = strings.TrimSpace(value)
	base := 10
	if strings.HasPrefix(strings.ToLower(value), "0x") {
		base, value = 16, value[2:]
	}
	parsed, ok := new(big.Int).SetString(value, base)
	return parsed, ok && parsed.Sign() >= 0
}
