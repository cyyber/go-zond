// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
)

type recordingNetworks struct {
	call       string
	deadline   time.Time
	inspectErr error
}

func (networks *recordingNetworks) Start(ctx context.Context, directory, image string) error {
	networks.call = "start:" + directory + ":" + image
	networks.deadline, _ = ctx.Deadline()
	return nil
}

func (networks *recordingNetworks) Inspect(_ context.Context, directory string) (network.Environment, error) {
	networks.call = "status:" + directory
	return network.Environment{}, networks.inspectErr
}

func (networks *recordingNetworks) Stop(_ context.Context, directory string) error {
	networks.call = "stop:" + directory
	return nil
}

func TestRun(t *testing.T) {
	networkDir := t.TempDir()
	statusErr := errors.New("network is not running")
	for _, test := range []struct {
		name, output, call, errorText string
		arguments                     []string
		inspectErr                    error
		timeout                       bool
	}{
		{
			"start", "network ready\n", "start:" + networkDir + ":local/go-qrl:test", "",
			[]string{"start", "--network-dir", networkDir, "--execution-image", "local/go-qrl:test", "--timeout", "17m"}, nil, true,
		},
		{"status", "network ready\n", "status:" + networkDir, "", []string{"status", "--network-dir", networkDir}, nil, false},
		{"stop", "network stopped\n", "stop:" + networkDir, "", []string{"stop", "--network-dir", networkDir}, nil, false},
		{"unknown command", "", "", "unknown command", []string{"network", "status"}, nil, false},
		{"missing directory", "", "", "--network-dir is required", []string{"status"}, nil, false},
		{
			"missing image", "", "", "--execution-image is required",
			[]string{"start", "--network-dir", networkDir}, nil, false,
		},
		{"failed status", "", "status:" + networkDir, statusErr.Error(), []string{"status", "--network-dir", networkDir}, statusErr, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			networks := &recordingNetworks{inspectErr: test.inspectErr}
			var stdout, stderr bytes.Buffer
			before := time.Now()
			err := run(t.Context(), test.arguments, &stdout, &stderr, networks)
			after := time.Now()
			if test.errorText == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.errorText)
			}
			require.Equal(t, test.output, stdout.String())
			require.Equal(t, test.call, networks.call)
			require.Empty(t, stderr.String())
			if test.inspectErr != nil {
				require.ErrorIs(t, err, test.inspectErr)
			}
			if test.timeout {
				require.WithinRange(t, networks.deadline, before.Add(17*time.Minute), after.Add(17*time.Minute))
			}
		})
	}
}
