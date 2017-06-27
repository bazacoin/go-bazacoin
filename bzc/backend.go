// Copyright 2014 The go-bazacoin Authors
// This file is part of the go-bazacoin library.
//
// The go-bazacoin library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-bazacoin library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-bazacoin library. If not, see <http://www.gnu.org/licenses/>.

// Package bzc implements the Bazacoin protocol.
package bzc

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/bazacoin/go-bazacoin/accounts"
	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/common/hexutil"
	"github.com/bazacoin/go-bazacoin/consensus"
	"github.com/bazacoin/go-bazacoin/consensus/clique"
	"github.com/bazacoin/go-bazacoin/consensus/bzhash"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/core/types"
	"github.com/bazacoin/go-bazacoin/core/vm"
	"github.com/bazacoin/go-bazacoin/bzc/downloader"
	"github.com/bazacoin/go-bazacoin/bzc/filters"
	"github.com/bazacoin/go-bazacoin/bzc/gasprice"
	"github.com/bazacoin/go-bazacoin/bzcdb"
	"github.com/bazacoin/go-bazacoin/event"
	"github.com/bazacoin/go-bazacoin/internal/bzcapi"
	"github.com/bazacoin/go-bazacoin/log"
	"github.com/bazacoin/go-bazacoin/miner"
	"github.com/bazacoin/go-bazacoin/node"
	"github.com/bazacoin/go-bazacoin/p2p"
	"github.com/bazacoin/go-bazacoin/params"
	"github.com/bazacoin/go-bazacoin/rlp"
	"github.com/bazacoin/go-bazacoin/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
}

// Bazacoin implements the Bazacoin full node service.
type Bazacoin struct {
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan  chan bool // Channel for shutting down the bazacoin
	stopDbUpgrade func()    // stop chain db sequential key upgrade
	// Handlers
	txPool          *core.TxPool
	txMu            sync.Mutex
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer
	// DB interfaces
	chainDb bzcdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	ApiBackend *BzcApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	bazacoinbase common.Address

	networkId     uint64
	netRPCService *bzcapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and bazacoinbase)
}

func (s *Bazacoin) AddLesServer(ls LesServer) {
	s.lesServer = ls
}

// New creates a new Bazacoin object (including the
// initialisation of the common Bazacoin object)
func New(ctx *node.ServiceContext, config *Config) (*Bazacoin, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run bzc.Bazacoin in light sync mode, use les.LightBazacoin")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}

	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeSequentialKeys(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	bzc := &Bazacoin{
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		bazacoinbase:      config.Bazacoinbase,
	}

	if err := addMipmapBloomBins(chainDb); err != nil {
		return nil, err
	}
	log.Info("Initialising Bazacoin protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run geth upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	bzc.blockchain, err = core.NewBlockChain(chainDb, bzc.chainConfig, bzc.engine, bzc.eventMux, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		bzc.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	newPool := core.NewTxPool(config.TxPool, bzc.chainConfig, bzc.EventMux(), bzc.blockchain.State, bzc.blockchain.GasLimit)
	bzc.txPool = newPool

	maxPeers := config.MaxPeers
	if config.LightServ > 0 {
		// if we are running a light server, limit the number of BZC peers so that we reserve some space for incoming LES connections
		// temporary solution until the new peer connectivity API is finished
		halfPeers := maxPeers / 2
		maxPeers -= config.LightPeers
		if maxPeers < halfPeers {
			maxPeers = halfPeers
		}
	}

	if bzc.protocolManager, err = NewProtocolManager(bzc.chainConfig, config.SyncMode, config.NetworkId, maxPeers, bzc.eventMux, bzc.txPool, bzc.engine, bzc.blockchain, chainDb); err != nil {
		return nil, err
	}

	bzc.miner = miner.New(bzc, bzc.chainConfig, bzc.EventMux(), bzc.engine)
	bzc.miner.SetExtra(makeExtraData(config.ExtraData))

	bzc.ApiBackend = &BzcApiBackend{bzc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	bzc.ApiBackend.gpo = gasprice.NewOracle(bzc.ApiBackend, gpoParams)

	return bzc, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"geth",
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

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (bzcdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*bzcdb.LDBDatabase); ok {
		db.Meter("bzc/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Bazacoin service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db bzcdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowFake:
		log.Warn("Bzhash used in fake mode")
		return bzhash.NewFaker()
	case config.PowTest:
		log.Warn("Bzhash used in test mode")
		return bzhash.NewTester()
	case config.PowShared:
		log.Warn("Bzhash used in shared mode")
		return bzhash.NewShared()
	default:
		engine := bzhash.New(ctx.ResolvePath(config.BzhashCacheDir), config.BzhashCachesInMem, config.BzhashCachesOnDisk,
			config.BzhashDatasetDir, config.BzhashDatasetsInMem, config.BzhashDatasetsOnDisk)
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the bazacoin package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Bazacoin) APIs() []rpc.API {
	apis := bzcapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "bzc",
			Version:   "1.0",
			Service:   NewPublicBazacoinAPI(s),
			Public:    true,
		}, {
			Namespace: "bzc",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "bzc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "bzc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Bazacoin) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Bazacoin) Bazacoinbase() (eb common.Address, err error) {
	s.lock.RLock()
	bazacoinbase := s.bazacoinbase
	s.lock.RUnlock()

	if bazacoinbase != (common.Address{}) {
		return bazacoinbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			return accounts[0].Address, nil
		}
	}
	return common.Address{}, fmt.Errorf("bazacoinbase address must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Bazacoin) SetBazacoinbase(bazacoinbase common.Address) {
	self.lock.Lock()
	self.bazacoinbase = bazacoinbase
	self.lock.Unlock()

	self.miner.SetBazacoinbase(bazacoinbase)
}

func (s *Bazacoin) StartMining(local bool) error {
	eb, err := s.Bazacoinbase()
	if err != nil {
		log.Error("Cannot start mining without bazacoinbase", "err", err)
		return fmt.Errorf("bazacoinbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Bazacoinbase account unavailable locally", "err", err)
			return fmt.Errorf("singer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Bazacoin) StopMining()         { s.miner.Stop() }
func (s *Bazacoin) IsMining() bool      { return s.miner.Mining() }
func (s *Bazacoin) Miner() *miner.Miner { return s.miner }

func (s *Bazacoin) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Bazacoin) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Bazacoin) TxPool() *core.TxPool               { return s.txPool }
func (s *Bazacoin) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Bazacoin) Engine() consensus.Engine           { return s.engine }
func (s *Bazacoin) ChainDb() bzcdb.Database            { return s.chainDb }
func (s *Bazacoin) IsListening() bool                  { return true } // Always listening
func (s *Bazacoin) BzcVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Bazacoin) NetVersion() uint64                 { return s.networkId }
func (s *Bazacoin) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Bazacoin) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	} else {
		return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
	}
}

// Start implements node.Service, starting all internal goroutines needed by the
// Bazacoin protocol implementation.
func (s *Bazacoin) Start(srvr *p2p.Server) error {
	s.netRPCService = bzcapi.NewPublicNetAPI(srvr, s.NetVersion())

	s.protocolManager.Start()
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Bazacoin protocol.
func (s *Bazacoin) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
