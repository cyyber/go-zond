// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
)

type recordingNetworks struct {
	call     string
	deadline time.Time
}

func (networks *recordingNetworks) Start(ctx context.Context, directory, image string) error {
	networks.call = "start:" + directory + ":" + image
	networks.deadline, _ = ctx.Deadline()
	return nil
}

func (networks *recordingNetworks) Inspect(_ context.Context, directory string) (network.Environment, error) {
	networks.call = "status:" + directory
	return network.Environment{}, nil
}

func (networks *recordingNetworks) Stop(_ context.Context, directory string) error {
	networks.call = "stop:" + directory
	return nil
}

func TestRun(t *testing.T) {
	networkDir := t.TempDir()
	for _, test := range []struct {
		name, output, call string
		arguments          []string
		timeout            bool
	}{
		{
			"start", "network ready\n", "start:" + networkDir + ":local/go-qrl:test",
			[]string{"start", "--network-dir", networkDir, "--execution-image", "local/go-qrl:test", "--timeout", "17m"}, true,
		},
		{"status", "network ready\n", "status:" + networkDir, []string{"status", "--network-dir", networkDir}, false},
		{"stop", "network stopped\n", "stop:" + networkDir, []string{"stop", "--network-dir", networkDir}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			networks := new(recordingNetworks)
			var stdout, stderr bytes.Buffer
			before := time.Now()
			require.NoError(t, run(t.Context(), test.arguments, &stdout, &stderr, networks))
			after := time.Now()
			require.Equal(t, test.output, stdout.String())
			require.Equal(t, test.call, networks.call)
			require.Empty(t, stderr.String())
			if test.timeout {
				require.WithinRange(t, networks.deadline, before.Add(17*time.Minute), after.Add(17*time.Minute))
			}
		})
	}
}
