// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/qrlclient"
)

const chainAdvancementWindow = 30 * time.Second

func probeNetwork(ctx context.Context, rpcURL string, address common.Address) error {
	client, err := qrlclient.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("dial RPC: %w", err)
	}
	defer client.Close()

	firstBlock, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("read block number: %w", err)
	}

	advancementCtx, cancel := context.WithTimeout(ctx, chainAdvancementWindow)
	defer cancel()
	if err := retryUntil(advancementCtx, func() error {
		block, err := client.BlockNumber(advancementCtx)
		if err != nil {
			return fmt.Errorf("read advancing block number: %w", err)
		}
		if block <= firstBlock {
			return fmt.Errorf("block number remains at %d", block)
		}
		return nil
	}); err != nil {
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
		return fmt.Errorf("development wallet %s has no balance", address.Hex())
	}
	return nil
}
