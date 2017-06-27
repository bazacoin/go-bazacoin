// Copyright 2015 The go-bazacoin Authors
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

package bzc

import (
	"context"
	"math/big"

	"github.com/bazacoin/go-bazacoin/accounts"
	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/common/math"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/core/state"
	"github.com/bazacoin/go-bazacoin/core/types"
	"github.com/bazacoin/go-bazacoin/core/vm"
	"github.com/bazacoin/go-bazacoin/bzc/downloader"
	"github.com/bazacoin/go-bazacoin/bzc/gasprice"
	"github.com/bazacoin/go-bazacoin/bzcdb"
	"github.com/bazacoin/go-bazacoin/event"
	"github.com/bazacoin/go-bazacoin/internal/bzcapi"
	"github.com/bazacoin/go-bazacoin/params"
	"github.com/bazacoin/go-bazacoin/rpc"
)

// BzcApiBackend implements bzcapi.Backend for full nodes
type BzcApiBackend struct {
	bzc *Bazacoin
	gpo *gasprice.Oracle
}

func (b *BzcApiBackend) ChainConfig() *params.ChainConfig {
	return b.bzc.chainConfig
}

func (b *BzcApiBackend) CurrentBlock() *types.Block {
	return b.bzc.blockchain.CurrentBlock()
}

func (b *BzcApiBackend) SetHead(number uint64) {
	b.bzc.protocolManager.downloader.Cancel()
	b.bzc.blockchain.SetHead(number)
}

func (b *BzcApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.bzc.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.bzc.blockchain.CurrentBlock().Header(), nil
	}
	return b.bzc.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *BzcApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.bzc.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.bzc.blockchain.CurrentBlock(), nil
	}
	return b.bzc.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *BzcApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (bzcapi.State, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.bzc.miner.Pending()
		return BzcApiState{state}, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.bzc.BlockChain().StateAt(header.Root)
	return BzcApiState{stateDb}, header, err
}

func (b *BzcApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.bzc.blockchain.GetBlockByHash(blockHash), nil
}

func (b *BzcApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.bzc.chainDb, blockHash, core.GetBlockNumber(b.bzc.chainDb, blockHash)), nil
}

func (b *BzcApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.bzc.blockchain.GetTdByHash(blockHash)
}

func (b *BzcApiBackend) GetEVM(ctx context.Context, msg core.Message, state bzcapi.State, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	statedb := state.(BzcApiState).state
	from := statedb.GetOrNewStateObject(msg.From())
	from.SetBalance(math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.bzc.BlockChain(), nil)
	return vm.NewEVM(context, statedb, b.bzc.chainConfig, vmCfg), vmError, nil
}

func (b *BzcApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	b.bzc.txPool.SetLocal(signedTx)
	return b.bzc.txPool.Add(signedTx)
}

func (b *BzcApiBackend) RemoveTx(txHash common.Hash) {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	b.bzc.txPool.Remove(txHash)
}

func (b *BzcApiBackend) GetPoolTransactions() (types.Transactions, error) {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	pending, err := b.bzc.txPool.Pending()
	if err != nil {
		return nil, err
	}

	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *BzcApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	return b.bzc.txPool.Get(hash)
}

func (b *BzcApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	return b.bzc.txPool.State().GetNonce(addr), nil
}

func (b *BzcApiBackend) Stats() (pending int, queued int) {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	return b.bzc.txPool.Stats()
}

func (b *BzcApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	b.bzc.txMu.Lock()
	defer b.bzc.txMu.Unlock()

	return b.bzc.TxPool().Content()
}

func (b *BzcApiBackend) Downloader() *downloader.Downloader {
	return b.bzc.Downloader()
}

func (b *BzcApiBackend) ProtocolVersion() int {
	return b.bzc.BzcVersion()
}

func (b *BzcApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *BzcApiBackend) ChainDb() bzcdb.Database {
	return b.bzc.ChainDb()
}

func (b *BzcApiBackend) EventMux() *event.TypeMux {
	return b.bzc.EventMux()
}

func (b *BzcApiBackend) AccountManager() *accounts.Manager {
	return b.bzc.AccountManager()
}

type BzcApiState struct {
	state *state.StateDB
}

func (s BzcApiState) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return s.state.GetBalance(addr), nil
}

func (s BzcApiState) GetCode(ctx context.Context, addr common.Address) ([]byte, error) {
	return s.state.GetCode(addr), nil
}

func (s BzcApiState) GetState(ctx context.Context, a common.Address, b common.Hash) (common.Hash, error) {
	return s.state.GetState(a, b), nil
}

func (s BzcApiState) GetNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return s.state.GetNonce(addr), nil
}
