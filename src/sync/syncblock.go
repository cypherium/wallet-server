package sync

import (
	. "github.com/cypherium/cph-service/src/const"
	"github.com/cypherium/cph-service/src/model"

	"math/big"
	// "qoobing.com/utillib.golang/log"
	"github.com/golang/glog"
)

func DropBlok(height int64) error {

	if height <= 0 {
		return nil
	}

	//find block ,reset
	block, err := (&model.Block{}).FindBlockByHeight(c.Mysql(), height)
	if err != nil {
		glog.Info("FindBlockByHeight:%s", err.Error())
		if err.Error() != DATA_NOT_EXIST {
			glog.Info("FindBlockByHeight,error:%s", err.Error())
			return err
		}
		return nil
	}

	glog.Info("Find fork block:%d,hash:%s", height, block.F_hash)

	block.F_status = FORK
	err = block.UpdateBlockStatus(c.Mysql())
	if err != nil {
		glog.Info("UpdateBlockStatus,error:%s", err.Error())
		return err
	}

	//find transaction ,reset
	transactions, err := (&model.Transaction{}).FindTrasactionByHeight(c.Mysql(), height)
	if err != nil {
		if err.Error() != DATA_NOT_EXIST {
			glog.Info("FindTrasactionByHeight,error:%s", err.Error())
			return err
		}
	}
	glog.Info("Find old transacions:%d,num:%d", height, len(transactions))

	block_fees := big.NewInt(0)
	block_reward, b := big.NewInt(0).SetString(block.F_reward, 10)
	if b == false {
		glog.Info("big.NewInt(0).SetString,fale")
		return err
	}

	for _, transaction := range transactions {
		tx_fee, b := big.NewInt(0).SetString(transaction.F_tx_fee, 10)
		if b == false {
			glog.Info("big.NewInt(0).SetString,fale")
			return err
		}
		block_fees.Add(block_fees, tx_fee)

		transaction.F_status = FORK
		transaction.UpdateTransactionStatus(c.Mysql())
	}

	//find block_reward
	miner_reward, err := (&model.MinerReward{}).FindRewardByMiner(c.Mysql(), block.F_miner)
	if err != nil {
		glog.Info("FindRewardByMiner,error:%s", err.Error())
		return err
	}

	reward, b := big.NewInt(0).SetString(miner_reward.F_total_reward, 10)
	if b == false {
		glog.Info("big.NewInt(0).SetString,fale")
		return err
	}

	fee, b := big.NewInt(0).SetString(miner_reward.F_total_fees, 10)
	if b == false {
		glog.Info("big.NewInt(0).SetString,fale")
		return err
	}

	glog.Info("Drop example,height:%d,old:%s,block_reward:%s", height, reward.String(), block_reward)
	reward.Sub(reward, block_reward)
	miner_reward.F_total_reward = reward.String()
	fee.Sub(reward, block_fees)
	miner_reward.F_total_fees = fee.String()

	miner_reward.UpdateMinerReward(c.Mysql())

	glog.Info("Drop example,height:%d,block_reward_new:%s", height, reward.String())
	//time.Sleep(time.Second*10)

	return nil

}
