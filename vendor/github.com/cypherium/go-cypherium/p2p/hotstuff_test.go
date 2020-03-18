//go test -tags bn256 -run TestHotStuff
package p2p

import (
	"github.com/cypherium/go-cypherium/crypto/bls"
	"testing"
)

func TestHotStuff(t *testing.T) {
	bls.Init(bls.CurveFp254BNb)

	testHotstuff_4_Nodes_KState_1_f(t)
	//testHotstuff_4_Nodes_TState(t)
	//testHotstuff_4_Nodes_KTState(t)
	//
	//testHotstuff_7_Nodes_KState(t)
	//testHotstuff_7_Nodes_TState(t)
	//testHotstuff_7_Nodes_KTState(t)
}
