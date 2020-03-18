package sync

import (
	//"github.com/gomodule/redigo/redis"
	// "qoobing.com/utillib.golang/log"

	"math/big"

	"github.com/cypherium/cph-service/src/model"

	"errors"
	"go-web3/dto"
	"os"

	. "github.com/cypherium/cph-service/src/const"
	"github.com/cypherium/cph-service/src/util"
	// "qoobing.com/utillib.golang/gls"
	"strings"
	"time"

	"github.com/cypherium/go-cypherium/log"
)

var logid string
var c = new(Connect)
var GLastBlock *dto.Block = nil

//func init (
//
//)

func StartSyncLastBlock() {

	defer c.Close()

	logid = "sync" + util.GetRandomCharacter(4)
	// gls.SetGlsValue("logid", logid)

	//var c = new(Connect)
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

				log.Info("Eth.GetLastBlock", "hegiht", blockNumber.Int64())
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
						log.Info("SyncOneBlock", "error", err.Error())
						time.Sleep(time.Millisecond * 100)
						continue
					}

					log.Info("SyncOneBlock success", "number", c.GetBlockNOw()+1)
					c.AddBlockNow(1)
				}

				time.Sleep(time.Second * 1)
			}
		}()

		<-msg
		time.Sleep(time.Second * 10)
	}

}

func SyncOneBlock(height int64) error {
	transactions := make(map[string]dto.TransactionResponse)
	transactionReceipts := make(map[string]dto.TransactionReceipt)
	// height = 1
	log.Info("Start sync block", "height", height)
	if err := DropBlok(height); err != nil {
		return err
	}
	//1.get block and parent block
	chain_block, err := c.Web3().Eth.GetBlockByNumber(big.NewInt(height), true)
	if err != nil {
		log.Info("Eth.GetBlockByNumber", "error", err.Error())
		return err
	}
	GLastBlock = chain_block
	log.Info("Get chain_block success", "number", chain_block.Number, "hash", chain_block.Hash, "detail", chain_block)

	var chain_parent_block *dto.Block
	if height > 0 {
		chain_parent_block, err = c.Web3().Eth.GetBlockByNumber(big.NewInt(height-1), true)
		if err != nil {
			log.Info("Eth.GetBlockByNumber", "height", height-1, "error", err.Error())
			return err
		}
		log.Info("Get chain_parent_block success", "number", chain_parent_block.Number, "hash", chain_parent_block.Hash)
	}

	//2.get transcations and transreceipts
	for _, hash := range chain_block.Transactions {
		transaction, err := c.Web3().Eth.GetTransactionByHash(hash)
		if err != nil {
			log.Info("Eth.GetTransactionByHash", "hash", hash, "error", err.Error())
			return err
		}

		receipt, err := c.Web3().Eth.GetTransactionReceipt(hash)
		if err != nil {
			log.Info("Eth.GetTransactionReceipt", "hash", hash, "error", err.Error())
			return err
		}

		transactions[hash] = *transaction
		transactionReceipts[hash] = *receipt

		log.Info("Get %s  transaction and receipt success", "hash", hash, "transaction", *transaction, "transactionReceipts", *receipt)
	}

	//3.check parent block
	databases_block_parent, err := (&model.Block{}).FindBlockByHeight(c.Mysql(), height-1)
	if err != nil && err.Error() != DATA_NOT_EXIST {
		log.Info("FindBlockByHeight", "height", height, "error", err.Error())
		return err
	}

	if err != nil && err.Error() == DATA_NOT_EXIST && height != 0 {
		c.AddBlockNow(-1)
		log.Info("FindBlockByHeight,DATA_NOT_EXIST,sync from parent_block", "height", height)
		return err
	}

	if height > 0 && chain_parent_block.Hash != databases_block_parent.F_hash {
		c.AddBlockNow(-1)
		log.Info("chain_parent_block and databases_block_parent not equal,sync from parent_block",
			"pHash", chain_parent_block.Hash, "dHash", databases_block_parent.F_hash)
		return errors.New("parent hash not equal")
	}

	//4.write block
	err = WriteBlock(c, *chain_block, transactions, transactionReceipts)
	if err != nil {
		log.Info("WriteBlock failed", "hash", chain_block.Hash)
		return err
	}
	//write transcations
	err = WriteTransactions(c, *chain_block, transactions, transactionReceipts)
	if err != nil {
		log.Info("WriteTransactions failed", "number", chain_block.Number.Int64())
		return err
	}

	//todo add map[miner]miner to recount miner reward there .

	return nil
}

func WriteTransactions(c *Connect, chain_block dto.Block, transactions map[string]dto.TransactionResponse, receipts map[string]dto.TransactionReceipt) error {

	////1.find trans by height
	//old_transcations, err := (&model.Transaction{}).FindTrasactionByHeight(c.Mysql(), chain_block.Number.Int64())
	//
	//if err != nil {
	//	if err.Error() != DATA_NOT_EXIST {
	//		fatal_list("FindTrasactionByHeight:%d error:%s", chain_block.Number.Int64(), err.Error())
	//		return err
	//	}
	//} else {
	//	glog.Info("Find old_transcations ,height:%d", chain_block.Number.Int64())
	//
	//	for _, trans := range old_transcations {
	//		if trans.F_status == NORMAL {
	//			trans.F_status = FORK
	//			err = trans.UpdateTransactionStatus(c.Mysql())
	//			if err != nil {
	//				fatal_list("UpdateTransactionStatus:%s error:%s", trans.F_tx_hash, err.Error())
	//				return err
	//			}
	//
	//			glog.Info("UpdateTransactionStatus :%s success", trans.F_tx_hash)
	//		}
	//	}
	//
	//}

	//height
	for tx_hash, transaction := range transactions {
		databases_trans, err := (&model.Transaction{}).FindTrasactionByHash(c.Mysql(), tx_hash)
		if err != nil {
			if err.Error() != DATA_NOT_EXIST {
				log.Info("FindTrasactionByHash", "hash", tx_hash, "error", err.Error())
				return err
			}
		} else {
			if databases_trans.F_status != NORMAL {
				databases_trans.F_status = NORMAL
				err = databases_trans.UpdateTransactionStatus(c.Mysql())
				if err != nil {
					log.Info("UpdateTransactionStatus:%s error", "hash", tx_hash, "error", err.Error())
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

		err = databases_trans.CreateTransaction(c.mysql)
		if err != nil {
			log.Info("CreateTransaction", "hash", tx_hash, "error", err.Error())
			return err
		}

		log.Info("CreateTransaction success", "databases_trans", databases_trans)
	}

	return nil
}

func CalcTransactionType(transaction dto.TransactionResponse) (txtype int64, txtypeext string) {
	log.Info("CalcTransactionType", "transaction.To", transaction.To, "transaction.Input", transaction.Input,
		"MORTGAGECONTRACTADDR", MORTGAGECONTRACTADDR, "MORTGAGECONTRACT_FUNC_MORTGAGE", MORTGAGECONTRACT_FUNC_MORTGAGE)
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
			log.Info("FindBlockByHash", "hash", chain_block.Hash, "error", err.Error())
			return err
		}
	} else {
		log.Info("FindBlockByHash", "Hash", chain_block.Hash)

		if databases_block.F_status != NORMAL {
			databases_block.F_status = NORMAL
			err = databases_block.UpdateBlockStatus(c.Mysql())
			if err != nil {
				log.Info("UpdateBlockStatus", "Hash", chain_block.Hash, "error", err.Error())
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

	log.Info("Block fees", "fees", fees.String())

	//3.write in block
	databases_block.F_block = chain_block.Number.Int64()
	databases_block.F_hash = chain_block.Hash
	databases_block.F_timestamp = chain_block.Timestamp.Int64()
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
		log.Info("CreateBlock", "number", chain_block.Number.Int64(), "error", err.Error())
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
