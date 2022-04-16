package get_richlist

import (
	"fmt"
	. "github.com/cypherium/wallet-server/src/apicontext"
	. "github.com/cypherium/wallet-server/src/const"
	"github.com/cypherium/wallet-server/src/model"
	"github.com/labstack/echo"
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
	RichList RichListInfo `json:"richList"`
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

	rsp := OutputRsp{}

	argc := new(InputReq)

	if err := c.BindInput(argc); err != nil {
		return c.RESULT_PARAMETER_ERROR(err.Error())
	}
	// log.Debugf("receive GetTopNRecords: %+v", argc)
	//getAgain:
	records, err := model.GetTopNRecords(c.Mysql(), 100)
	if err != nil {
		// log.Debugf("GetTopNRecords error:%s", err.Error())
		return c.RESULT_ERROR(GET_BLOCKS_ERROR, fmt.Sprintf("GetTopNRecords error:%s", err.Error())) //c.RESULT(rsp)
	}
	var richListInfo richListInfo
	for index, record := range records {
		richListInfo.Index = index + 1
		richListInfo.Address = record.F_address
		richListInfo.Balance = record.F_balance
		rsp.RichList = append(rsp.RichList, richListInfo)
	}
	return c.RESULT(rsp)
}
