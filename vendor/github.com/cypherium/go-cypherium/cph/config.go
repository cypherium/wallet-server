// Copyright 2017 The go-cypherium Authors
// This file is part of the cypherBFT library.
//
// The cypherBFT library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The cypherBFT library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the cypherBFT library. If not, see <http://www.gnu.org/licenses/>.

package cph

import (
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/common/hexutil"
	"github.com/cypherium/go-cypherium/core"
	"github.com/cypherium/go-cypherium/cph/downloader"
	"github.com/cypherium/go-cypherium/cph/gasprice"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/pow/cphash"
)

// DefaultConfig contains default settings for use on the Cypherium main net.
var DefaultConfig = Config{
	SyncMode: downloader.FullSync,
	Cphash: cphash.Config{
		CacheDir:       "cphash",
		CachesInMem:    2,
		CachesOnDisk:   3,
		DatasetsInMem:  1,
		DatasetsOnDisk: 2,
	},
	NetworkId:     1,
	LightPeers:    100,
	DatabaseCache: 768,
	TrieCache:     256,
	TrieTimeout:   60 * time.Minute,
	GasPrice:      big.NewInt(18 * params.Shannon),

	TxPool: core.DefaultTxPoolConfig,
	GPO: gasprice.Config{
		Blocks:     20,
		Percentile: 60,
	},
}

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if user, err := user.Current(); err == nil {
			home = user.HomeDir
		}
	}
	if runtime.GOOS == "windows" {
		DefaultConfig.Cphash.DatasetDir = filepath.Join(home, "AppData", "Cphash")
	} else {
		DefaultConfig.Cphash.DatasetDir = filepath.Join(home, ".cphash")
	}
}

//go:generate gencodec -type Config -field-override configMarshaling -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Cypherium main net block is used.
	GenesisKey *core.GenesisKey `toml:",omitempty"`
	Genesis    *core.Genesis    `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode
	NoPruning bool

	// Light client options
	LightServ  int `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightPeers int `toml:",omitempty"` // Maximum number of LES client peers

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	TrieCache          int
	TrieTimeout        time.Duration

	// Mining-related options
	Cpherbase    common.Address `toml:",omitempty"`
	MinerThreads int            `toml:",omitempty"`
	ExtraData    []byte         `toml:",omitempty"`
	GasPrice     *big.Int

	// Cphash options
	Cphash cphash.Config

	// Transaction pool options
	TxPool                      core.TxPoolConfig
	LocalTestConfig core.LocalTestIpConfig
	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Miscellaneous options
	DocRoot      string `toml:"-"`
	PublicKeyDir string
	OnetPort     string
}

type configMarshaling struct {
	ExtraData hexutil.Bytes
}
