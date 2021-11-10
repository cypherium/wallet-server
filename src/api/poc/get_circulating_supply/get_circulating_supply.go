package get_circulating_supply

import (
	"github.com/cypherium/cypherBFT/go-cypherium/log"
	. "github.com/cypherium/wallet-server/src/apicontext"
	"github.com/cypherium/wallet-server/src/go-web3/eth/block"
	"github.com/labstack/echo"
	"math/big"
)

const BASEACCOUNT = "0xCdd16747E54BE3e2B98eC4e8623f7438f1C435Ce"
const BASEACCOUNTBALANCE = 800000000

func Main(cc echo.Context) error {
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	totalSupply := big.NewInt(BASEACCOUNTBALANCE)
	if balance, err := c.Web3().Eth.GetBalance(BASEACCOUNT, block.LATEST); err != nil {
		log.Error("GetBalance failed", "base account balance", balance, "error", err.Error())
	} else {
		balance.Div(balance, big.NewInt(1e18))
		totalSupply.Sub(totalSupply, balance)
	}
	return c.RESULT(totalSupply.Uint64())
}
