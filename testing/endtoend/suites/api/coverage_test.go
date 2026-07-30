// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package api

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"unicode"

	qrldebug "github.com/theQRL/go-qrl/internal/debug"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	qrlnode "github.com/theQRL/go-qrl/qrl"
	"github.com/theQRL/go-qrl/qrl/downloader"
	"github.com/theQRL/go-qrl/qrl/filters"
	"github.com/theQRL/go-qrl/qrl/tracers"
)

const (
	coverageBehavior   = "live behavior"
	coverageShape      = "live response shape"
	coverageDispatch   = "live dispatch and error contract"
	coverageUnsafe     = "excluded: mutates node configuration, chain, or files"
	coverageNotExposed = "excluded: not exposed by the devnet profile"
	coverageInternal   = "excluded: internal compatibility callback"

	scenarioNodeMetadata             = "node-metadata"
	scenarioChainState               = "chain-state"
	scenarioTransactions             = "transactions"
	scenarioTxPool                   = "txpool"
	scenarioRuntimeDiagnostics       = "runtime-diagnostics"
	scenarioHistoricalLogs           = "historical-logs"
	scenarioBlockFilter              = "block-filter"
	scenarioPendingFilter            = "pending-filter"
	scenarioSubscriptionEvents       = "subscription-events"
	scenarioSubscriptionRegistration = "subscription-registration"
	scenarioRawDebug                 = "raw-debug"
	scenarioDebugState               = "debug-state"
	scenarioDebugTracing             = "debug-tracing"
	scenarioDebugErrorPaths          = "debug-error-paths"
	scenarioGraphQLSchema            = "graphql-schema"
	scenarioGraphQLQueries           = "graphql-queries"
	scenarioGraphQLMutation          = "graphql-mutation"
	scenarioGraphQLPending           = "graphql-pending"
)

type apiCoverageEntry struct {
	kind     string
	scenario string
}

func behavior(scenario string) apiCoverageEntry {
	return apiCoverageEntry{kind: coverageBehavior, scenario: scenario}
}

func shape(scenario string) apiCoverageEntry {
	return apiCoverageEntry{kind: coverageShape, scenario: scenario}
}

func dispatch(scenario string) apiCoverageEntry {
	return apiCoverageEntry{kind: coverageDispatch, scenario: scenario}
}

func excluded(kind string) apiCoverageEntry {
	return apiCoverageEntry{kind: kind}
}

var scenarioDescriptions = map[string]string{
	scenarioNodeMetadata:             "covers node and network metadata APIs",
	scenarioChainState:               "covers chain, account, state, call, proof, and fee APIs",
	scenarioTransactions:             "covers transaction lookup and raw encoding APIs",
	scenarioTxPool:                   "covers non-empty transaction-pool inspection APIs",
	scenarioRuntimeDiagnostics:       "covers runtime diagnostic APIs",
	scenarioHistoricalLogs:           "covers historical log filtering APIs",
	scenarioBlockFilter:              "covers newly mined block filters",
	scenarioPendingFilter:            "covers pending-transaction filters",
	scenarioSubscriptionEvents:       "covers emitted WebSocket events",
	scenarioSubscriptionRegistration: "covers passive WebSocket subscription registration",
	scenarioRawDebug:                 "covers raw debug chain APIs",
	scenarioDebugState:               "covers debug state diagnostics",
	scenarioDebugTracing:             "covers debug tracing APIs",
	scenarioDebugErrorPaths:          "covers registered debug and node-control error paths",
	scenarioGraphQLSchema:            "covers the GraphQL schema",
	scenarioGraphQLQueries:           "covers GraphQL query fields",
	scenarioGraphQLMutation:          "covers the GraphQL transaction mutation",
	scenarioGraphQLPending:           "covers GraphQL pending transactions",
}

var apiCoverage = map[string]apiCoverageEntry{
	"rpc_modules": behavior(scenarioNodeMetadata),

	"web3_clientVersion": behavior(scenarioNodeMetadata),
	"web3_sha3":          behavior(scenarioNodeMetadata),

	"net_listening": behavior(scenarioNodeMetadata),
	"net_peerCount": shape(scenarioNodeMetadata),
	"net_version":   behavior(scenarioNodeMetadata),

	"admin_addPeer":           dispatch(scenarioDebugErrorPaths),
	"admin_removePeer":        dispatch(scenarioDebugErrorPaths),
	"admin_addTrustedPeer":    dispatch(scenarioDebugErrorPaths),
	"admin_removeTrustedPeer": dispatch(scenarioDebugErrorPaths),
	"admin_peerEvents":        shape(scenarioSubscriptionRegistration),
	"admin_startHTTP":         excluded(coverageUnsafe),
	"admin_stopHTTP":          excluded(coverageUnsafe),
	"admin_startWS":           excluded(coverageUnsafe),
	"admin_stopWS":            excluded(coverageUnsafe),
	"admin_peers":             shape(scenarioNodeMetadata),
	"admin_nodeInfo":          behavior(scenarioNodeMetadata),
	"admin_datadir":           behavior(scenarioNodeMetadata),
	"admin_exportChain":       dispatch(scenarioDebugErrorPaths),
	"admin_importChain":       dispatch(scenarioDebugErrorPaths),

	"qrl_gasPrice":                               behavior(scenarioChainState),
	"qrl_maxPriorityFeePerGas":                   behavior(scenarioChainState),
	"qrl_feeHistory":                             behavior(scenarioChainState),
	"qrl_syncing":                                behavior(scenarioChainState),
	"qrl_accounts":                               excluded(coverageNotExposed),
	"qrl_chainId":                                behavior(scenarioChainState),
	"qrl_blockNumber":                            behavior(scenarioChainState),
	"qrl_getBalance":                             behavior(scenarioChainState),
	"qrl_getProof":                               behavior(scenarioChainState),
	"qrl_getHeaderByNumber":                      behavior(scenarioChainState),
	"qrl_getHeaderByHash":                        behavior(scenarioChainState),
	"qrl_getBlockByNumber":                       behavior(scenarioChainState),
	"qrl_getBlockByHash":                         behavior(scenarioChainState),
	"qrl_getCode":                                behavior(scenarioChainState),
	"qrl_getStorageAt":                           behavior(scenarioChainState),
	"qrl_getBlockReceipts":                       behavior(scenarioChainState),
	"qrl_call":                                   behavior(scenarioChainState),
	"qrl_estimateGas":                            behavior(scenarioChainState),
	"qrl_createAccessList":                       behavior(scenarioChainState),
	"qrl_getBlockTransactionCountByNumber":       behavior(scenarioTransactions),
	"qrl_getBlockTransactionCountByHash":         behavior(scenarioTransactions),
	"qrl_getTransactionByBlockNumberAndIndex":    behavior(scenarioTransactions),
	"qrl_getTransactionByBlockHashAndIndex":      behavior(scenarioTransactions),
	"qrl_getRawTransactionByBlockNumberAndIndex": behavior(scenarioTransactions),
	"qrl_getRawTransactionByBlockHashAndIndex":   behavior(scenarioTransactions),
	"qrl_getTransactionCount":                    behavior(scenarioTransactions),
	"qrl_getTransactionByHash":                   behavior(scenarioTransactions),
	"qrl_getRawTransactionByHash":                behavior(scenarioTransactions),
	"qrl_getTransactionReceipt":                  behavior(scenarioTransactions),
	"qrl_sendTransaction":                        excluded(coverageNotExposed),
	"qrl_fillTransaction":                        behavior(scenarioTransactions),
	"qrl_sendRawTransaction":                     behavior(scenarioTransactions),
	"qrl_sign":                                   excluded(coverageNotExposed),
	"qrl_signTransaction":                        excluded(coverageNotExposed),
	"qrl_pendingTransactions":                    behavior(scenarioTxPool),

	"qrl_newPendingTransactionFilter": behavior(scenarioPendingFilter),
	"qrl_newPendingTransactions":      behavior(scenarioSubscriptionEvents),
	"qrl_newBlockFilter":              behavior(scenarioBlockFilter),
	"qrl_newHeads":                    behavior(scenarioSubscriptionEvents),
	"qrl_logs":                        behavior(scenarioSubscriptionEvents),
	"qrl_newFilter":                   behavior(scenarioHistoricalLogs),
	"qrl_getLogs":                     behavior(scenarioHistoricalLogs),
	"qrl_uninstallFilter":             behavior(scenarioHistoricalLogs),
	"qrl_getFilterLogs":               behavior(scenarioHistoricalLogs),
	"qrl_getFilterChanges":            behavior(scenarioHistoricalLogs),
	"qrl_subscribeSyncStatus":         excluded(coverageInternal),

	"txpool_content":     behavior(scenarioTxPool),
	"txpool_contentFrom": behavior(scenarioTxPool),
	"txpool_status":      behavior(scenarioTxPool),
	"txpool_inspect":     behavior(scenarioTxPool),

	"debug_getRawHeader":                behavior(scenarioRawDebug),
	"debug_getRawBlock":                 behavior(scenarioRawDebug),
	"debug_getRawReceipts":              behavior(scenarioRawDebug),
	"debug_getRawTransaction":           behavior(scenarioRawDebug),
	"debug_printBlock":                  behavior(scenarioRawDebug),
	"debug_dbGet":                       behavior(scenarioRawDebug),
	"debug_dbAncient":                   dispatch(scenarioDebugErrorPaths),
	"debug_dbAncients":                  behavior(scenarioRawDebug),
	"debug_chaindbProperty":             dispatch(scenarioDebugErrorPaths),
	"debug_chaindbCompact":              excluded(coverageUnsafe),
	"debug_setHead":                     excluded(coverageUnsafe),
	"debug_dumpBlock":                   behavior(scenarioDebugState),
	"debug_preimage":                    dispatch(scenarioDebugErrorPaths),
	"debug_getBadBlocks":                behavior(scenarioDebugState),
	"debug_accountRange":                behavior(scenarioDebugState),
	"debug_storageRangeAt":              behavior(scenarioDebugState),
	"debug_getModifiedAccountsByNumber": dispatch(scenarioDebugState),
	"debug_getModifiedAccountsByHash":   dispatch(scenarioDebugState),
	"debug_getAccessibleState":          dispatch(scenarioDebugErrorPaths),
	"debug_setTrieFlushInterval":        dispatch(scenarioDebugErrorPaths),
	"debug_getTrieFlushInterval":        dispatch(scenarioDebugErrorPaths),
	"debug_traceChain":                  behavior(scenarioDebugTracing),
	"debug_traceBlockByNumber":          behavior(scenarioDebugTracing),
	"debug_traceBlockByHash":            behavior(scenarioDebugTracing),
	"debug_traceBlock":                  behavior(scenarioDebugTracing),
	"debug_traceBlockFromFile":          dispatch(scenarioDebugErrorPaths),
	"debug_traceBadBlock":               dispatch(scenarioDebugErrorPaths),
	"debug_standardTraceBlockToFile":    excluded(coverageUnsafe),
	"debug_intermediateRoots":           behavior(scenarioDebugTracing),
	"debug_standardTraceBadBlockToFile": dispatch(scenarioDebugErrorPaths),
	"debug_traceTransaction":            behavior(scenarioDebugTracing),
	"debug_traceCall":                   behavior(scenarioDebugTracing),

	"debug_verbosity":               excluded(coverageUnsafe),
	"debug_vmodule":                 excluded(coverageUnsafe),
	"debug_memStats":                behavior(scenarioRuntimeDiagnostics),
	"debug_gcStats":                 shape(scenarioRuntimeDiagnostics),
	"debug_cpuProfile":              excluded(coverageUnsafe),
	"debug_startCPUProfile":         excluded(coverageUnsafe),
	"debug_stopCPUProfile":          excluded(coverageUnsafe),
	"debug_goTrace":                 excluded(coverageUnsafe),
	"debug_startGoTrace":            excluded(coverageUnsafe),
	"debug_stopGoTrace":             excluded(coverageUnsafe),
	"debug_blockProfile":            excluded(coverageUnsafe),
	"debug_setBlockProfileRate":     excluded(coverageUnsafe),
	"debug_writeBlockProfile":       excluded(coverageUnsafe),
	"debug_mutexProfile":            excluded(coverageUnsafe),
	"debug_setMutexProfileFraction": excluded(coverageUnsafe),
	"debug_writeMutexProfile":       excluded(coverageUnsafe),
	"debug_writeMemProfile":         excluded(coverageUnsafe),
	"debug_stacks":                  behavior(scenarioRuntimeDiagnostics),
	"debug_freeOSMemory":            excluded(coverageUnsafe),
	"debug_setGCPercent":            excluded(coverageUnsafe),

	"miner_setExtra":    excluded(coverageNotExposed),
	"miner_setGasPrice": excluded(coverageNotExposed),
	"miner_setGasLimit": excluded(coverageNotExposed),
}

func TestAPICoverageManifest(t *testing.T) {
	for method, entry := range apiCoverage {
		if strings.TrimSpace(entry.kind) == "" {
			t.Errorf("%s has no coverage category", method)
		}
		if strings.HasPrefix(entry.kind, "live ") {
			if _, ok := scenarioDescriptions[entry.scenario]; !ok {
				t.Errorf("%s references unknown live scenario %q", method, entry.scenario)
			}
		} else if entry.scenario != "" {
			t.Errorf("%s is excluded but references scenario %q", method, entry.scenario)
		}
	}

	actual := map[string]struct{}{}
	for namespace, serviceTypes := range map[string][]reflect.Type{
		"qrl": {
			reflect.TypeFor[*qrlapi.QRLAPI](),
			reflect.TypeFor[*qrlapi.QRLAccountAPI](),
			reflect.TypeFor[*qrlapi.BlockChainAPI](),
			reflect.TypeFor[*qrlapi.TransactionAPI](),
			reflect.TypeFor[*downloader.DownloaderAPI](),
			reflect.TypeFor[*filters.FilterAPI](),
		},
		"txpool": {reflect.TypeFor[*qrlapi.TxPoolAPI]()},
		"net":    {reflect.TypeFor[*qrlapi.NetAPI]()},
		"admin":  {reflect.TypeFor[*qrlnode.AdminAPI]()},
		"debug": {
			reflect.TypeFor[*qrlapi.DebugAPI](),
			reflect.TypeFor[*qrlnode.DebugAPI](),
			reflect.TypeFor[*tracers.API](),
			reflect.TypeFor[*qrldebug.HandlerT](),
		},
		"miner": {reflect.TypeFor[*qrlnode.MinerAPI]()},
	} {
		for _, serviceType := range serviceTypes {
			for index := range serviceType.NumMethod() {
				method := serviceType.Method(index)
				actual[namespace+"_"+lowerFirst(method.Name)] = struct{}{}
			}
		}
	}
	for _, method := range []string{
		"rpc_modules",
		"web3_clientVersion",
		"web3_sha3",
		"admin_addPeer",
		"admin_removePeer",
		"admin_addTrustedPeer",
		"admin_removeTrustedPeer",
		"admin_peerEvents",
		"admin_startHTTP",
		"admin_stopHTTP",
		"admin_startWS",
		"admin_stopWS",
		"admin_peers",
		"admin_nodeInfo",
		"admin_datadir",
	} {
		actual[method] = struct{}{}
	}

	var missing, stale []string
	for method := range actual {
		if _, ok := apiCoverage[method]; !ok {
			missing = append(missing, method)
		}
	}
	for method := range apiCoverage {
		if _, ok := actual[method]; !ok {
			stale = append(stale, method)
		}
	}
	slices.Sort(missing)
	slices.Sort(stale)
	if len(missing) != 0 || len(stale) != 0 {
		t.Fatalf("API coverage manifest mismatch\nmissing: %v\nstale: %v", missing, stale)
	}
}

func lowerFirst(value string) string {
	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
