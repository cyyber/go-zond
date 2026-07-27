// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
)

func TestConsumeStarlarkCompletionSuppressesSecretBearingTranscript(t *testing.T) {
	const secret = "seed-that-must-never-reach-errors"
	progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	tests := []struct {
		name, want string
		lines      []*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine
	}{
		{"success", "", []*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
			binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(secret, time.Second),
		}},
		{"failure", "failed", []*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
			progress, binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent(),
		}},
		{"missing terminal event", "without a terminal event", []*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{progress}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, len(test.lines))
			for _, line := range test.lines {
				stream <- line
			}
			close(stream)
			err := consumeStarlarkCompletion(stream)
			if test.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.want)
			require.NotContains(t, err.Error(), secret)
		})
	}
}
