// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/qrlclient"
)

const chainAdvancementWindow = 30 * time.Second

func probeNetwork(ctx context.Context, rpcURL, walletAddress string) error {
	address, err := common.NewAddressFromString(walletAddress)
	if err != nil {
		return errors.New("signer readiness requires a valid wallet address")
	}
	client, err := qrlclient.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("dial RPC: %w", err)
	}
	defer client.Close()

	actualChainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("read chain ID: %w", err)
	}
	if actualChainID == nil || actualChainID.Sign() <= 0 {
		return fmt.Errorf("chain ID must be positive, got %v", actualChainID)
	}
	firstBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("read block number: %w", err)
	}
	if firstBlock == 0 {
		return errors.New("chain has not produced a post-genesis block")
	}
	advancementCtx, cancel := context.WithTimeout(ctx, chainAdvancementWindow)
	defer cancel()
	if err := retryUntil(advancementCtx, 500*time.Millisecond, 2*time.Second, func(attempt context.Context) error {
		block, err := client.BlockNumber(attempt)
		if err != nil {
			return fmt.Errorf("read advancing block number: %w", err)
		}
		if block <= firstBlock {
			return fmt.Errorf("block number remains at %d", block)
		}
		return nil
	}); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("chain advancement probe interrupted: %w", err)
		}
		return fmt.Errorf(
			"chain did not advance beyond block %d within %s: %w",
			firstBlock,
			chainAdvancementWindow,
			err,
		)
	}
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return fmt.Errorf("read development wallet balance: %w", err)
	}
	if balance.Sign() <= 0 {
		return fmt.Errorf("development wallet %s has no balance", walletAddress)
	}
	return nil
}
