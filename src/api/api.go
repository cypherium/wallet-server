/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package api

import (
	"github.com/cypherium/cph-service/src/api/block_query"
	"github.com/cypherium/cph-service/src/api/block_query/block_number"
	"github.com/cypherium/cph-service/src/api/block_query/get_block_by_height"
	"github.com/cypherium/cph-service/src/api/mining"
	"github.com/cypherium/cph-service/src/api/mining/get_mined_block_by_addr_and_date"
	"github.com/cypherium/cph-service/src/api/poc/get_balance"
	"github.com/cypherium/cph-service/src/api/poc/get_exchange_rate"
	"github.com/cypherium/cph-service/src/api/poc/get_summary"
	"github.com/cypherium/cph-service/src/api/transaction"
	"github.com/cypherium/cph-service/src/api/transaction/get_addr_pending"
	"github.com/cypherium/cph-service/src/api/transaction/get_hash_pending"
	"github.com/cypherium/cph-service/src/api/transaction/get_transaction_by_hash"
)

var (
	//Login    = login.Main
	//LogOut   = logout.Main

	//block
	GetBlockNumber   = block_number.Main
	GetBlockByHeight = get_block_by_height.Main
	GetBlocks        = block_query.Get_Blocks
	GetBlockByHash   = block_query.Get_by_hash

	//transaction
	GetTransactionByHash = get_transaction_by_hash.Main
	GetByAddr            = transaction.Get_by_addr
	GetByAddrAndType     = transaction.Get_by_addr_and_type
	GetByHeight          = transaction.Get_by_height
	GetTransactions      = transaction.Get_transactions
	GetAddrPending       = get_addr_pending.Main
	GetHashPending       = get_hash_pending.Main

	//mining
	GetMinedBlocks             = mining.Get_mined_block_by_addr
	GetAddrMiningRewards       = mining.Main
	GetMinedblockByAddrAndDate = get_mined_block_by_addr_and_date.Main

	//poc
	GetExchangeRate = get_exchange_rate.Main
	GetBalance      = get_balance.Main
	GetSummary      = get_summary.Main
)
