//go test -tags bn256 -run TestTransport
package p2p

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cypherium/go-cypherium/crypto"
	"github.com/cypherium/go-cypherium/crypto/bls"
	"github.com/cypherium/go-cypherium/log"
)

type app struct {
	port      int
	self      string
	secretKey *bls.SecretKey
	publicKey *bls.PublicKey

	transport *Transport

	received uint64
}

type AppMessage struct {
	index uint64
	name  string
}

func (a *app) Self() string {
	return a.self
}

func (a *app) SignString(s string) (sign []byte, pubKey []byte) {
	return a.secretKey.SignHash(crypto.Keccak256([]byte(s))).Serialize(), a.publicKey.Serialize()
}

func (a *app) VerifySignature(pubKey *bls.PublicKey, s string, bSign []byte) error {
	var sign bls.Sign
	if err := sign.Deserialize(bSign); err != nil {
		return err
	}

	if !sign.VerifyHash(pubKey, crypto.Keccak256([]byte(s))) {
		return ErrAuthSignature
	}

	return nil
}

func (a *app) DHKeyExchange(pubKey *bls.PublicKey) bls.PublicKey {
	return bls.DHKeyExchange(a.secretKey, pubKey)
}

func (a *app) OnMessage(n *Node, m *Msg) {
	var payload AppMessage
	if err := m.Decode(&payload); err != nil {
		if n != nil {
			log.Debug("Decode message failed", "from", n.addr, "code", m.Code)
		} else {
			log.Debug("Decode local loop message failed")
		}

	}

	//log.Info("Receive message", "from", n.addr, "code", m.Code)
	a.received += 1

	if n != nil {
		a.transport.WriteMsg(n, payload)
	}
}

func (a *app) OnConnected(n *Node) {
	//log.Info("Establish connection ", "between", a.self, "with", n.addr)
}

func (a *app) OnDisconnected(n *Node) {
	log.Info("Diconnect ", "between", a.self, "with", n.addr)
}

func (a *app) AuthNode(pubKey []byte) bool {
	return false
}

func createKeyPairs(n int) ([]*bls.PublicKey, []*bls.SecretKey) {
	pubVec := make([]*bls.PublicKey, 0)
	secVec := make([]*bls.SecretKey, 0)

	for i := 0; i < n; i++ {
		sec := new(bls.SecretKey)
		sec.SetByCSPRNG()
		pub := sec.GetPublicKey()
		pubVec = append(pubVec, pub)
		secVec = append(secVec, sec)
	}
	return pubVec, secVec
}

func createNode(total int, startPortNumber int) ([]*app, []*Node) {
	pubVec, secVec := createKeyPairs(total)

	apps := make([]*app, 0)
	nodes := make([]*Node, 0)

	for i := 0; i < total; i++ {
		port := startPortNumber + i

		n := &Node{
			addr:   fmt.Sprintf("192.168.0.199:%d", port),
			pubKey: pubVec[i],
		}

		a := &app{
			port:      port,
			self:      n.addr,
			publicKey: pubVec[i],
			secretKey: secVec[i],
		}

		apps = append(apps, a)
		nodes = append(nodes, n)
	}

	return apps, nodes
}

func TestTransport(test *testing.T) {
	bls.Init(bls.CurveFp254BNb)
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	//glogger.Verbosity(log.Lvl(3)) // INFO
	glogger.Verbosity(log.Lvl(5)) // DEBUG
	log.Root().SetHandler(glogger)

	log.Info("----------- Setup Transport -----------")

	total := 7

	apps, nodes := createNode(total, 10000)

	transports := make([]*Transport, 0)

	for i, app := range apps {
		t, err := NewTransport(nodes, app.port, app)
		if err != nil {
			test.Fatal("fail to create app", "index", i, "port", app.port, "error", err)
			return
		}
		app.transport = t

		transports = append(transports, t)
	}

	// check all nodes are connected with each other
	time.Sleep(time.Second * 2)
	for _, t := range transports {
		if len(t.connecting) != 0 {
			test.Errorf("%s connecting should be zero", t.app.Self())

			return
		}

		if len(t.connected) != total-1 {
			test.Errorf("%s connected should be %d", t.app.Self(), total-1)

			return
		}
	}
	time.Sleep(time.Second * 1)

	log.Info("-------------- Read/write transport -------------- ")
	// send and receive data
	apps[0].transport.BroadcastMsg(AppMessage{name: "mike", index: 0x5555}, false, "")
	time.Sleep(time.Second * 1)

	log.Info("-------------- Close transport -------------- ")

	for _, t := range transports {
		t.Close()
	}

	for i, app := range apps {
		log.Info("app received", "index", i, "total", app.received)
	}

	time.Sleep(time.Second)
}

func TestRemoveNode(t *testing.T) {
	bls.Init(bls.CurveFp254BNb)
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	//glogger.Verbosity(log.Lvl(3)) // INFO
	glogger.Verbosity(log.Lvl(5)) // DEBUG
	log.Root().SetHandler(glogger)

	log.Info("----------- Test Remove Node -----------")

	total := 3

	apps, nodes := createNode(total, 10000)

	transports := make([]*Transport, 0)

	for i, app := range apps {
		t, err := NewTransport(nodes, app.port, app)
		if err != nil {
			log.Error("fail to create app", "index", i, "port", app.port)
			continue
		}
		app.transport = t

		transports = append(transports, t)
	}

	// check all nodes are connected with each other
	time.Sleep(time.Second * 2)
	for _, t := range transports {
		if len(t.connecting) != 0 {
			log.Error(fmt.Sprintf("%s connecting should be zero", t.app.Self()))
		}

		if len(t.connected) != total-1 {
			log.Error(fmt.Sprintf("%s connected should be %d", t.app.Self(), total-1))
		}
	}
	time.Sleep(time.Second * 1)

	log.Info("-------------- Read/write transport -------------- ")
	// send and receive data
	apps[0].transport.BroadcastMsg(AppMessage{name: "mike", index: 0x5555}, false, "")
	time.Sleep(time.Second * 1)

	log.Info("-------------- Close the last node  -------------- ")
	transports[total-1].Close()
	for _, t := range transports {
		t.RemoveNode(nodes[total-1])
	}

	log.Info("-------------- Wait and Check if the node removed -------------- ")
	time.Sleep(time.Second * 4)
	for i, t := range transports {
		if i == total-1 {
			break
		}
		if len(t.connecting) != 0 {
			log.Error(fmt.Sprintf("%s connecting should be zero", t.app.Self()))
		}

		if len(t.connected) != total-2 {
			log.Error(fmt.Sprintf("%s connected should be %d", t.app.Self(), total-2))
		}
	}

	log.Info("-------------- Tear down other nodes  -------------- ")
	for _, t := range transports {
		t.Close()
	}

	for i, app := range apps {
		log.Info("app received", "index", i, "total", app.received)
	}

	time.Sleep(time.Second)
}
