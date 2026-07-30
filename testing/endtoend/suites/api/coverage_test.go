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
	coverageLive       = "live behavior"
	coverageLiveError  = "live dispatch and error contract"
	coverageUnsafe     = "excluded: mutates node configuration, chain, or files"
	coverageNotExposed = "excluded: not exposed by the devnet profile"
	coverageInternal   = "excluded: internal compatibility callback"
)

var apiCoverage = map[string]string{
	"rpc_modules": coverageLive,

	"web3_clientVersion": coverageLive,
	"web3_sha3":          coverageLive,

	"net_listening": coverageLive,
	"net_peerCount": coverageLive,
	"net_version":   coverageLive,

	"admin_addPeer":           coverageLiveError,
	"admin_removePeer":        coverageLiveError,
	"admin_addTrustedPeer":    coverageLiveError,
	"admin_removeTrustedPeer": coverageLiveError,
	"admin_peerEvents":        coverageLive,
	"admin_startHTTP":         coverageUnsafe,
	"admin_stopHTTP":          coverageUnsafe,
	"admin_startWS":           coverageUnsafe,
	"admin_stopWS":            coverageUnsafe,
	"admin_peers":             coverageLive,
	"admin_nodeInfo":          coverageLive,
	"admin_datadir":           coverageLive,
	"admin_exportChain":       coverageLiveError,
	"admin_importChain":       coverageLiveError,

	"qrl_gasPrice":                               coverageLive,
	"qrl_maxPriorityFeePerGas":                   coverageLive,
	"qrl_feeHistory":                             coverageLive,
	"qrl_syncing":                                coverageLive,
	"qrl_accounts":                               coverageLive,
	"qrl_chainId":                                coverageLive,
	"qrl_blockNumber":                            coverageLive,
	"qrl_getBalance":                             coverageLive,
	"qrl_getProof":                               coverageLive,
	"qrl_getHeaderByNumber":                      coverageLive,
	"qrl_getHeaderByHash":                        coverageLive,
	"qrl_getBlockByNumber":                       coverageLive,
	"qrl_getBlockByHash":                         coverageLive,
	"qrl_getCode":                                coverageLive,
	"qrl_getStorageAt":                           coverageLive,
	"qrl_getBlockReceipts":                       coverageLive,
	"qrl_call":                                   coverageLive,
	"qrl_estimateGas":                            coverageLive,
	"qrl_createAccessList":                       coverageLive,
	"qrl_getBlockTransactionCountByNumber":       coverageLive,
	"qrl_getBlockTransactionCountByHash":         coverageLive,
	"qrl_getTransactionByBlockNumberAndIndex":    coverageLive,
	"qrl_getTransactionByBlockHashAndIndex":      coverageLive,
	"qrl_getRawTransactionByBlockNumberAndIndex": coverageLive,
	"qrl_getRawTransactionByBlockHashAndIndex":   coverageLive,
	"qrl_getTransactionCount":                    coverageLive,
	"qrl_getTransactionByHash":                   coverageLive,
	"qrl_getRawTransactionByHash":                coverageLive,
	"qrl_getTransactionReceipt":                  coverageLive,
	"qrl_sendTransaction":                        coverageLive,
	"qrl_fillTransaction":                        coverageLive,
	"qrl_sendRawTransaction":                     coverageLive,
	"qrl_sign":                                   coverageLive,
	"qrl_signTransaction":                        coverageLive,
	"qrl_pendingTransactions":                    coverageLive,

	"qrl_newPendingTransactionFilter": coverageLive,
	"qrl_newPendingTransactions":      coverageLive,
	"qrl_newBlockFilter":              coverageLive,
	"qrl_newHeads":                    coverageLive,
	"qrl_logs":                        coverageLive,
	"qrl_newFilter":                   coverageLive,
	"qrl_getLogs":                     coverageLive,
	"qrl_uninstallFilter":             coverageLive,
	"qrl_getFilterLogs":               coverageLive,
	"qrl_getFilterChanges":            coverageLive,
	"qrl_subscribeSyncStatus":         coverageInternal,

	"txpool_content":     coverageLive,
	"txpool_contentFrom": coverageLive,
	"txpool_status":      coverageLive,
	"txpool_inspect":     coverageLive,

	"debug_getRawHeader":                coverageLive,
	"debug_getRawBlock":                 coverageLive,
	"debug_getRawReceipts":              coverageLive,
	"debug_getRawTransaction":           coverageLive,
	"debug_printBlock":                  coverageLive,
	"debug_dbGet":                       coverageLive,
	"debug_dbAncient":                   coverageLiveError,
	"debug_dbAncients":                  coverageLive,
	"debug_chaindbProperty":             coverageLiveError,
	"debug_chaindbCompact":              coverageUnsafe,
	"debug_setHead":                     coverageUnsafe,
	"debug_dumpBlock":                   coverageLive,
	"debug_preimage":                    coverageLiveError,
	"debug_getBadBlocks":                coverageLive,
	"debug_accountRange":                coverageLive,
	"debug_storageRangeAt":              coverageLive,
	"debug_getModifiedAccountsByNumber": coverageLive,
	"debug_getModifiedAccountsByHash":   coverageLive,
	"debug_getAccessibleState":          coverageLiveError,
	"debug_setTrieFlushInterval":        coverageLiveError,
	"debug_getTrieFlushInterval":        coverageLiveError,
	"debug_traceChain":                  coverageLive,
	"debug_traceBlockByNumber":          coverageLive,
	"debug_traceBlockByHash":            coverageLive,
	"debug_traceBlock":                  coverageLive,
	"debug_traceBlockFromFile":          coverageLiveError,
	"debug_traceBadBlock":               coverageLiveError,
	"debug_standardTraceBlockToFile":    coverageUnsafe,
	"debug_intermediateRoots":           coverageLive,
	"debug_standardTraceBadBlockToFile": coverageLiveError,
	"debug_traceTransaction":            coverageLive,
	"debug_traceCall":                   coverageLive,

	"debug_verbosity":               coverageUnsafe,
	"debug_vmodule":                 coverageUnsafe,
	"debug_memStats":                coverageLive,
	"debug_gcStats":                 coverageLive,
	"debug_cpuProfile":              coverageUnsafe,
	"debug_startCPUProfile":         coverageUnsafe,
	"debug_stopCPUProfile":          coverageUnsafe,
	"debug_goTrace":                 coverageUnsafe,
	"debug_startGoTrace":            coverageUnsafe,
	"debug_stopGoTrace":             coverageUnsafe,
	"debug_blockProfile":            coverageUnsafe,
	"debug_setBlockProfileRate":     coverageUnsafe,
	"debug_writeBlockProfile":       coverageUnsafe,
	"debug_mutexProfile":            coverageUnsafe,
	"debug_setMutexProfileFraction": coverageUnsafe,
	"debug_writeMutexProfile":       coverageUnsafe,
	"debug_writeMemProfile":         coverageUnsafe,
	"debug_stacks":                  coverageLive,
	"debug_freeOSMemory":            coverageUnsafe,
	"debug_setGCPercent":            coverageUnsafe,

	"miner_setExtra":    coverageNotExposed,
	"miner_setGasPrice": coverageNotExposed,
	"miner_setGasLimit": coverageNotExposed,
}

func TestAPICoverageManifest(t *testing.T) {
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

func TestAPICoverageManifestCategories(t *testing.T) {
	for method, category := range apiCoverage {
		if strings.TrimSpace(category) == "" {
			t.Errorf("%s has no coverage category", method)
		}
	}
}
