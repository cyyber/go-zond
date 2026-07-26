// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
)

func TestConsumeStarlarkCompletionSuppressesSecretBearingTranscript(t *testing.T) {
	const secret = "seed-that-must-never-reach-errors"
	stream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
	close(stream)
	err := consumeStarlarkCompletion(stream)
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("secret-bearing error = %v", err)
	}

	incomplete := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	incomplete <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	close(incomplete)
	err = consumeStarlarkCompletion(incomplete)
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "without a terminal event") {
		t.Fatalf("incomplete error = %v", err)
	}

	success := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	success <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(secret, time.Second)
	close(success)
	if err := consumeStarlarkCompletion(success); err != nil {
		t.Fatal(err)
	}
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
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, want error = %t", err, test.wantError)
			}
			if test.wantDestroyCause && !errors.Is(err, test.destroyErr) {
				t.Fatalf("error %v does not preserve destroy cause", err)
			}
			if test.wantInspectCause && !errors.Is(err, test.inspectErr) {
				t.Fatalf("error %v does not preserve inspection cause", err)
			}
		})
	}
}
