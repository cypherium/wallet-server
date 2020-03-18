// Copyright 2014 The go-cypherium Authors
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

package core

import (
	"math/big"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/core/types"
)

// NewTxsEvent is posted when a batch of transactions enter the transaction pool.
type NewTxsEvent struct{ Txs []*types.Transaction }

// PendingLogsEvent is posted pre mining and notifies of pending logs.
type PendingLogsEvent struct {
	Logs []*types.Log
}

// PendingStateEvent is posted pre mining and notifies of pending state changes.
type PendingStateEvent struct{}

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *types.Block }

type NewCandidateEvent struct {
	Candidate *types.Candidate
}

type RosterConfigEvent struct {
	RosterConfigData interface{}
}

type RosterConfigDoneEvent struct {
	IsMember bool
}

type CanStartChangeEvent struct {
	CanStartValue int32
}

type LeaderIdentityProofEvent struct {
}

// RemovedLogsEvent is posted when a reorg happens
type RemovedLogsEvent struct{ Logs []*types.Log }

type ChainEvent struct {
	Block *types.Block
	Hash  common.Hash
	Logs  []*types.Log
}

type ChainHeadEvent struct{ Block *types.Block }

type KeyChainHeadEvent struct {
	KeyBlock *types.KeyBlock
}

type ClearObsoleteEvent struct {
	Number *big.Int
}

type IdentityChangeEvent struct {
	Identity int
}

type SelfRoleChangeEvent struct {
	IsMember bool
}

type ResetRoleChangeEvent struct {
	NodeRole uint8
}
