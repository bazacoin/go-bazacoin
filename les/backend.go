// Copyright 2016 The go-bazacoin Authors
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

// Package les implements the Light Bazacoin Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/bazacoin/go-bazacoin/accounts"
	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/common/hexutil"
	"github.com/bazacoin/go-bazacoin/consensus"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/core/types"
	"github.com/bazacoin/go-bazacoin/bzc"
	"github.com/bazacoin/go-bazacoin/bzc/downloader"
	"github.com/bazacoin/go-bazacoin/bzc/filters"
	"github.com/bazacoin/go-bazacoin/bzc/gasprice"
	"github.com/bazacoin/go-bazacoin/bzcdb"
	"github.com/bazacoin/go-bazacoin/event"
	"github.com/bazacoin/go-bazacoin/internal/bzcapi"
	"github.com/bazacoin/go-bazacoin/light"
	"github.com/bazacoin/go-bazacoin/log"
	"github.com/bazacoin/go-bazacoin/node"
	"github.com/bazacoin/go-bazacoin/p2p"
	"github.com/bazacoin/go-bazacoin/p2p/discv5"
	"github.com/bazacoin/go-bazacoin/params"
	rpc "github.com/bazacoin/go-bazacoin/rpc"
)

type LightBazacoin struct {
	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb bzcdb.Database // Block chain database

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *bzcapi.PublicNetAPI

	quitSync chan struct{}
	wg       sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *bzc.Config) (*LightBazacoin, error) {
	chainDb, err := bzc.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	bzc := &LightBazacoin{
		chainConfig:    chainConfig,
		chainDb:        chainDb,
		eventMux:       ctx.EventMux,
		peers:          peers,
		reqDist:        newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		engine:         bzc.CreateConsensusEngine(ctx, config, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
	}

	bzc.relay = NewLesTxRelay(peers, bzc.reqDist)
	bzc.serverPool = newServerPool(chainDb, quitSync, &bzc.wg)
	bzc.retriever = newRetrieveManager(peers, bzc.reqDist, bzc.serverPool)
	bzc.odr = NewLesOdr(chainDb, bzc.retriever)
	if bzc.blockchain, err = light.NewLightChain(bzc.odr, bzc.chainConfig, bzc.engine, bzc.eventMux); err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		bzc.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	bzc.txPool = light.NewTxPool(bzc.chainConfig, bzc.eventMux, bzc.blockchain, bzc.relay)
	if bzc.protocolManager, err = NewProtocolManager(bzc.chainConfig, true, config.NetworkId, bzc.eventMux, bzc.engine, bzc.peers, bzc.blockchain, nil, chainDb, bzc.odr, bzc.relay, quitSync, &bzc.wg); err != nil {
		return nil, err
	}
	bzc.ApiBackend = &LesApiBackend{bzc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	bzc.ApiBackend.gpo = gasprice.NewOracle(bzc.ApiBackend, gpoParams)
	return bzc, nil
}

func lesTopic(genesisHash common.Hash) discv5.Topic {
	return discv5.Topic("LES@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Bazacoinbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Bazacoinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Bazacoinbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the bazacoin package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightBazacoin) APIs() []rpc.API {
	return append(bzcapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "bzc",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "bzc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "bzc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightBazacoin) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightBazacoin) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightBazacoin) TxPool() *light.TxPool              { return s.txPool }
func (s *LightBazacoin) Engine() consensus.Engine           { return s.engine }
func (s *LightBazacoin) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightBazacoin) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightBazacoin) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightBazacoin) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Bazacoin protocol implementation.
func (s *LightBazacoin) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = bzcapi.NewPublicNetAPI(srvr, s.networkId)
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash()))
	s.protocolManager.Start()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Bazacoin protocol.
func (s *LightBazacoin) Stop() error {
	s.odr.Stop()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
