package model

import (
	"errors"
	_ "fmt"
	"github.com/cypherium/wallet-server/src/util"
	"github.com/jinzhu/gorm"
)

type IcoAccountsBalanceRecord struct {
	F_id      uint64 `gorm:"column:F_id"`
	F_address string `gorm:"column:F_address"`
	F_balance uint64 `gorm:"column:F_balance"`
}

func (r *IcoAccountsBalanceRecord) TableName() string {
	return "t_rich_record"
}

func (record *IcoAccountsBalanceRecord) CreateIcoAccountsBalanceRecord(db *gorm.DB) (err error) {
	//log.Info("CreateIcoAccountsBalanceRecord", "record", record)
	util.ASSERT(record.F_address != "", "create record, F_address can't be nul")
	rdb := db.Create(&record)
	return rdb.Error
}

func (record *IcoAccountsBalanceRecord) UpdateIcoAccountsBalanceRecordColumn(db *gorm.DB, balance map[string]interface{}) (err error) {
	var icoAccountRecord = &IcoAccountsBalanceRecord{}
	rdb := db.Where("F_address = ?", record.F_address).First(&icoAccountRecord)
	if rdb.RecordNotFound() {
		return record.CreateIcoAccountsBalanceRecord(db)
	}
	tx := db.Begin()
	rdb = tx.Where("F_address = ?", record.F_address).Model(&record).Update(balance)

	if rdb.Error != nil {
		tx.Rollback()
		return rdb.Error
	}

	tx.Commit()
	return rdb.Error
}

func GetAllIcoAccountsBalanceRecord(db *gorm.DB) (records []IcoAccountsBalanceRecord, err error) {
	rdb := db.Find(&records)
	if rdb.Error != nil {
		err = errors.New("GetRecentIcoAccountsBalanceRecords error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return records, err
}
