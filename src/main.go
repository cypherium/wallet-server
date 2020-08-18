/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/cypherium/cypherBFT/go-cypherium/log"
	"github.com/cypherium/wallet-server/src/apicontext"
	"github.com/cypherium/wallet-server/src/config"
	"github.com/cypherium/wallet-server/src/model"
	"github.com/cypherium/wallet-server/src/util"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	glog "github.com/labstack/gommon/log"
	// "qoobing.com/utillib.golang/gls"
	// "qoobing.com/utillib.golang/log"

	"github.com/cypherium/wallet-server/src/api"
	"github.com/cypherium/wallet-server/src/statistics/stats"
	"github.com/cypherium/wallet-server/src/sync"
)

func configLogger(e *echo.Echo) {
	// 定义日志级别
	e.Logger.SetLevel(glog.INFO)
	// 记录业务日志
	echoLog, err := os.OpenFile("log/echo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	// 同时输出到文件和终端
	e.Logger.SetOutput(io.MultiWriter(os.Stdout, echoLog))
}
func main() {

	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(5))
	log.Root().SetHandler(glogger)

	model.InitDatabase()
	exec.LookPath(os.Args[0])
	// filePath, _ := exec.LookPath(os.Args[0])
	// log.Debugf("Program file: %s", filePath)
	go stats.Start()
	go sync.StartSyncLastBlock()
	//go sync.StartSyncRate()
	//go sync.CheckReward()

	e := echo.New()
	configLogger(e)
	e.Static("/", "assets")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"127.0.0.1", "http://localhost:8100"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	//// 设置中间件
	//setMiddleware(e)
	//
	//// 注册路由
	//RegisterRoutes(e)
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		//timer.CountMining()
		return func(c echo.Context) error {
			cc := apicontext.New(c)
			req := cc.Request()
			cc.RecordTime()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = util.GetRandomString(12)
			}
			// gls.SetGlsValue("logid", id)

			// log.Debugf("apicontext created")

			return h(cc)
		}
	})

	////user
	//e.POST("/user/login", api.Login)
	//e.POST("/user/logout", api.LogOut)

	//block
	e.POST("/block/get_block_number", api.GetBlockNumber)
	e.POST("/block/get_by_height", api.GetBlockByHeight)
	e.POST("/block/get_blocks", api.GetBlocks)
	e.POST("/block/get_by_hash", api.GetBlockByHash)

	//transaction
	e.POST("/transaction/get_by_hash", api.GetTransactionByHash)
	e.POST("/transaction/get_transactions", api.GetTransactions)
	e.POST("/transaction/get_by_addr", api.GetByAddr)
	e.POST("/transaction/get_by_addr_and_type", api.GetByAddrAndType)
	e.POST("/transaction/get_by_height", api.GetByHeight)
	e.POST("/transaction/get_addr_pending", api.GetAddrPending)
	e.POST("/transaction/get_hash_pending", api.GetHashPending)

	//mining
	e.POST("/mining/get_mined_block_by_addr", api.GetMinedBlocks)
	e.POST("/mining/get_addr_mining_rewards", api.GetAddrMiningRewards)
	e.POST("/mining/get_mined_block_by_addr_and_date", api.GetMinedblockByAddrAndDate)

	//poc
	e.POST("/cph/get_exchange_rate", api.GetExchangeRate)
	e.GET("/cph/get_exchange_rate", api.GetExchangeRate)
	e.POST("/cph/get_summary", api.GetSummary)
	e.GET("/cph/get_summary", api.GetSummary)
	e.POST("/cph/get_balance", api.GetBalance)

	e.Logger.Fatal(e.Start(":" + config.Config().Port))

}
