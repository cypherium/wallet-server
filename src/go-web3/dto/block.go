/********************************************************************************
   This file is part of go-web3.
   go-web3 is free software: you can redistribute it and/or modify
   it under the terms of the GNU Lesser General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-web3 is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Lesser General Public License for more details.
   You should have received a copy of the GNU Lesser General Public License
   along with go-web3.  If not, see <http://www.gnu.org/licenses/>.
*********************************************************************************/

/**
 * @file block.go
 * @authors:
 *   Jérôme Laurens <jeromelaurens@gmail.com>
 * @date 2017
 */

package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/cypherium/cypherBFT/go-cypherium/common"
	"github.com/cypherium/cypherBFT/go-cypherium/core/types"
	"github.com/cypherium/cypherBFT/go-cypherium/log"
)

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

type transaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	Version      uint64          `json:"version"  gencodec:"required"`
	SenderKey    []byte          `json:"senderKey" gencodec:"required"`
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type Block struct {
	Number       *big.Int              `json:"number"`
	Timestamp    *big.Int              `json:"timestamp"`
	Transactions []TransactionResponse `json:"transactions"` //todo 这里默认使用查询transactions的rps调用方式
	Hash         string                `json:"hash"`
	ParentHash   string                `json:"parentHash"`
	// Miner          string   `json:"miner,omitempty"`
	// MixHash        string   `json:"mixHash"`
	// Difficulty     *big.Int `json:"difficulty"`
	// TotalDifficult *big.Int `json:"totalDifficulty"`
	Size     *big.Int `json:"size"`
	GasUsed  *big.Int `json:"gasUsed"`
	GasLimit *big.Int `json:"gasLimit"`
	// Nonce          *big.Int `json:"nonce"`
	ExtraData string `json:"extraData"`
	// Reward         *big.Int `json:"reward"`
	// TxFees         *big.Int `json:"txfees"`
	//Author         string                `json:"author,omitempty"`
	Root        string `json:"stateRoot"`
	TxHash      string `json:"transactionsRoot"`
	ReceiptHash string `json:"receiptsRoot"`
	BlockType   uint8  `json:"BlockType"`
	KeyHash     string `json:"keyHash"`
	// Signature   []byte `json:"Signature"`
	Exceptions []byte `json:"exceptions"`
	// 	fields["txN"] = len(b.Transactions())
}

type preBlock struct {
	Number       string                `json:"number"`
	Timestamp    string                `json:"timestamp"`
	Transactions []TransactionResponse `json:"transactions"` //todo 这里默认使用查询transactions的rps调用方式
	Hash         string                `json:"hash"`
	ParentHash   string                `json:"parentHash"`
	// Miner          string   `json:"miner,omitempty"`
	// MixHash        string   `json:"mixHash"`
	// Difficulty     *big.Int `json:"difficulty"`
	// TotalDifficult *big.Int `json:"totalDifficulty"`
	Size     string `json:"size"`
	GasUsed  string `json:"gasUsed"`
	GasLimit string `json:"gasLimit"`
	// Nonce          *big.Int `json:"nonce"`
	ExtraData string `json:"extraData"`
	// Reward         *big.Int `json:"reward"`
	// TxFees         *big.Int `json:"txfees"`
	//Author         string                `json:"author,omitempty"`
	Root        string `json:"stateRoot"`
	TxHash      string `json:"transactionsRoot"`
	ReceiptHash string `json:"receiptsRoot"`
	BlockType   uint8  `json:"BlockType"`
	KeyHash     string `json:"keyHash"`
	// Signature   []byte `json:"Signature"`
	Exceptions []byte `json:"exceptions"`
	// 	fields["txN"] = len(b.Transactions())
}

// fields["exceptions"] = hexutil.Bytes(b.Exceptions())
//type Block struct {
//	Number     *big.Int `json:"number"`
//	Hash       string   `json:"hash"`
//	ParentHash string   `json:"parentHash"`
//	Author     string   `json:"author,omitempty"`
//	Miner      string   `json:"miner,omitempty"`
//	Size       *big.Int `json:"size"`
//	GasUsed    *big.Int `json:"gasUsed"`
//	Nonce      *big.Int `json:"nonce"`
//	Timestamp  *big.Int `json:"timestamp"`
//}

/**
 * How to un-marshal the block struct using the Big.Int rather than the
 * `complexReturn` type.
 */
func (b *Block) UnmarshalJSON(data []byte) error {
	type Alias Block
	temp := &struct {
		Number   string `json:"number"`
		Size     string `json:"size"`
		GasUsed  string `json:"gasUsed"`
		GasLimit string `json:"gasLimit"`
		// Nonce           string `json:"nonce"`
		Timestamp string `json:"timestamp"`
		// Difficulty      string `json:"difficulty"`
		// TotalDifficulty string `json:"totalDifficulty"`
		// Reward          string `json:"reward"`
		// TxFees          string `json:"txfees"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		log.Info("UnmarshalJSON", "error", err)
		return err
	}

	num, success := big.NewInt(0).SetString(temp.Number[2:], 16)
	if !success {
		return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Number))
	}

	size, success := big.NewInt(0).SetString(temp.Size[2:], 16)
	if !success {
		return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Size))
	}

	gas, success := big.NewInt(0).SetString(temp.GasUsed[2:], 16)
	if !success {
		return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.GasUsed))
	}
	// nonce, success := big.NewInt(0).SetString(temp.Nonce[2:], 16)

	// if !success {
	// 	return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Nonce))
	// }
	timestamp, success := big.NewInt(0).SetString(temp.Timestamp[2:], 16)

	if !success {
		return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Timestamp))
	}
	// difficult, success := big.NewInt(0).SetString(temp.Difficulty[2:], 16)

	// if !success {
	// 	return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Timestamp))
	// }
	// totaldifficult, success := big.NewInt(0).SetString(temp.TotalDifficulty[2:], 16)

	// if !success {
	// 	return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Timestamp))
	// }
	gaslimit, success := big.NewInt(0).SetString(temp.GasLimit[2:], 16)

	if !success {
		return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.GasLimit))
	}
	// reward, success := big.NewInt(0).SetString(temp.Reward[2:], 16)

	// if !success {
	// 	return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Timestamp))
	// }
	// txfees, success := big.NewInt(0).SetString(temp.TxFees[2:], 16)

	// if !success {
	// 	return errors.New(fmt.Sprintf("Error converting %s to bigInt", temp.Timestamp))
	// }

	b.Number = num
	b.Size = size
	b.GasUsed = gas
	// b.Nonce = nonce
	b.Timestamp = timestamp
	// b.Difficulty = difficult
	// b.TotalDifficult = totaldifficult
	b.GasLimit = gaslimit
	// b.Reward = reward
	// b.TxFees = txfees
	// log.Info("UnmarshalJSON", "block", b)
	return nil
}
