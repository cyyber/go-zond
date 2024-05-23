// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/naoina/toml"
	"github.com/theQRL/go-zond/accounts"
	"github.com/theQRL/go-zond/accounts/external"
	"github.com/theQRL/go-zond/cmd/utils"
	"github.com/theQRL/go-zond/internal/flags"
	"github.com/theQRL/go-zond/internal/version"
	"github.com/theQRL/go-zond/internal/zondapi"
	"github.com/theQRL/go-zond/log"
	"github.com/theQRL/go-zond/metrics"
	"github.com/theQRL/go-zond/node"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/zond/catalyst"
	"github.com/theQRL/go-zond/zond/downloader"
	"github.com/theQRL/go-zond/zond/zondconfig"
	"github.com/urfave/cli/v2"
)

var (
	dumpConfigCommand = &cli.Command{
		Action:      dumpConfig,
		Name:        "dumpconfig",
		Usage:       "Export configuration values in a TOML format",
		ArgsUsage:   "<dumpfile (optional)>",
		Flags:       flags.Merge(nodeFlags, rpcFlags),
		Description: `Export configuration values in TOML format (to stdout by default).`,
	}

	configFileFlag = &cli.StringFlag{
		Name:     "config",
		Usage:    "TOML configuration file",
		Category: flags.ZondCategory,
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		id := fmt.Sprintf("%s.%s", rt.String(), field)
		if deprecated(id) {
			log.Warn("Config field is deprecated and won't have an effect", "name", id)
			return nil
		}
		var link string
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type ethstatsConfig struct {
	URL string `toml:",omitempty"`
}

type gzondConfig struct {
	Zond      zondconfig.Config
	Node      node.Config
	Zondstats ethstatsConfig
	Metrics   metrics.Config
}

func loadConfig(file string, cfg *gzondConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	git, _ := version.VCS()
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(git.Commit, git.Date)
	cfg.HTTPModules = append(cfg.HTTPModules, "zond")
	cfg.WSModules = append(cfg.WSModules, "zond")
	cfg.IPCPath = "gzond.ipc"
	return cfg
}

// loadBaseConfig loads the gethConfig based on the given command line
// parameters and config file.
func loadBaseConfig(ctx *cli.Context) gzondConfig {
	// Load defaults.
	cfg := gzondConfig{
		Zond:    zondconfig.Defaults,
		Node:    defaultNodeConfig(),
		Metrics: metrics.DefaultConfig,
	}

	// Load config file.
	if file := ctx.String(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	return cfg
}

// makeConfigNode loads gzond configuration and creates a blank node instance.
func makeConfigNode(ctx *cli.Context) (*node.Node, gzondConfig) {
	cfg := loadBaseConfig(ctx)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	// Node doesn't by default populate account manager backends
	if err := setAccountManagerBackends(stack.Config(), stack.AccountManager()); err != nil {
		utils.Fatalf("Failed to set account manager backends: %v", err)
	}

	utils.SetZondConfig(ctx, stack, &cfg.Zond)
	if ctx.IsSet(utils.ZondStatsURLFlag.Name) {
		cfg.Zondstats.URL = ctx.String(utils.ZondStatsURLFlag.Name)
	}
	applyMetricConfig(ctx, &cfg)

	return stack, cfg
}

// makeFullNode loads gzond configuration and creates the Ethereum backend.
func makeFullNode(ctx *cli.Context) (*node.Node, zondapi.Backend) {
	stack, cfg := makeConfigNode(ctx)
	backend, zond := utils.RegisterZondService(stack, &cfg.Zond)

	// Create gauge with gzond system and build information
	if zond != nil {
		var protos []string
		for _, p := range zond.Protocols() {
			protos = append(protos, fmt.Sprintf("%v/%d", p.Name, p.Version))
		}
		metrics.NewRegisteredGaugeInfo("gzond/info", nil).Update(metrics.GaugeInfoValue{
			"arch":      runtime.GOARCH,
			"os":        runtime.GOOS,
			"version":   cfg.Node.Version,
			"protocols": strings.Join(protos, ","),
		})
	}

	// Configure log filter RPC API.
	filterSystem := utils.RegisterFilterAPI(stack, backend, &cfg.Zond)

	// Configure GraphQL if requested.
	if ctx.IsSet(utils.GraphQLEnabledFlag.Name) {
		utils.RegisterGraphQLService(stack, backend, filterSystem, &cfg.Node)
	}

	// Add the Ethereum Stats daemon if requested.
	if cfg.Zondstats.URL != "" {
		utils.RegisterZondStatsService(stack, backend, cfg.Zondstats.URL)
	}

	// Configure full-sync tester service if requested
	if ctx.IsSet(utils.SyncTargetFlag.Name) && cfg.Zond.SyncMode == downloader.FullSync {
		utils.RegisterFullSyncTester(stack, zond, ctx.Path(utils.SyncTargetFlag.Name))
	}

	// Start the dev mode if requested, or launch the engine API for
	// interacting with external consensus client.
	if ctx.IsSet(utils.DeveloperFlag.Name) {
		simBeacon, err := catalyst.NewSimulatedBeacon(ctx.Uint64(utils.DeveloperPeriodFlag.Name), zond)
		if err != nil {
			utils.Fatalf("failed to register dev mode catalyst service: %v", err)
		}
		catalyst.RegisterSimulatedBeaconAPIs(stack, simBeacon)
		stack.RegisterLifecycle(simBeacon)
	} else {
		err := catalyst.Register(stack, zond)
		if err != nil {
			utils.Fatalf("failed to register catalyst service: %v", err)
		}
	}
	return stack, backend
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Zond.Genesis != nil {
		cfg.Zond.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}

func applyMetricConfig(ctx *cli.Context, cfg *gzondConfig) {
	if ctx.IsSet(utils.MetricsEnabledFlag.Name) {
		cfg.Metrics.Enabled = ctx.Bool(utils.MetricsEnabledFlag.Name)
	}
	if ctx.IsSet(utils.MetricsEnabledExpensiveFlag.Name) {
		cfg.Metrics.EnabledExpensive = ctx.Bool(utils.MetricsEnabledExpensiveFlag.Name)
	}
	if ctx.IsSet(utils.MetricsHTTPFlag.Name) {
		cfg.Metrics.HTTP = ctx.String(utils.MetricsHTTPFlag.Name)
	}
	if ctx.IsSet(utils.MetricsPortFlag.Name) {
		cfg.Metrics.Port = ctx.Int(utils.MetricsPortFlag.Name)
	}
	if ctx.IsSet(utils.MetricsEnableInfluxDBFlag.Name) {
		cfg.Metrics.EnableInfluxDB = ctx.Bool(utils.MetricsEnableInfluxDBFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBEndpointFlag.Name) {
		cfg.Metrics.InfluxDBEndpoint = ctx.String(utils.MetricsInfluxDBEndpointFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBDatabaseFlag.Name) {
		cfg.Metrics.InfluxDBDatabase = ctx.String(utils.MetricsInfluxDBDatabaseFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBUsernameFlag.Name) {
		cfg.Metrics.InfluxDBUsername = ctx.String(utils.MetricsInfluxDBUsernameFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBPasswordFlag.Name) {
		cfg.Metrics.InfluxDBPassword = ctx.String(utils.MetricsInfluxDBPasswordFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBTagsFlag.Name) {
		cfg.Metrics.InfluxDBTags = ctx.String(utils.MetricsInfluxDBTagsFlag.Name)
	}
	if ctx.IsSet(utils.MetricsEnableInfluxDBV2Flag.Name) {
		cfg.Metrics.EnableInfluxDBV2 = ctx.Bool(utils.MetricsEnableInfluxDBV2Flag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBTokenFlag.Name) {
		cfg.Metrics.InfluxDBToken = ctx.String(utils.MetricsInfluxDBTokenFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBBucketFlag.Name) {
		cfg.Metrics.InfluxDBBucket = ctx.String(utils.MetricsInfluxDBBucketFlag.Name)
	}
	if ctx.IsSet(utils.MetricsInfluxDBOrganizationFlag.Name) {
		cfg.Metrics.InfluxDBOrganization = ctx.String(utils.MetricsInfluxDBOrganizationFlag.Name)
	}
}

func deprecated(field string) bool {
	return false
}

func setAccountManagerBackends(conf *node.Config, am *accounts.Manager) error {
	// Assemble the supported backends
	if len(conf.ExternalSigner) > 0 {
		log.Info("Using external signer", "url", conf.ExternalSigner)
		if extBackend, err := external.NewExternalBackend(conf.ExternalSigner); err == nil {
			am.AddBackend(extBackend)
			return nil
		} else {
			return fmt.Errorf("error connecting to external signer: %v", err)
		}
	}

	return nil
}
