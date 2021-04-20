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
 * @file request-result.go
 * @authors:
 *   Reginaldo Costa <regcostajr@gmail.com>
 * @date 2017
 */

package dto

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cypherium/wallet-server/src/go-web3/complex/types"
	"github.com/cypherium/wallet-server/src/go-web3/constants"

	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cypherium/cypherBFT/go-cypherium/log"
)

type RequestResult struct {
	ID      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error,omitempty"`
	Data    string      `json:"data,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (pointer *RequestResult) ToStringArray() ([]string, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.([]interface{})

	new := make([]string, len(result))
	for i, v := range result {
		new[i] = v.(string)
	}

	return new, nil

}

func (pointer *RequestResult) ToComplexString() (types.ComplexString, error) {

	if err := pointer.checkResponse(); err != nil {
		return "", err
	}

	result := (pointer).Result.(interface{})

	return types.ComplexString(result.(string)), nil

}

func (pointer *RequestResult) ToString() (string, error) {

	if err := pointer.checkResponse(); err != nil {
		return "", err
	}

	result := (pointer).Result.(interface{})

	return result.(string), nil

}

func (pointer *RequestResult) ToInt() (int64, error) {

	if err := pointer.checkResponse(); err != nil {
		return 0, err
	}

	result := (pointer).Result.(interface{})

	hex := result.(string)

	numericResult, err := strconv.ParseInt(hex, 16, 64)

	return numericResult, err

}

func (pointer *RequestResult) ToFloat() (float64, error) {

	if err := pointer.checkResponse(); err != nil {
		return 0, err
	}

	result := (pointer).Result.(interface{})

	numericResult, err := result.(float64)
	if !err {
		return 0, fmt.Errorf("can not convert result to float64")
	}

	return numericResult, nil

}

func (pointer *RequestResult) ToBigInt() (*big.Int, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	res := (pointer).Result.(interface{})

	ret, success := big.NewInt(0).SetString(res.(string)[2:], 16)

	if !success {
		return nil, errors.New(fmt.Sprintf("Failed to convert %s to BigInt", res.(string)))
	}

	return ret, nil
}

func (pointer *RequestResult) ToComplexIntResponse() (types.ComplexIntResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return types.ComplexIntResponse(0), err
	}

	result := (pointer).Result.(interface{})

	var hex string

	switch v := result.(type) {
	//Testrpc returns a float64
	case float64:
		hex = strconv.FormatFloat(v, 'E', 16, 64)
		break
	default:
		hex = result.(string)
	}

	cleaned := strings.TrimPrefix(hex, "0x")

	return types.ComplexIntResponse(cleaned), nil

}

func (pointer *RequestResult) ToBoolean() (bool, error) {

	if err := pointer.checkResponse(); err != nil {
		return false, err
	}

	result := (pointer).Result.(interface{})

	return result.(bool), nil

}

func (pointer *RequestResult) ToSignTransactionResponse() (*SignTransactionResponse, error) {
	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	signTransactionResponse := &SignTransactionResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), signTransactionResponse)

	return signTransactionResponse, err
}

func (pointer *RequestResult) ToTransactionResponse() (*TransactionResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	transactionResponse := &TransactionResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), transactionResponse)

	return transactionResponse, err

}

func (pointer *RequestResult) ToTransactionReceipt() (*TransactionReceipt, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	transactionReceipt := &TransactionReceipt{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), transactionReceipt)

	return transactionReceipt, err

}

func (pointer *RequestResult) ToBlock() (*Block, error) {

	if err := pointer.checkResponse(); err != nil {
		log.Info("checkResponse", "error", err.Error())
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		log.Info("checkResponse 1", "error", customerror.EMPTYRESPONSE.Error())
		return nil, customerror.EMPTYRESPONSE
	}

	preblock := &preBlock{}
	log.Info("Marshal", "result", result)

	marshal, err := json.Marshal(result)
	if err != nil {
		log.Info("Marshal", "error", err.Error())
		return nil, customerror.UNPARSEABLEINTERFACE
	}
	json.Unmarshal(marshal, &preblock)
	//if err := json.Unmarshal(marshal, &preblock); err != nil {
	//	log.Info("Unmarshal", "error", err,"block",preblock)
	//
	//	return nil,err
	//}
	log.Info("Unmarshal ok", "block", preblock)
	num, success := big.NewInt(0).SetString(preblock.Number[2:], 16)
	if !success {
		return nil, errors.New(fmt.Sprintf("Error converting %s to bigInt", preblock.Number))
	}

	timestamp, success := big.NewInt(0).SetString(preblock.Timestamp[2:], 16)
	if !success {
		return nil, errors.New(fmt.Sprintf("Error converting %s to bigInt", preblock.Timestamp))
	}

	size, success := big.NewInt(0).SetString(preblock.Size[2:], 16)
	if !success {
		return nil, errors.New(fmt.Sprintf("Error converting %s to bigInt", preblock.Size))
	}

	gasUsed, success := big.NewInt(0).SetString(preblock.GasUsed[2:], 16)
	if !success {
		return nil, errors.New(fmt.Sprintf("Error converting %s to bigInt", preblock.GasUsed))
	}

	gasLimit, success := big.NewInt(0).SetString(preblock.GasLimit[2:], 16)
	if !success {
		return nil, errors.New(fmt.Sprintf("Error converting %s to bigInt", preblock.GasLimit))
	}

	block := &Block{
		Number:       num,
		Timestamp:    timestamp,
		Transactions: preblock.Transactions,
		Hash:         preblock.Hash,
		ParentHash:   preblock.ParentHash,
		Size:         size,
		GasUsed:      gasUsed,
		GasLimit:     gasLimit,
		ExtraData:    preblock.ExtraData,
		Root:         preblock.Root,
		TxHash:       preblock.TxHash,
		ReceiptHash:  preblock.ReceiptHash,
		BlockType:    preblock.BlockType,
		KeyHash:      preblock.KeyHash,
		Exceptions:   preblock.Exceptions,
	}
	return block, err
}

func (pointer *RequestResult) ToPoc() (*Poc, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	cph := &Poc{}

	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), cph)

	return cph, err

}

func (pointer *RequestResult) ToSyncingResponse() (*SyncingResponse, error) {

	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	var result map[string]interface{}

	switch (pointer).Result.(type) {
	case bool:
		return &SyncingResponse{}, nil
	case map[string]interface{}:
		result = (pointer).Result.(map[string]interface{})
	default:
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	syncingResponse := &SyncingResponse{}

	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	json.Unmarshal([]byte(marshal), syncingResponse)

	return syncingResponse, nil

}

// To avoid a conversion of a nil interface
func (pointer *RequestResult) checkResponse() error {

	if pointer.Error != nil {
		return errors.New(pointer.Error.Message)
	}

	if pointer.Result == nil {
		return customerror.EMPTYRESPONSE
	}

	return nil

}

func (pointer *RequestResult) ToContent() (*Content, error) {
	if err := pointer.checkResponse(); err != nil {
		return nil, err
	}

	result := (pointer).Result.(map[string]interface{})
	if len(result) == 0 {
		return nil, customerror.EMPTYRESPONSE
	}

	content := &Content{}
	marshal, err := json.Marshal(result)

	if err != nil {
		return nil, customerror.UNPARSEABLEINTERFACE
	}

	err = json.Unmarshal([]byte(marshal), content)

	return content, err
}
