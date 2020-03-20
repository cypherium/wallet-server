package sync

import (
	. "github.com/cypherium/cph-service/src/const"
	"github.com/cypherium/cph-service/src/model"

	"math/big"
	// "qoobing.com/utillib.golang/log"
	"github.com/cypherium/go-cypherium/log"
)

func DropBlok(height int64) error {

	if height <= 1 {
		return nil
	}

	//find block ,reset
	block, err := (&model.Block{}).FindBlockByHeight(c.Mysql(), height)
	if err != nil {
		if err.Error() != DATA_NOT_EXIST {
			log.Info("FindBlockByHeight", "height", height, "error", err.Error())
			return err
		}
		return nil
	}

	log.Info("Find fork block", "height", height, "hash", block.F_hash)

	block.F_status = FORK
	err = block.UpdateBlockStatus(c.Mysql())
	if err != nil {
		log.Info("UpdateBlockStatus", "error", err.Error())
		return err
	}

	//find transaction ,reset
	transactions, err := (&model.Transaction{}).FindTrasactionByHeight(c.Mysql(), height)
	if err != nil {
		if err.Error() != DATA_NOT_EXIST {
			log.Info("FindTrasactionByHeight", "error", err.Error())
			return err
		}
	}
	log.Info("Find old transacions", "height", height, "len", len(transactions))

	// block_fees := big.NewInt(0)
	// block_reward, b := big.NewInt(0).SetString(block.F_reward, 10)
	// if b == false {
	// 	log.Info("DropBlok block_reward should ignore")
	// 	return err
	// }

	for _, transaction := range transactions {
		tx_fee, b := big.NewInt(0).SetString(transaction.F_tx_fee, 10)
		if b == false {
			log.Info("big.NewInt(0).SetString,fale", "tx_fee", tx_fee)
			return err
		}
		// block_fees.Add(block_fees, tx_fee)

		transaction.F_status = FORK
		transaction.UpdateTransactionStatus(c.Mysql())
	}

	//find block_reward
	// miner_reward, err := (&model.MinerReward{}).FindRewardByMiner(c.Mysql(), block.F_miner)
	// if err != nil {
	// 	log.Info("FindRewardByMiner", "error", err.Error())
	// 	return err
	// }

	// reward, b := big.NewInt(0).SetString(miner_reward.F_total_reward, 10)
	// if b == false {
	// 	log.Info("big.NewInt(0).SetString,fale")
	// 	return err
	// }

	// fee, b := big.NewInt(0).SetString(miner_reward.F_total_fees, 10)
	// if b == false {
	// 	log.Info("big.NewInt(0).SetString,fale")
	// 	return err
	// }

	// log.Info("Drop example", "height", height, "old", reward.String(), "block_reward", block_reward)
	// reward.Sub(reward, block_reward)
	// miner_reward.F_total_reward = reward.String()
	// fee.Sub(reward, block_fees)
	// miner_reward.F_total_fees = fee.String()

	// miner_reward.UpdateMinerReward(c.Mysql())

	// log.Info("Drop example", "height", height, "block_reward_new", reward.String())
	log.Info("Drop example", "height", height)
	//time.Sleep(time.Second*10)

	return nil

}
