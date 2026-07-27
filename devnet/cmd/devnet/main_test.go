// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-qrl. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/devnet/internal/network"
)

type recordingNetworks struct {
	call  string
	start network.StartOptions
}

func (networks *recordingNetworks) Start(_ context.Context, options network.StartOptions) error {
	networks.call = "start"
	networks.start = options
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
	paramsFile := filepath.Join(t.TempDir(), "params.json")
	require.NoError(t, os.WriteFile(paramsFile, []byte(`{"custom":true}`), 0o600))
	for _, test := range []struct {
		name, output, call string
		arguments          []string
		parameters         []byte
	}{
		{
			"start with custom parameters", "network ready\n", "start",
			[]string{
				"start", "--network-dir", networkDir,
				"--execution-image", "local/go-qrl:test",
				"--params-file", paramsFile,
			},
			[]byte(`{"custom":true}`),
		},
		{"status", "network ready\n", "status:" + networkDir, []string{"status", "--network-dir", networkDir}, nil},
		{"stop", "network stopped\n", "stop:" + networkDir, []string{"stop", "--network-dir", networkDir}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			networks := new(recordingNetworks)
			var stdout, stderr bytes.Buffer
			require.NoError(t, run(t.Context(), test.arguments, &stdout, &stderr, networks))
			require.Equal(t, test.output, stdout.String())
			require.Equal(t, test.call, networks.call)
			require.Empty(t, stderr.String())
			if test.call == "start" {
				require.Equal(t, network.StartOptions{
					Directory:      networkDir,
					ExecutionImage: "local/go-qrl:test",
					Parameters:     test.parameters,
				}, networks.start)
			}
		})
	}
}
