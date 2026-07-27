// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"errors"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	engine_bindings "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
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

func TestFindExactEnclaveDistinguishesAbsenceIdentityAndAmbiguity(t *testing.T) {
	const (
		name  = "go-qrl-e2e-slot"
		uuid  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		uuid2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	info := func(name, uuid string) *engine_bindings.EnclaveInfo {
		return &engine_bindings.EnclaveInfo{Name: name, EnclaveUuid: uuid}
	}
	check := func(running map[string]*engine_bindings.EnclaveInfo, want EnclaveRef, wantErr string) {
		ref, found, err := findExactEnclave(running, name)
		if wantErr != "" {
			require.ErrorContains(t, err, wantErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, want.Name != "", found)
		require.Equal(t, want, ref)
	}
	check(map[string]*engine_bindings.EnclaveInfo{uuid: info("another-slot", uuid)}, EnclaveRef{}, "")
	check(map[string]*engine_bindings.EnclaveInfo{uuid: info(name, uuid)}, EnclaveRef{Name: name, UUID: uuid}, "")
	check(map[string]*engine_bindings.EnclaveInfo{uuid: info(name, "")}, EnclaveRef{}, "empty enclave identity")
	check(map[string]*engine_bindings.EnclaveInfo{
		uuid: info(name, uuid), uuid2: info(name, uuid2),
	}, EnclaveRef{}, "multiple running")
}

func TestDestroyStillExistsChecksUUIDAndDeterministicName(t *testing.T) {
	const (
		name        = "go-qrl-e2e-slot"
		uuid        = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		replacement = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	ref := EnclaveRef{Name: name, UUID: uuid}
	info := func(name, uuid string) *engine_bindings.EnclaveInfo {
		return &engine_bindings.EnclaveInfo{Name: name, EnclaveUuid: uuid}
	}
	for _, test := range []struct {
		name    string
		running map[string]*engine_bindings.EnclaveInfo
		want    bool
		wantErr string
	}{
		{"absent", map[string]*engine_bindings.EnclaveInfo{}, false, ""},
		{"original UUID remains", map[string]*engine_bindings.EnclaveInfo{
			uuid: info(name, uuid),
		}, true, ""},
		{"same-name replacement remains", map[string]*engine_bindings.EnclaveInfo{
			replacement: info(name, replacement),
		}, true, ""},
		{"unrelated enclave remains", map[string]*engine_bindings.EnclaveInfo{
			replacement: info("another-slot", replacement),
		}, false, ""},
		{"nil original identity", map[string]*engine_bindings.EnclaveInfo{
			uuid: nil,
		}, false, "nil identity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			exists, err := destroyStillExists(test.running, ref)
			require.Equal(t, test.want, exists)
			if test.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestReconcileDestroy(t *testing.T) {
	destroyErr := errors.New("destroy response lost")
	inspectErr := errors.New("enclave listing failed")
	for name, test := range map[string]struct {
		destroyErr, inspectErr error
		exists                 bool
		want                   []error
	}{
		"successful destroy":        {},
		"lost response but absent":  {destroyErr: destroyErr},
		"rejected and still exists": {destroyErr: destroyErr, exists: true, want: []error{destroyErr}},
		"inspection failed":         {destroyErr: destroyErr, inspectErr: inspectErr, want: []error{destroyErr, inspectErr}},
	} {
		t.Run(name, func(t *testing.T) {
			err := reconcileDestroy(test.destroyErr, test.inspectErr, test.exists)
			if len(test.want) == 0 {
				require.NoError(t, err)
				return
			}
			for _, cause := range test.want {
				require.ErrorIs(t, err, cause)
			}
		})
	}
	require.ErrorContains(
		t,
		reconcileDestroy(nil, nil, true),
		"deterministic enclave slot remains occupied",
	)
}
