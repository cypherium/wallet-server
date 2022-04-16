package get_ico_accounts_circulation

import (
	"fmt"
	"github.com/cypherium/cypherBFT/log"
	"github.com/cypherium/wallet-server/src/api/poc/get_circulating_supply"
	. "github.com/cypherium/wallet-server/src/apicontext"
	. "github.com/cypherium/wallet-server/src/const"
	"github.com/cypherium/wallet-server/src/model"
	"github.com/labstack/echo"
	"math/big"
)

type InputReq struct {
	PageIndex int `json:"pageIndex" form:"pageIndex"` //范围起点
	PageSize  int `json:"pageSize" form:"pageSize"`   //范围重点
}

type OutputRsp struct {
	IcoAccountsCirculation uint64
}

func Main(cc echo.Context) error {
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Mysql()

	//Step 2. parameters initial

	rsp := OutputRsp{}

	argc := new(InputReq)

	if err := c.BindInput(argc); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	// log.Debugf("receive GetTopNRecords: %+v", argc)
	//getAgain:
	allIcoAccountsBalanceRecord, err := model.GetAllIcoAccountsBalanceRecord(c.Mysql())
	if err != nil {
		// log.Debugf("GetTopNRecords error:%s", err.Error())
		return c.RESULT_ERROR(GET_BLOCKS_ERROR, fmt.Sprintf("GetTopNRecords error:%s", err.Error())) //c.RESULT(rsp)
	}

	var currentTotalCirculationSupplyAmmount, curentIcoAllAccountsAmmount, currentIcoAllAccountsCirculationAmmount uint64
	currentTotalCirculationSupplyAmmount = get_circulating_supply.GetTotalSupply(c).Uint64()
	for _, record := range allIcoAccountsBalanceRecord {
		balance := big.NewInt(int64(record.F_balance))
		log.Info("get_ico_accouns_circulation Main", "balance", balance.Uint64())
		curentIcoAllAccountsAmmount += balance.Uint64()
	}
	currentIcoAllAccountsCirculationAmmount = currentTotalCirculationSupplyAmmount - curentIcoAllAccountsAmmount
	rsp.IcoAccountsCirculation = currentIcoAllAccountsCirculationAmmount
	log.Info("get_ico_accouns_circulation Main", "currentTotalCirculationSupplyAmmount", currentTotalCirculationSupplyAmmount, "curentIcoAllAccountsAmmount", curentIcoAllAccountsAmmount, "currentIcoAllAccountsCirculationAmmount", currentIcoAllAccountsCirculationAmmount)
	return c.RESULT(rsp)
}
