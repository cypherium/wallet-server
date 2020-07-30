/***********************************************************************
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.
//******
// Filename:
// Description:
// Author:
// CreateTime:
/***********************************************************************/
package config

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	// "qoobing.com/utillib.golang/log"
	"sync"
)

type appConfig struct {
	Server string
	IP     string
	Port   string
	Gate   string

	DB    database `toml:"database"`
	Redis string

	TimeOut timeout

	RateSyncInterval int64
	RateInRedis      int64

	Stats stats
}

type database struct {
	Database     string
	MaxOpenCoons int
	MaxIdleCoons int
	Schema       string
}

type timeout struct {
	BlockchainTimeout int64
	RPCTimeOut        int32
}

type stats struct {
	StatAddr string
	ServerId string
}

//

var (
	cfg  appConfig
	once sync.Once
)

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func Config() *appConfig {
	once.Do(func() {
		parentDirectory := getParentDirectory(getCurrentDirectory())
		fmt.Println(parentDirectory)
		confPaht := path.Join(parentDirectory, "/conf/scan.conf")
		fmt.Println(confPaht)
		doc, err := ioutil.ReadFile(confPaht)
		if err != nil {
			panic("initial config, read config file error:" + err.Error())
		}
		if err := toml.Unmarshal(doc, &cfg); err != nil {
			panic("initial config, unmarshal config file error:" + err.Error())
		}

		if cfg.Stats.StatAddr == "" {
			cfg.Stats.StatAddr = ":3000"
		}

		if cfg.Stats.ServerId == "" {
			cfg.Stats.ServerId = "Scan&Stats"
		}

		//log.Debugf("config:%+v\n", cfg)
	})
	return &cfg
}
