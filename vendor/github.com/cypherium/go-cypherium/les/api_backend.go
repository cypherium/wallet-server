// Copyright 2016 The go-cypherium Authors
// This file is part of the cypherBFT library.
//
// The cypherBFT library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The cypherBFT library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the cypherBFT library. If not, see <http://www.gnu.org/licenses/>.

package les

import (
	"context"
	"math/big"

	"github.com/cypherium/go-cypherium/accounts"
	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/common/math"
	"github.com/cypherium/go-cypherium/core"
	"github.com/cypherium/go-cypherium/core/bloombits"
	"github.com/cypherium/go-cypherium/core/rawdb"
	"github.com/cypherium/go-cypherium/core/state"
	"github.com/cypherium/go-cypherium/core/types"
	"github.com/cypherium/go-cypherium/core/vm"
	"github.com/cypherium/go-cypherium/cph/downloader"
	"github.com/cypherium/go-cypherium/cph/gasprice"
	"github.com/cypherium/go-cypherium/cphdb"
	"github.com/cypherium/go-cypherium/event"
	"github.com/cypherium/go-cypherium/light"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/rpc"
	"github.com/pkg/errors"
)

type LesApiBackend struct {
	cph *LightCphereum
	gpo *gasprice.Oracle
}

func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.cph.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.cph.BlockChain().CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.cph.protocolManager.downloader.Cancel()
	b.cph.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.cph.blockchain.CurrentHeader(), nil
	}

	return b.cph.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) KeyBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.KeyBlock, error) {
	return nil, errors.New("To be supported in the future.")
}

func (b *LesApiBackend) KeyBlockByHash(ctx context.Context, blockHash common.Hash) (*types.KeyBlock, error) {
	return nil, errors.New("To be supported in the future.")
}

func (b *LesApiBackend) CommitteeMembers(ctx context.Context, blockNr rpc.BlockNumber) ([]*common.Cnode, error) {
	return nil, nil
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewState(ctx, header, b.cph.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.cph.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.cph.chainDb, hash); number != nil {
		return light.GetBlockReceipts(ctx, b.cph.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	if number := rawdb.ReadHeaderNumber(b.cph.chainDb, hash); number != nil {
		return light.GetBlockLogs(ctx, b.cph.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetTd(hash common.Hash) *big.Int {
	return b.cph.blockchain.GetTdByHash(hash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	context := core.NewEVMContext(msg, header, b.cph.blockchain)
	evm := vm.NewEVM(context, state, b.cph.chainConfig, vmCfg, b.cph.blockchain)
	return evm, state.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.cph.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.cph.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.cph.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.cph.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.cph.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.cph.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.cph.txPool.Content()
}

func (b *LesApiBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.cph.txPool.SubscribeNewTxsEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.cph.blockchain.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.cph.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.cph.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.cph.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return b.cph.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.cph.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() cphdb.Database {
	return b.cph.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.cph.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.cph.accountManager
}

func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
	if b.cph.bloomIndexer == nil {
		return 0, 0
	}
	sections, _, _ := b.cph.bloomIndexer.Sections()
	return light.BloomTrieFrequency, sections
}

func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.cph.bloomRequests)
	}
}

func (b *LesApiBackend) CandidatePool() *core.CandidatePool {
	return nil
}
