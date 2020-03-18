// Copyright 2015 The go-cypherium Authors
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

// Package cphapi implements the general Cypherium API functions.
package cphapi

import (
	"context"
	"math/big"

	"github.com/cypherium/go-cypherium/accounts"
	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/core"
	"github.com/cypherium/go-cypherium/core/state"
	"github.com/cypherium/go-cypherium/core/types"
	"github.com/cypherium/go-cypherium/core/vm"
	"github.com/cypherium/go-cypherium/cph/downloader"
	"github.com/cypherium/go-cypherium/cphdb"
	"github.com/cypherium/go-cypherium/event"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/rpc"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Cypherium API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() cphdb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager
	CandidatePool() *core.CandidatePool

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	KeyBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.KeyBlock, error)
	KeyBlockByHash(ctx context.Context, blockHash common.Hash) (*types.KeyBlock, error)
	RollbackKeyChainFrom(lockHash common.Hash) error
	RollbackTxChainFrom(lockHash common.Hash) error
	CommitteeMembers(ctx context.Context, blockNr rpc.BlockNumber) ([]*common.Cnode, error)
	MockKeyBlock(int64)
	GetKeyBlockChain() *core.KeyBlockChain
	AnnounceBlock(blockNr rpc.BlockNumber)
	KeyBlockNumber() uint64
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription

	// TxPool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Block
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "cph",
			Version:   "1.0",
			Service:   NewPublicCphereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "cph",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "cph",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "cph",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPublicPowCandidateAPI(apiBackend),
			Public:    true,
		},
	}
}
