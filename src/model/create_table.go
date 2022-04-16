package model

import (
	"github.com/cypherium/cypherBFT/log"
	"github.com/cypherium/wallet-server/src/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var Schema = config.Config().DB.Schema

var Table = map[string]string{
	"t_transaction": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_transaction (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_tx_hash` varchar(128) NOT NULL DEFAULT ''," +
		"`F_block` int(64)  NOT NULL DEFAULT -1," +
		"`F_timestamp` bigint unsigned   NOT NULL DEFAULT 0," +
		"`F_from` varchar(128) NOT NULL DEFAULT ''," +
		"`F_to` varchar(128) NOT NULL DEFAULT ''," +
		"`F_value` varchar(128) NOT NULL DEFAULT ''," +
		"`F_tx_fee` varchar(128) NOT NULL DEFAULT ''," +
		"`F_status` int(4)  NOT NULL DEFAULT 0," +
		"`F_tx_type` bigint(20)  NOT NULL DEFAULT 0," +
		"`F_tx_type_ext` varchar(128) NOT NULL DEFAULT ''," +
		"`F_create_time` datetime NOT NULL," +
		"`F_modify_time` datetime NOT NULL," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_tx_hash`)," +
		"INDEX (`F_from`)," +
		"INDEX (`F_to`)," +
		"INDEX (`F_block`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",

	"t_pending": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_pending (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_tx_hash` varchar(128) NOT NULL DEFAULT ''," +
		"`F_from` varchar(128) NOT NULL DEFAULT ''," +
		"`F_to` varchar(128) NOT NULL DEFAULT ''," +
		"`F_value` varchar(128) NOT NULL DEFAULT ''," +
		"`F_tx_fee` varchar(128) NOT NULL DEFAULT ''," +
		"`F_status` int(4)  NOT NULL DEFAULT 0," +
		"`F_create_time` datetime NOT NULL," +
		"`F_modify_time` datetime NOT NULL," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_tx_hash`)," +
		"INDEX (`F_from`)," +
		"INDEX (`F_to`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",

	"t_block": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_block (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_block` int(64)  NOT NULL DEFAULT -1," +
		"`F_timestamp` bigint unsigned   NOT NULL DEFAULT 0," +
		"`F_txn` int(64)  NOT NULL DEFAULT -1," +
		"`F_miner` varchar(128) NOT NULL DEFAULT ''," +
		"`F_gas_used` varchar(128) NOT NULL DEFAULT ''," +
		"`F_gas_limit` varchar(128) NOT NULL DEFAULT ''," +
		"`F_hash` varchar(128) NOT NULL DEFAULT ''," +
		"`F_parent_hash` varchar(128) NOT NULL DEFAULT ''," +
		"`F_reward` varchar(128) NOT NULL DEFAULT ''," +
		"`F_fees` varchar(128) NOT NULL DEFAULT ''," +
		"`F_status` int(4)  NOT NULL DEFAULT 0," +
		"`F_create_time` datetime NOT NULL," +
		"`F_modify_time` datetime NOT NULL," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_hash`)," +
		"INDEX (`F_miner`)," +
		"INDEX (`F_block`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",

	"t_miner_reward": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_miner_reward (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_miner` varchar(128) NOT NULL DEFAULT ''," +
		"`F_total_reward` varchar(128) NOT NULL DEFAULT ''," +
		"`F_total_fees` varchar(128) NOT NULL DEFAULT ''," +
		"`F_create_time` datetime NOT NULL," +
		"`F_modify_time` datetime NOT NULL," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_miner`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",

	"t_rich_record": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_rich_record (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_address` varchar(128) NOT NULL DEFAULT ''," +
		"`F_balance` varchar(128) NOT NULL DEFAULT ''," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_address`)," +
		"INDEX (`F_balance`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",
	"t_ico_accounts_balance_record": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_rich_record (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_address` varchar(128) NOT NULL DEFAULT ''," +
		"`F_balance` varchar(128) NOT NULL DEFAULT ''," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_address`)," +
		"INDEX (`F_balance`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",

	"t_rate": "CREATE TABLE IF NOT EXISTS " + Schema + ".t_rate (" +
		"`F_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`F_eth` double  NOT NULL DEFAULT 0," +
		"`F_btc` double  NOT NULL DEFAULT 0," +
		"`F_usd` double  NOT NULL DEFAULT 0," +
		"`F_kwr` double  NOT NULL DEFAULT 0," +
		"`F_timestamp` bigint unsigned  NOT NULL DEFAULT 0," +
		"`F_create_time` datetime NOT NULL," +
		"`F_modify_time` datetime NOT NULL," +

		"PRIMARY KEY (`F_id`)," +
		"UNIQUE KEY (`F_timestamp`)" +
		") ENGINE=InnoDB  DEFAULT CHARSET=utf8 ;",
}

func InitDatabase() {
	db, err := gorm.Open("mysql", config.Config().DB.Database)
	defer db.Close()
	if err != nil {
		log.Error("connect mysql failed", "Database", config.Config().DB.Database, "error", err)
		panic("connect mysql failed")
	}

	if !db.HasTable(&Transaction{}) {
		db.CreateTable(&Transaction{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Transaction{})
	}

	if !db.HasTable(&Block{}) {
		db.CreateTable(&Block{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Block{})
	}

	if !db.HasTable(&MinerReward{}) {
		db.CreateTable(&MinerReward{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&MinerReward{})
	}

	if !db.HasTable(&Rate{}) {
		db.CreateTable(&Rate{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&Rate{})
	}

	if !db.HasTable(&RichRecord{}) {
		db.CreateTable(&RichRecord{})
		db.Model(&RichRecord{}).AddUniqueIndex("F_address", "F_address")
	}
	if !db.HasTable(&IcoAccountsBalanceRecord{}) {
		db.CreateTable(&IcoAccountsBalanceRecord{})
		db.Model(&IcoAccountsBalanceRecord{}).AddUniqueIndex("F_address", "F_address")
	}
	for _, value := range Table {
		db.Exec(value)
	}
}
