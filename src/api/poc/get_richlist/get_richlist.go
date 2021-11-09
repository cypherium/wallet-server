package get_richlist

import (
	"fmt"
	"github.com/cypherium/cypherBFT/go-cypherium/log"
	. "github.com/cypherium/wallet-server/src/apicontext"
	. "github.com/cypherium/wallet-server/src/const"
	"github.com/cypherium/wallet-server/src/go-web3/eth/block"
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

const BASEACCOUNT = "0xCdd16747E54BE3e2B98eC4e8623f7438f1C435Ce"
const BASEACCOUNTBALANCE = 800000000
const FIRSTRICHMINVALUE = 1000000
const TOPARRYALEN = 100

type InputReq struct {
	PageIndex int `json:"pageIndex" form:"pageIndex"` //范围起点
	PageSize  int `json:"pageSize" form:"pageSize"`   //范围重点
}

type OutputRsp struct {
	ErrNo       int          `json:"err_no"`
	ErrMsg      string       `json:"err_msg"`
	Circulation string       `json:"circulation"`
	RichList    RichListInfo `json:"richList"`
}

type richListInfo struct {
	Index   int    `json:"index"`
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

type RichListInfo []richListInfo

var TopNRecords [TOPARRYALEN]model.RichRecord

func Main(cc echo.Context) error {
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	c.Mysql()

	//Step 2. parameters initial

	rsp := OutputRsp{
		ErrNo:  0,
		ErrMsg: "success",
	}

	argc := new(InputReq)

	if err := c.BindInput(argc); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	// log.Debugf("receive GetTopNRecords: %+v", argc)
getAgain:
	records, err := model.GetTopNRecords(c.Mysql(), 100)
	if err != nil {
		// log.Debugf("GetTopNRecords error:%s", err.Error())
		return c.RESULT_ERROR(GET_BLOCKS_ERROR, fmt.Sprintf("GetTopNRecords error:%s", err.Error())) //c.RESULT(rsp)
	}
	if records[0].F_balance < FIRSTRICHMINVALUE && TopNRecords[0].F_balance < FIRSTRICHMINVALUE {
		goto getAgain
	} else if records[0].F_balance < FIRSTRICHMINVALUE && TopNRecords[0].F_balance > FIRSTRICHMINVALUE {
		records = TopNRecords[:]
		copy(records, TopNRecords[:])
	} else {
		copy(TopNRecords[:TOPARRYALEN-1], records[:TOPARRYALEN-1])
	}
	var richListInfo richListInfo
	for index, record := range records {
		richListInfo.Index = index + 1
		richListInfo.Address = record.F_address
		richListInfo.Balance = record.F_balance
		//log.Info("GetBalance", "Address", record.F_address)
		//log.Info("GetBalance", "balance", record.F_balance)
		rsp.RichList = append(rsp.RichList, richListInfo)
	}
	if balance, err := c.Web3().Eth.GetBalance(BASEACCOUNT, block.LATEST); err != nil {
		log.Error("GetBalance failed", "base account balance", balance, "error", err.Error())
	} else {
		totalSupply := big.NewInt(BASEACCOUNTBALANCE)
		balance.Div(balance, big.NewInt(1e18))
		//log.Info("GetBalance", "circulation", circulation)
		//log.Info("GetBalance", "balance", balance.Uint64())

		totalSupply.Sub(totalSupply, balance)
		rsp.Circulation = totalSupply.String()
		//log.Info("GetBalance", "circulation", rsp.Circulation)
	}

	return c.RESULT(rsp)
}
