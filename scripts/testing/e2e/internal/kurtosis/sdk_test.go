// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"errors"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
)

func TestConsumeStarlarkCompletionSuppressesSecretBearingTranscript(t *testing.T) {
	const secret = "seed-that-must-never-reach-errors"
	stream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
	close(stream)
	err := consumeStarlarkCompletion(stream)
	require.Error(t, err)
	require.NotContains(t, err.Error(), secret)

	incomplete := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	incomplete <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	close(incomplete)
	err = consumeStarlarkCompletion(incomplete)
	require.Error(t, err)
	require.NotContains(t, err.Error(), secret)
	require.ErrorContains(t, err, "without a terminal event")

	success := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	success <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(secret, time.Second)
	close(success)
	require.NoError(t, consumeStarlarkCompletion(success))
}

func TestReconcileDestroy(t *testing.T) {
	destroyErr := errors.New("destroy response lost")
	inspectErr := errors.New("enclave listing failed")
	for name, test := range map[string]struct {
		destroyErr, inspectErr error
		exists                 bool
		wantError              bool
		wantDestroyCause       bool
		wantInspectCause       bool
	}{
		"successful destroy":          {},
		"lost response but absent":    {destroyErr: destroyErr},
		"already absent":              {destroyErr: errors.New("not found")},
		"destroy rejected and exists": {destroyErr: destroyErr, exists: true, wantError: true, wantDestroyCause: true},
		"inspection failed": {
			destroyErr: destroyErr, inspectErr: inspectErr,
			wantError: true, wantDestroyCause: true, wantInspectCause: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := reconcileDestroy(test.destroyErr, test.inspectErr, test.exists)
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if test.wantDestroyCause {
				require.ErrorIs(t, err, test.destroyErr)
			}
			if test.wantInspectCause {
				require.ErrorIs(t, err, test.inspectErr)
			}
		})
	}
}
