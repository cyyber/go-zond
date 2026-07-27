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

func (networks *recordingNetworks) Stop(_ context.Context, name string) error {
	networks.call = "stop:" + name
	return nil
}

func TestRun(t *testing.T) {
	const enclaveName = "go-qrl-devnet-test"
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
				"start", "--enclave-name", enclaveName,
				"--execution-image", "local/go-qrl:test",
				"--params-file", paramsFile,
			},
			[]byte(`{"custom":true}`),
		},
		{"stop", "network stopped\n", "stop:" + enclaveName, []string{"stop", "--enclave-name", enclaveName}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			networks := new(recordingNetworks)
			var stdout, stderr bytes.Buffer
			app := newApp(networks)
			app.Writer, app.ErrWriter = &stdout, &stderr
			require.NoError(t, app.RunContext(t.Context(), append([]string{"devnet"}, test.arguments...)))
			require.Equal(t, test.output, stdout.String())
			require.Equal(t, test.call, networks.call)
			require.Empty(t, stderr.String())
			if test.call == "start" {
				require.Equal(t, network.StartOptions{
					EnclaveName:    enclaveName,
					ExecutionImage: "local/go-qrl:test",
					Parameters:     test.parameters,
				}, networks.start)
			}
		})
	}
}
