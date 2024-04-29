// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package zond implements the Zond protocol.
package zond

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"

	"github.com/theQRL/go-zond/accounts"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/go-zond/consensus"
	"github.com/theQRL/go-zond/core"
	"github.com/theQRL/go-zond/core/bloombits"
	"github.com/theQRL/go-zond/core/rawdb"
	"github.com/theQRL/go-zond/core/state/pruner"
	"github.com/theQRL/go-zond/core/txpool"
	"github.com/theQRL/go-zond/core/txpool/legacypool"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/core/vm"
	"github.com/theQRL/go-zond/event"
	"github.com/theQRL/go-zond/internal/shutdowncheck"
	"github.com/theQRL/go-zond/internal/zondapi"
	"github.com/theQRL/go-zond/log"
	"github.com/theQRL/go-zond/miner"
	"github.com/theQRL/go-zond/node"
	"github.com/theQRL/go-zond/p2p"
	"github.com/theQRL/go-zond/p2p/dnsdisc"
	"github.com/theQRL/go-zond/p2p/enode"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/rlp"
	"github.com/theQRL/go-zond/rpc"
	"github.com/theQRL/go-zond/zond/downloader"
	"github.com/theQRL/go-zond/zond/gasprice"
	"github.com/theQRL/go-zond/zond/protocols/snap"
	"github.com/theQRL/go-zond/zond/protocols/zond"
	"github.com/theQRL/go-zond/zond/zondconfig"
	"github.com/theQRL/go-zond/zonddb"
)

// Zond implements the Zond full node service.
type Zond struct {
	config *zondconfig.Config

	// Handlers
	txPool *txpool.TxPool

	blockchain         *core.BlockChain
	handler            *handler
	ethDialCandidates  enode.Iterator
	snapDialCandidates enode.Iterator

	// DB interfaces
	chainDb zonddb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer      *core.ChainIndexer             // Bloom indexer operating during block imports
	closeBloomHandler chan struct{}

	APIBackend *ZondAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	etherbase common.Address

	networkID     uint64
	netRPCService *zondapi.NetAPI

	p2pServer *p2p.Server

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)

	shutdownTracker *shutdowncheck.ShutdownTracker // Tracks if and when the node has shutdown ungracefully
}

// New creates a new Zond object (including the
// initialisation of the common Zond object)
func New(stack *node.Node, config *zondconfig.Config) (*Zond, error) {
	// Ensure configuration values are compatible and sane
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	if config.Miner.GasPrice == nil || config.Miner.GasPrice.Cmp(common.Big0) <= 0 {
		log.Warn("Sanitizing invalid miner gas price", "provided", config.Miner.GasPrice, "updated", zondconfig.Defaults.Miner.GasPrice)
		config.Miner.GasPrice = new(big.Int).Set(zondconfig.Defaults.Miner.GasPrice)
	}
	if config.NoPruning && config.TrieDirtyCache > 0 {
		if config.SnapshotCache > 0 {
			config.TrieCleanCache += config.TrieDirtyCache * 3 / 5
			config.SnapshotCache += config.TrieDirtyCache * 2 / 5
		} else {
			config.TrieCleanCache += config.TrieDirtyCache
		}
		config.TrieDirtyCache = 0
	}
	log.Info("Allocated trie memory caches", "clean", common.StorageSize(config.TrieCleanCache)*1024*1024, "dirty", common.StorageSize(config.TrieDirtyCache)*1024*1024)

	// Assemble the Zond object
	chainDb, err := stack.OpenDatabaseWithFreezer("chaindata", config.DatabaseCache, config.DatabaseHandles, config.DatabaseFreezer, "eth/db/chaindata/", false)
	if err != nil {
		return nil, err
	}
	// Try to recover offline state pruning only in hash-based.
	if config.StateScheme == rawdb.HashScheme {
		if err := pruner.RecoverPruning(stack.ResolvePath(""), chainDb); err != nil {
			log.Error("Failed to recover state", "error", err)
		}
	}
	// Transfer mining-related config to the ethash config.
	chainConfig, err := core.LoadChainConfig(chainDb, config.Genesis)
	if err != nil {
		return nil, err
	}
	engine, err := zondconfig.CreateConsensusEngine(chainConfig, chainDb)
	if err != nil {
		return nil, err
	}
	zond := &Zond{
		config:            config,
		chainDb:           chainDb,
		eventMux:          stack.EventMux(),
		accountManager:    stack.AccountManager(),
		engine:            engine,
		closeBloomHandler: make(chan struct{}),
		networkID:         config.NetworkId,
		gasPrice:          config.Miner.GasPrice,
		etherbase:         config.Miner.Etherbase,
		bloomRequests:     make(chan chan *bloombits.Retrieval),
		bloomIndexer:      core.NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms),
		p2pServer:         stack.Server(),
		shutdownTracker:   shutdowncheck.NewShutdownTracker(chainDb),
	}
	bcVersion := rawdb.ReadDatabaseVersion(chainDb)
	var dbVer = "<nil>"
	if bcVersion != nil {
		dbVer = fmt.Sprintf("%d", *bcVersion)
	}
	log.Info("Initialising Zond protocol", "network", config.NetworkId, "dbversion", dbVer)

	if !config.SkipBcVersionCheck {
		if bcVersion != nil && *bcVersion > core.BlockChainVersion {
			return nil, fmt.Errorf("database version is v%d, Gzond %s only supports v%d", *bcVersion, params.VersionWithMeta, core.BlockChainVersion)
		} else if bcVersion == nil || *bcVersion < core.BlockChainVersion {
			if bcVersion != nil { // only print warning on upgrade, not on init
				log.Warn("Upgrade blockchain database version", "from", dbVer, "to", core.BlockChainVersion)
			}
			rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
		}
	}
	var (
		vmConfig = vm.Config{
			EnablePreimageRecording: config.EnablePreimageRecording,
		}
		cacheConfig = &core.CacheConfig{
			TrieCleanLimit:      config.TrieCleanCache,
			TrieCleanNoPrefetch: config.NoPrefetch,
			TrieDirtyLimit:      config.TrieDirtyCache,
			TrieDirtyDisabled:   config.NoPruning,
			TrieTimeLimit:       config.TrieTimeout,
			SnapshotLimit:       config.SnapshotCache,
			Preimages:           config.Preimages,
			StateHistory:        config.StateHistory,
			StateScheme:         config.StateScheme,
		}
	)
	// Override the chain config with provided settings.
	zond.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, config.Genesis, zond.engine, vmConfig, zond.shouldPreserve, &config.TransactionHistory)
	if err != nil {
		return nil, err
	}
	zond.bloomIndexer.Start(zond.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = stack.ResolvePath(config.TxPool.Journal)
	}
	legacyPool := legacypool.New(config.TxPool, zond.blockchain)

	zond.txPool, err = txpool.New(new(big.Int).SetUint64(config.TxPool.PriceLimit), zond.blockchain, []txpool.SubPool{legacyPool})
	if err != nil {
		return nil, err
	}
	// Permit the downloader to use the trie cache allowance during fast sync
	cacheLimit := cacheConfig.TrieCleanLimit + cacheConfig.TrieDirtyLimit + cacheConfig.SnapshotLimit
	if zond.handler, err = newHandler(&handlerConfig{
		Database:       chainDb,
		Chain:          zond.blockchain,
		TxPool:         zond.txPool,
		Network:        config.NetworkId,
		Sync:           config.SyncMode,
		BloomCache:     uint64(cacheLimit),
		EventMux:       zond.eventMux,
		RequiredBlocks: config.RequiredBlocks,
	}); err != nil {
		return nil, err
	}

	zond.miner = miner.New(zond, &config.Miner, zond.blockchain.Config(), zond.EventMux(), zond.engine, zond.isLocalBlock)
	zond.miner.SetExtra(makeExtraData(config.Miner.ExtraData))

	zond.APIBackend = &ZondAPIBackend{stack.Config().ExtRPCEnabled(), zond, nil}

	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.Miner.GasPrice
	}
	zond.APIBackend.gpo = gasprice.NewOracle(zond.APIBackend, gpoParams)

	// Setup DNS discovery iterators.
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	zond.ethDialCandidates, err = dnsclient.NewIterator(zond.config.ZondDiscoveryURLs...)
	if err != nil {
		return nil, err
	}
	zond.snapDialCandidates, err = dnsclient.NewIterator(zond.config.SnapDiscoveryURLs...)
	if err != nil {
		return nil, err
	}

	// Start the RPC service
	zond.netRPCService = zondapi.NewNetAPI(zond.p2pServer, config.NetworkId)

	// Register the backend on the node
	stack.RegisterAPIs(zond.APIs())
	stack.RegisterProtocols(zond.Protocols())
	stack.RegisterLifecycle(zond)

	// Successful startup; push a marker and check previous unclean shutdowns.
	zond.shutdownTracker.MarkStartup()

	return zond, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gzond",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// APIs return the collection of RPC services the zond package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Zond) APIs() []rpc.API {
	apis := zondapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "zond",
			Service:   NewZondAPI(s),
		}, {
			Namespace: "miner",
			Service:   NewMinerAPI(s),
		}, {
			Namespace: "zond",
			Service:   downloader.NewDownloaderAPI(s.handler.downloader, s.eventMux),
		}, {
			Namespace: "admin",
			Service:   NewAdminAPI(s),
		}, {
			Namespace: "debug",
			Service:   NewDebugAPI(s),
		}, {
			Namespace: "net",
			Service:   s.netRPCService,
		},
	}...)
}

func (s *Zond) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Zond) Etherbase() (eb common.Address, err error) {
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()

	if etherbase != (common.Address{}) {
		return etherbase, nil
	}
	return common.Address{}, errors.New("etherbase must be explicitly specified")
}

// isLocalBlock checks whether the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: etherbase
// and accounts specified via `txpool.locals` flag.
func (s *Zond) isLocalBlock(header *types.Header) bool {
	author, err := s.engine.Author(header)
	if err != nil {
		log.Warn("Failed to retrieve block author", "number", header.Number.Uint64(), "hash", header.Hash(), "err", err)
		return false
	}
	// Check whether the given address is etherbase.
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()
	if author == etherbase {
		return true
	}
	// Check whether the given address is specified by `txpool.local`
	// CLI flag.
	for _, account := range s.config.TxPool.Locals {
		if account == author {
			return true
		}
	}
	return false
}

// shouldPreserve checks whether we should preserve the given block
// during the chain reorg depending on whether the author of block
// is a local account.
func (zond *Zond) shouldPreserve(header *types.Header) bool {
	return zond.isLocalBlock(header)
}

// SetEtherbase sets the mining reward address.
func (zond *Zond) SetEtherbase(etherbase common.Address) {
	zond.lock.Lock()
	zond.etherbase = etherbase
	zond.lock.Unlock()

	zond.miner.SetEtherbase(etherbase)
}

func (s *Zond) IsMining() bool      { return s.miner.Mining() }
func (s *Zond) Miner() *miner.Miner { return s.miner }

func (s *Zond) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Zond) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Zond) TxPool() *txpool.TxPool             { return s.txPool }
func (s *Zond) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Zond) Engine() consensus.Engine           { return s.engine }
func (s *Zond) ChainDb() zonddb.Database           { return s.chainDb }
func (s *Zond) IsListening() bool                  { return true } // Always listening
func (s *Zond) Downloader() *downloader.Downloader { return s.handler.downloader }
func (s *Zond) Synced() bool                       { return s.handler.acceptTxs.Load() }
func (s *Zond) SetSynced()                         { s.handler.enableSyncedFeatures() }
func (s *Zond) ArchiveMode() bool                  { return s.config.NoPruning }
func (s *Zond) BloomIndexer() *core.ChainIndexer   { return s.bloomIndexer }

// Protocols returns all the currently configured
// network protocols to start.
func (s *Zond) Protocols() []p2p.Protocol {
	protos := zond.MakeProtocols((*zondHandler)(s.handler), s.networkID, s.ethDialCandidates)
	if s.config.SnapshotCache > 0 {
		protos = append(protos, snap.MakeProtocols((*snapHandler)(s.handler), s.snapDialCandidates)...)
	}
	return protos
}

// Start implements node.Lifecycle, starting all internal goroutines needed by the
// Zond protocol implementation.
func (s *Zond) Start() error {
	zond.StartENRUpdater(s.blockchain, s.p2pServer.LocalNode())

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers(params.BloomBitsBlocks)

	// Regularly update shutdown marker
	s.shutdownTracker.Start()

	// Figure out a max peers count based on the server limits
	maxPeers := s.p2pServer.MaxPeers

	// Start the networking layer if requested
	s.handler.Start(maxPeers)
	return nil
}

// Stop implements node.Lifecycle, terminating all internal goroutines used by the
// Zond protocol.
func (s *Zond) Stop() error {
	// Stop all the peer-related stuff first.
	s.ethDialCandidates.Close()
	s.snapDialCandidates.Close()
	s.handler.Stop()

	// Then stop everything else.
	s.bloomIndexer.Close()
	close(s.closeBloomHandler)
	s.txPool.Close()
	s.miner.Close()
	s.blockchain.Stop()
	s.engine.Close()

	// Clean shutdown marker as the last thing before closing db
	s.shutdownTracker.Stop()

	s.chainDb.Close()
	s.eventMux.Stop()

	return nil
}
