package model

import (
	"errors"
	_ "fmt"
	. "github.com/cypherium/wallet-server/src/const"
	"github.com/cypherium/wallet-server/src/util"
	"github.com/jinzhu/gorm"
)

type RichRecord struct {
	F_id      uint64 `gorm:"column:F_id"`
	F_address string `gorm:"column:F_address"`
	F_balance uint64 `gorm:"column:F_balance"`
}

func (r *RichRecord) TableName() string {
	return "t_rich_record"
}

func (record *RichRecord) CreateRichRecord(db *gorm.DB) (err error) {
	//log.Info("CreateRichRecord", "record", record)
	util.ASSERT(record.F_address != "", "create record, F_address can't be nul")
	rdb := db.Create(&record)
	return rdb.Error
}

func (record *RichRecord) UpdateRichRecord(db *gorm.DB) (err error) {
	var rcd = &RichRecord{}
	rdb := db.Where("F_address = ?", record.F_address).First(&rcd)
	if rdb.RecordNotFound() {
		return record.CreateRichRecord(db)
	}
	balanceinfo := map[string]interface{}{"F_balance": record.F_balance}
	return record.updateRichRecordColumn(db, balanceinfo)

}

func (record *RichRecord) updateRichRecordColumn(db *gorm.DB, balance map[string]interface{}) (err error) {
	//log.Info("updateRichRecordColumn=", "balance", balance)

	tx := db.Begin()
	rdb := tx.Where("F_address = ?", record.F_address).Model(&record).Update(balance)

	if rdb.Error != nil {
		tx.Rollback()
		return rdb.Error
	}

	tx.Commit()
	return rdb.Error
}

func (r *RichRecord) FindRichRecordByHeight(db *gorm.DB, height int64) (record RichRecord, err error) {

	rdb := db.Where("F_record = ? and F_status = ?", height, NORMAL).First(&record)
	if rdb.RecordNotFound() {
		err = errors.New(DATA_NOT_EXIST)
	} else if rdb.Error != nil {
		panic("FindRichRecordByHeight error:" + rdb.Error.Error())
	} else {
		err = nil
	}

	return record, err
}

func GetTopNRecords(db *gorm.DB, size int) (records []RichRecord, err error) {
	rdb := db.Order("F_balance desc").Limit(size).Find(&records)
	if rdb.Error != nil {
		err = errors.New("GetRecentRichRecords error:" + rdb.Error.Error())
	} else {
		err = nil
	}
	return records, err
}
