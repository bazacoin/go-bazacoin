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

package les

import (
	"context"
	"math/big"

	"github.com/bazacoin/go-bazacoin/accounts"
	"github.com/bazacoin/go-bazacoin/common"
	"github.com/bazacoin/go-bazacoin/common/math"
	"github.com/bazacoin/go-bazacoin/core"
	"github.com/bazacoin/go-bazacoin/core/types"
	"github.com/bazacoin/go-bazacoin/core/vm"
	"github.com/bazacoin/go-bazacoin/bzc/downloader"
	"github.com/bazacoin/go-bazacoin/bzc/gasprice"
	"github.com/bazacoin/go-bazacoin/bzcdb"
	"github.com/bazacoin/go-bazacoin/event"
	"github.com/bazacoin/go-bazacoin/internal/bzcapi"
	"github.com/bazacoin/go-bazacoin/light"
	"github.com/bazacoin/go-bazacoin/params"
	"github.com/bazacoin/go-bazacoin/rpc"
)

type LesApiBackend struct {
	bzc *LightBazacoin
	gpo *gasprice.Oracle
}

func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.bzc.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.bzc.BlockChain().CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.bzc.protocolManager.downloader.Cancel()
	b.bzc.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.bzc.blockchain.CurrentHeader(), nil
	}

	return b.bzc.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (bzcapi.State, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewLightState(light.StateTrieID(header), b.bzc.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.bzc.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return light.GetBlockReceipts(ctx, b.bzc.odr, blockHash, core.GetBlockNumber(b.bzc.chainDb, blockHash))
}

func (b *LesApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.bzc.blockchain.GetTdByHash(blockHash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state bzcapi.State, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	stateDb := state.(*light.LightState).Copy()
	addr := msg.From()
	from, err := stateDb.GetOrNewStateObject(ctx, addr)
	if err != nil {
		return nil, nil, err
	}
	from.SetBalance(math.MaxBig256)

	vmstate := light.NewVMState(ctx, stateDb)
	context := core.NewEVMContext(msg, header, b.bzc.blockchain, nil)
	return vm.NewEVM(context, vmstate, b.bzc.chainConfig, vmCfg), vmstate.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.bzc.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.bzc.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.bzc.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.bzc.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.bzc.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.bzc.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.bzc.txPool.Content()
}

func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return b.bzc.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.bzc.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() bzcdb.Database {
	return b.bzc.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.bzc.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.bzc.accountManager
}
