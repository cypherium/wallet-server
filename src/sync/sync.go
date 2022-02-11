package sync

import (
	//"github.com/gomodule/redigo/redis"
	// "qoobing.com/utillib.golang/log"

	"github.com/cypherium/wallet-server/src/api/poc/get_richlist"
	"github.com/cypherium/wallet-server/src/go-web3"
	"github.com/cypherium/wallet-server/src/go-web3/eth/block"
	"github.com/cypherium/wallet-server/src/go-web3/providers"
	"math/big"

	"github.com/cypherium/wallet-server/src/model"

	"errors"
	. "github.com/cypherium/wallet-server/src/const"
	"github.com/cypherium/wallet-server/src/go-web3/dto"
	"github.com/cypherium/wallet-server/src/util"
	"os"
	// "qoobing.com/utillib.golang/gls"
	"github.com/cypherium/cypherBFT/log"
	"github.com/cypherium/wallet-server/src/config"
	"strings"
	"time"
)

var logid string
var c = new(Connect)
var GLastBlock *dto.Block = nil

func Init() {
	blockNumber, err := c.Web3().Eth.GetBlockNumber()
	if err != nil {
		log.Error("Init", "error", err)
	}
	richRecord := &model.RichRecord{F_address: "", F_balance: 0}
	for _, account := range get_richlist.GenesisAccounts {

		webthree := web3.NewWeb3(providers.NewHTTPProvider(config.Config().Gate, config.Config().TimeOut.RPCTimeOut, false))
		balance, err := webthree.Eth.GetBalance(account, block.LATEST)
		if err != nil {
			log.Error("GetBalance failed", "balance", balance, "BlockNumber", blockNumber.String(), "error", err.Error())
		} else {

			richRecord.F_address = account
			richRecord.F_balance = (balance.Div(balance, big.NewInt(1e18))).Uint64()
			richRecord.UpdateRichRecord(c.Mysql())
			log.Info("Init", "account", account, "balance", balance)
		}
	}
}

func StartSyncLastBlock() {

	defer c.Close()

	logid = "sync" + util.GetRandomCharacter(4)
	// gls.SetGlsValue("logid", logid)

	max_block, err := (&model.Block{}).GetMaxBlocNumber(c.Mysql())
	if err != nil && err.Error() != DATA_NOT_EXIST {
		log.Error("GetMaxBlocNumber", "error", err.Error())
		os.Exit(0)
	}

	c.SetBlockNow(max_block - 1)
	log.Info("Find databases sync block height, start sync from there.", "height", max_block-1)

	for {
		msg := make(chan int)
		go func() {

			defer func() { msg <- 1 }()
			defer c.PanicRecover()

			for {
				// gls.SetGlsValue("logid", logid+util.GetRandomCharacter(4))
				//var c = new(Connect)
				//defer c.Close()

				blockNumber, err := c.Web3().Eth.GetBlockNumber()

				if err != nil {
					log.Info("Eth.GetBlockNumber", "error", err)
					time.Sleep(time.Second * 5)
					continue
				}

				// log.Info("Eth.GetLastBlock", "hegiht", blockNumber.Int64())
				if blockNumber.Int64() < c.GetBlockNOw() {

					log.Error("blockNumber.Int64(),so sync from parent", "blockNumber", blockNumber.Int64(), "BlockNOw", c.GetBlockNOw())

					c.AddBlockNow(-1)
					DropBlok(c.GetBlockNOw())

					time.Sleep(time.Millisecond * 100)
					continue
				}

				for blockNumber.Int64() > c.GetBlockNOw() {

					err = SyncOneBlock(c.GetBlockNOw() + 1)
					if err != nil {
						log.Error("SyncOneBlock", "error", err.Error())
						time.Sleep(time.Millisecond * 100)
						continue
					}

					//log.Debug("SyncOneBlock success", "number", c.GetBlockNOw()+1)
					c.AddBlockNow(1)
				}

				time.Sleep(time.Second * 1)
			}
		}()

		<-msg
		time.Sleep(time.Second * 10)
	}

}

var userTransactionCount uint64

func SyncOneBlock(height int64) error {
	defer c.Close()
	transactions := make(map[string]dto.TransactionResponse)
	transactionReceipts := make(map[string]dto.TransactionReceipt)
	// height = 1
	//log.Debug("Start sync block", "height", height)
	if err := DropBlok(height); err != nil {
		return err
	}
	//log.Info("GetBlockByNumber", "height", height)
	//1.get block and parent block
	chain_block, err := c.Web3().Eth.GetBlockByNumber(big.NewInt(height), true)
	if err != nil {
		log.Info("Eth.GetBlockByNumber", "error", err.Error())
		return err
	}
	GLastBlock = chain_block
	//log.Info("Get chain_block success", "number", chain_block.Number, "hash", chain_block.Hash, "detail", chain_block)

	var chain_parent_block *dto.Block
	if height > 1 {
		chain_parent_block, err = c.Web3().Eth.GetBlockByNumber(big.NewInt(height-1), true)
		if err != nil {
			log.Error("Eth.GetBlockByNumber", "height", height-1, "error", err.Error())
			return err
		}
		//log.Info("Get chain_parent_block success", "number", chain_parent_block.Number, "hash", chain_parent_block.Hash)
	}

	//2.get transcations and transreceipts
	for _, transaction := range chain_block.Transactions {
		receipt, err := c.Web3().Eth.GetTransactionReceipt(transaction.Hash)
		if err != nil {
			log.Info("Eth.GetTransactionReceipt", "hash", transaction.Hash, "error", err.Error())
			return err
		}

		transactions[transaction.Hash] = transaction
		transactionReceipts[transaction.Hash] = *receipt

		//log.Info("Get %s  transaction and receipt success", "hash", transaction.Hash, "transaction", transaction, "transactionReceipts", *receipt)
	}

	//3.check parent block
	databases_block_parent, err := (&model.Block{}).FindBlockByHeight(c.Mysql(), height-1)
	if err != nil && err.Error() != DATA_NOT_EXIST {
		log.Error("FindBlockByHeight", "height", height, "error", err.Error())
		return err
	}

	if err != nil && err.Error() == DATA_NOT_EXIST && height != 0 {
		c.AddBlockNow(-1)
		log.Error("FindBlockByHeight,DATA_NOT_EXIST,sync from parent_block", "height", height)
		return err
	}

	if height > 1 && chain_parent_block.Hash != databases_block_parent.F_hash {
		c.AddBlockNow(-1)
		log.Error("chain_parent_block and databases_block_parent not equal,sync from parent_block",
			"pHash", chain_parent_block.Hash, "dHash", databases_block_parent.F_hash)
		return errors.New("parent hash not equal")
	}
	log.Info("WriteBlock")
	//4.write block
	err = WriteBlock(c, *chain_block, transactions, transactionReceipts)
	if err != nil {
		log.Error("WriteBlock failed", "hash", chain_block.Hash)
		return err
	}
	log.Info("transcations")
	//write transcations
	err = WriteTransactions(c, *chain_block, transactions, transactionReceipts)
	if err != nil {
		log.Error("WriteTransactions failed", "number", chain_block.Number.Int64())
		return err
	}

	//todo add map[miner]miner to recount miner reward there .

	return nil
}

func WriteTransactions(c *Connect, chain_block dto.Block, transactions map[string]dto.TransactionResponse, receipts map[string]dto.TransactionReceipt) error {
	//webthree := web3.NewWeb3(providers.NewHTTPProvider(config.Config().Gate, config.Config().TimeOut.RPCTimeOut, false))
	//height
	for tx_hash, transaction := range transactions {
		go func() {
			richRecord := &model.RichRecord{F_address: "", F_balance: 0}
			//log.Info("GetBalance and CreateRichRecord ")
			if balance, err := c.Web3().Eth.GetBalance(transaction.From, block.LATEST); err != nil {
				log.Error("GetBalance failed", "from address", transaction.From, "BlockNumber", transaction.BlockNumber.String(), "error", err.Error())
				//return err
			} else {
				log.Info("GetBalance", "from address", transaction.From, "BlockNumber", transaction.BlockNumber.String())
				richRecord.F_address = transaction.From
				richRecord.F_balance = (balance.Div(balance, big.NewInt(1e18))).Uint64()
				richRecord.UpdateRichRecord(c.Mysql())
			}

			if balance, err := c.Web3().Eth.GetBalance(transaction.To, block.LATEST); err != nil {
				log.Error("GetBalance failed", "to address", transaction.To, "BlockNumber", transaction.BlockNumber.String(), "error", err.Error())
				//return err
			} else {
				richRecord.F_address = transaction.To
				richRecord.F_balance = (balance.Div(balance, big.NewInt(1e18))).Uint64()
				richRecord.UpdateRichRecord(c.Mysql())
			}
		}()
		databases_trans, err := (&model.Transaction{}).FindTrasactionByHash(c.Mysql(), tx_hash)
		if err != nil {
			if err.Error() != DATA_NOT_EXIST {
				log.Error("FindTrasactionByHash", "hash", tx_hash, "error", err.Error())
				return err
			}
		} else {
			if databases_trans.F_status != NORMAL {
				databases_trans.F_status = NORMAL
				err = databases_trans.UpdateTransactionStatus(c.Mysql())
				if err != nil {
					log.Error("UpdateTransactionStatus:%s error", "hash", tx_hash, "error", err.Error())
					return err
				}

				log.Info("UpdateTransactionStatus success", "hash", tx_hash)
				continue
			}
		}

		receipt := receipts[tx_hash]
		tx_fee := big.NewInt(0).Mul(transaction.GasPrice, receipt.GasUsed)

		databases_trans.F_tx_hash = tx_hash
		databases_trans.F_block = chain_block.Number.Int64()
		databases_trans.F_timestamp = chain_block.Timestamp.Int64()
		databases_trans.F_from = transaction.From
		databases_trans.F_to = transaction.To
		databases_trans.F_value = transaction.Value.String()
		databases_trans.F_tx_fee = tx_fee.String()
		databases_trans.F_status = NORMAL
		databases_trans.F_tx_type, databases_trans.F_tx_type_ext = CalcTransactionType(transaction)
		if transaction.From != "0xeecbf083c05984db507fe47f004e1913bb042e06" && transaction.From != "0xaa09ea0d141e87f09fb9193b58cd03268c22ba9a" {
			userTransactionCount++
			//log.Info("CreateTransaction ++", "userTransactionCount", userTransactionCount)
		}

		err = databases_trans.CreateTransaction(c.mysql)
		if err != nil {
			log.Error("CreateTransaction", "hash", tx_hash, "error", err.Error())
			return err
		}
		//log.Info("CreateTransaction success", "databases_trans", databases_trans)

	}

	return nil
}

func CalcTransactionType(transaction dto.TransactionResponse) (txtype int64, txtypeext string) {
	//log.Debug("CalcTransactionType", "transaction.To", transaction.To, "transaction.Input", transaction.Input,
	//	"MORTGAGECONTRACTADDR", MORTGAGECONTRACTADDR, "MORTGAGECONTRACT_FUNC_MORTGAGE", MORTGAGECONTRACT_FUNC_MORTGAGE)
	if transaction.To == MORTGAGECONTRACTADDR {
		if strings.HasPrefix(transaction.Input, MORTGAGECONTRACT_FUNC_MORTGAGE) {
			txtype = model.TX_TYPE_ME_MORTGAGE
			txtypeext = "0x" + transaction.Input[len(transaction.Input)-64:]
		} else if strings.HasPrefix(transaction.Input, MORTGAGECONTRACT_FUNC_REDEEM) {
			txtype = model.TX_TYPE_ME_REDEEM
			txtypeext = "0x" + transaction.Input[len(transaction.Input)-64:]
		}
	}
	return
}

func WriteBlock(c *Connect, chain_block dto.Block, transactions map[string]dto.TransactionResponse, receipts map[string]dto.TransactionReceipt) (err error) {
	//1.find old block
	databases_block, err := (&model.Block{}).FindBlockByHash(c.Mysql(), chain_block.Hash)
	if err != nil {
		if err.Error() != DATA_NOT_EXIST {
			log.Error("FindBlockByHash", "hash", chain_block.Hash, "error", err.Error())
			return err
		}
	} else {
		log.Info("FindBlockByHash", "Hash", chain_block.Hash)

		if databases_block.F_status != NORMAL {
			databases_block.F_status = NORMAL
			err = databases_block.UpdateBlockStatus(c.Mysql())
			if err != nil {
				log.Error("UpdateBlockStatus", "Hash", chain_block.Hash, "error", err.Error())
				return err
			}
			log.Info("UpdateBlockStatus sucess,block hash", "Hash", chain_block.Hash)
		}

		return nil
	}

	//databases_block, err = (&model.Block{}).FindBlockByHeight(c.Mysql(), chain_block.Number.Int64())
	//if err != nil {
	//	if err.Error() != DATA_NOT_EXIST {
	//		fatal_list("FindBlockByHeight:%d error:%s", chain_block.Number.Int64(), err.Error())
	//		return err
	//	}
	//} else {
	//	glog.Info("FindBlockByHeight:%d", chain_block.Number.Int64())
	//
	//	databases_block.F_status = FORK
	//	err = databases_block.UpdateBlockStatus(c.Mysql())
	//	if err != nil {
	//		fatal_list("UpdateBlockStatus:%s error:%s", chain_block.Hash, err.Error())
	//		return err
	//	}
	//
	//	glog.Info("UpdateBlockStatus sucess,block height:%d", chain_block.Number.Int64())
	//}

	//2.count fees
	fees := big.NewInt(0)

	for tx, transaction := range transactions {
		gasPrice := transaction.GasPrice
		gasUsed := receipts[tx].GasUsed
		fees = big.NewInt(0).Add(fees, big.NewInt(0).Mul(gasPrice, gasUsed))
	}

	//log.Info("Block fees", "fees", fees.String())

	//3.write in block
	databases_block.F_block = chain_block.Number.Int64()
	databases_block.F_hash = chain_block.Hash
	databases_block.F_timestamp = chain_block.Timestamp.Uint64()
	databases_block.F_txn = int64(len(transactions))
	// databases_block.F_miner = chain_block.Miner
	databases_block.F_gas_used = chain_block.GasUsed.String()
	databases_block.F_gas_limit = chain_block.GasLimit.String()
	databases_block.F_parent_hash = chain_block.ParentHash
	// databases_block.F_reward = chain_block.Reward.String()
	databases_block.F_fees = fees.String()
	databases_block.F_status = NORMAL

	err = databases_block.CreateBlock(c.Mysql())
	if err != nil {
		log.Error("CreateBlock", "number", chain_block.Number.Int64(), "error", err.Error())
		return err
	}

	// err = WriteMinerRewards(c, chain_block.Miner, chain_block.Reward, fees)
	// if err != nil {
	// 	log.Info("WriteMinerRewards", "Miner", chain_block.Miner, "error", err.Error())
	// 	return err
	// }

	log.Info("CreateBlock success", "databases_block", databases_block)

	return nil
}

func WriteMinerRewards(c *Connect, miner string, reward *big.Int, fees *big.Int) error {
	log.Info("WriteMinerRewards", "miner", miner, "reward", reward, "fees", fees)
	miner_reward, err := (&model.MinerReward{}).FindRewardByMiner(c.Mysql(), miner)
	if err != nil {
		if err.Error() == DATA_NOT_EXIST {
			newMinerReward := &model.MinerReward{
				F_miner:        miner,
				F_total_reward: reward.String(),
				F_total_fees:   fees.String()}
			err := newMinerReward.CreateMinerReward(c.Mysql())
			return err
		}
		return err
	}

	total_old := big.NewInt(0)
	total_old.SetString(miner_reward.F_total_reward, 10)
	total := big.NewInt(0).Add(total_old, reward)
	miner_reward.F_total_reward = total.String()

	total_fees_old := big.NewInt(0)
	total_fees_old.SetString(miner_reward.F_total_fees, 10)
	total_fees := big.NewInt(0).Add(total_old, fees)
	miner_reward.F_total_fees = total_fees.String()

	err = miner_reward.UpdateMinerReward(c.Mysql())
	return err

}
