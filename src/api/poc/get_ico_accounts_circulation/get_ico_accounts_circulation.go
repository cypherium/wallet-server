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

var GenesisAccounts = []string{
	"0xBF79866DE2C7A6E93CCB22B265854C9A12B05887",
	"0x2DCC7D63F6497DA971CDC692B9E51F6B9CA0537B",
	"0x2F0AC2EA37084DC2093C2719ECAFCD05A11C4162",
	"0xD03CEB93E5B9F3FD3ADA6730CABF733213C1C68A",
	"0xcdd16747e54be3e2b98ec4e8623f7438f1c435ce",
}

type InputReq struct {
	PageIndex int `json:"pageIndex" form:"pageIndex"` //范围起点
	PageSize  int `json:"pageSize" form:"pageSize"`   //范围重点
}

type OutputRsp struct {
	IcoAccountCirculation uint64 `json:"ico_account_circulation"`
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
	rsp.IcoAccountCirculation = currentIcoAllAccountsCirculationAmmount
	log.Info("get_ico_accouns_circulation Main", "currentTotalCirculationSupplyAmmount", currentTotalCirculationSupplyAmmount, "curentIcoAllAccountsAmmount", curentIcoAllAccountsAmmount, "currentIcoAllAccountsCirculationAmmount", currentIcoAllAccountsCirculationAmmount)
	return c.RESULT(rsp)
}
