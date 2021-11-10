package get_total_supply

import (
	. "github.com/cypherium/wallet-server/src/apicontext"
	"github.com/labstack/echo"
)

const TOTALSUPPLY = 6828000000

func Main(cc echo.Context) error {
	c := cc.(ApiContext)
	defer c.PANIC_RECOVER()
	return c.RESULT(TOTALSUPPLY)
}
