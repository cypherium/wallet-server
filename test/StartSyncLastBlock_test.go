package gotest

import (
	"github.com/cypherium/wallet-server/src/sync"
	"testing"
)

func TestStartSyncLastBlock(t *testing.T) {
	sync.StartSyncLastBlock()
}
