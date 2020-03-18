package miner

import (
	"math/big"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/core"
	"github.com/cypherium/go-cypherium/core/types"
	"github.com/cypherium/go-cypherium/reconfig/bftview"

	"github.com/cypherium/go-cypherium/cphdb"
	"github.com/cypherium/go-cypherium/event"
	"github.com/cypherium/go-cypherium/log"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/pow"
	"github.com/cypherium/go-cypherium/pow/cphash"
)

const (
	resultQueueSize = 10
	// The number is referenced from the size of tx pool.
	chainHeadChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
	Start()
	Stop()
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config *params.ChainConfig
	signer types.Signer

	keyBlock  *types.KeyBlock // todo: should be the latest key block header
	candidate *types.Candidate

	createdAt time.Time
}

type Result struct {
	Work      *Work
	Candidate *types.Candidate
}

// worker is the main object which takes care of applying messages to the new state
type worker struct {
	config *params.ChainConfig
	engine pow.Engine

	candidatePool *core.CandidatePool
	mu            sync.Mutex

	// autoCommit loop
	mux        *event.TypeMux
	keyHeadCh  chan core.KeyChainHeadEvent // todo: should be the chain keyblock head
	keyHeadSub event.Subscription

	agents map[Agent]struct{}
	recv   chan *Result

	cph     Backend
	chain   *core.KeyBlockChain // todo: should be the key block
	chainDb cphdb.Database

	pubKey    ed25519.PublicKey
	coinBase  common.Address
	currentMu sync.Mutex
	current   *Work

	IP         []byte
	stopCommit chan chan bool

	// atomic status counters
	atWork  int32 // The number of in-flight pow engine work.
	running int32 // The indicator whether the pow engine is running or not.
}

func newWorker(config *params.ChainConfig, engine pow.Engine, cph Backend, mux *event.TypeMux, candidatePool *core.CandidatePool, extIP net.IP) *worker {
	worker := &worker{
		config:        config,
		engine:        engine,
		cph:           cph,
		mux:           mux,
		keyHeadCh:     make(chan core.KeyChainHeadEvent, chainHeadChanSize),
		chainDb:       cph.ChainDb(),
		recv:          make(chan *Result, resultQueueSize),
		chain:         cph.KeyBlockChain(),
		agents:        make(map[Agent]struct{}),
		IP:            make(net.IP, len(extIP)),
		candidatePool: candidatePool,
		stopCommit:    make(chan chan bool, 1),
		running:       0,
	}

	copy(worker.IP, extIP)
	go worker.wait()

	return worker
}

func (self *worker) start() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.running) == 1 {
		return
	}

	atomic.StoreInt32(&self.running, 1)
	self.keyHeadSub = self.cph.KeyBlockChain().SubscribeChainEvent(self.keyHeadCh)

	go self.autoCommit()

	for agent := range self.agents {
		agent.Start()
	}
}

func (self *worker) stop() {
	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.running) == 0 {
		return
	}

	self.keyHeadSub.Unsubscribe()

	atomic.StoreInt32(&self.running, 0)
	for agent := range self.agents {
		agent.Stop()
	}
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *worker) isRunning() bool {
	return atomic.LoadInt32(&self.running) == 1
}

func (self *worker) register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agents[agent] = struct{}{}
	agent.SetReturnCh(self.recv)
	if self.isRunning() {
		agent.Start()
	}
}

func (self *worker) unregister(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.agents, agent)
	agent.Stop()
}

func (self *worker) autoCommit() {
	log.Info("Miner worker start auto-committing")
	if bftview.IamMember() < 0 {
		self.commitNewWork()
	}

	for {
		select {
		case <-self.keyHeadCh:
			if bftview.IamMember() < 0 {
				self.commitNewWork()
			}
		// Err() channel will be closed when unsubscribing.
		case <-self.keyHeadSub.Err():
			log.Info("Miner worker stop auto-committing")
			return
		}
	}
}

func (self *worker) wait() {
	for {
		for result := range self.recv {
			atomic.AddInt32(&self.atWork, -1)

			if result == nil {
				continue
			}

			candidate := result.Candidate

			if error := self.engine.VerifyCandidate(self.chain, candidate); error != nil {
				log.Error("Fail to verify sealed candidate", "err", error)
				self.commitNewWork()
				continue
			}

			if error := self.candidatePool.AddLocal(candidate); error != nil {
				log.Error("Fail to add local candidate", "err", error)
				if error != core.ErrCandidateExisted {
					self.commitNewWork()
				}
			}

			if cphash.Mode(self.engine.PowMode()) == cphash.ModeLocalMock {
				self.LocalMockAutoTrigNextTermPow(candidate)
			}

		}
	}
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	for agent := range self.agents {
		atomic.AddInt32(&self.atWork, 1)
		if ch := agent.Work(); ch != nil {
			ch <- work
		}
	}
}

// makeCurrent creates a new environment for the current cycle.
func (self *worker) makeCurrent(keyBlock *types.KeyBlock, candidate *types.Candidate) error {
	work := &Work{
		config:    self.config,
		signer:    types.NewEIP155Signer(self.config.ChainID),
		keyBlock:  keyBlock,
		candidate: candidate,
		createdAt: time.Now(),
	}

	self.current = work
	return nil
}

func (self *worker) commitNewWork() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.currentMu.Lock()
	defer self.currentMu.Unlock()
	tstart := time.Now()
	keyBlock := self.chain.CurrentBlock() // todo: current keybBlock
	txBlock := self.cph.BlockChain().CurrentBlock()
	if txBlock.NumberU64() < keyBlock.T_Number() {
		return
	}
	if cphash.Mode(self.engine.PowMode()) != cphash.ModeLocalMock {
		if self.candidatePool.FoundCandidate(keyBlock.Number(), string(self.pubKey)) {
			log.Trace("Found existing candidate of head key block")
			return
		}
	}
	tstamp := tstart.Unix()
	if keyBlock.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = keyBlock.Time().Int64() + 1
	}
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		log.Info("Mining too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}

	port, _ := strconv.Atoi(self.config.OnetPort)
	candidate := types.NewCandidate(keyBlock.Hash(), nil, keyBlock.Number().Uint64()+uint64(1), txBlock.NumberU64(), nil, self.IP, common.HexString(self.pubKey), self.coinBase.String(), port)
	committeeSize := len(self.cph.KeyBlockChain().CurrentCommittee())

	if err := self.engine.PrepareCandidate(self.chain, candidate, committeeSize); err != nil {
		log.Error("Failed to prepare candidate for mining", "err", err)
		return
	}

	self.makeCurrent(keyBlock, candidate)
	work := self.current

	if self.isRunning() {
		//log.Info(fmt.Sprintf("Commit %d mining work", work.keyBlock.Number()), "elapsed", common.PrettyDuration(time.Since(tstart)))
		self.push(work)
	}
}

func (self *worker) SetPubKey(pubKey ed25519.PublicKey) {
	self.pubKey = pubKey
}

func (self *worker) SetCoinbase(eb common.Address) {
	self.coinBase = eb
}
func (self *worker) LocalMockAutoTrigNextTermPow(cand *types.Candidate) {

	block := self.chain.CurrentBlock()
	number := cand.KeyCandidate.Number
	block.SetNumber(number.Add(number, big.NewInt(1)))
	block.SetDifficulty(cand.KeyCandidate.Difficulty)
	block.Hash()
	block.SetTime(big.NewInt(time.Now().Unix()))
	self.chain.CurrentBlockStore(block)
	//log.Info("LocalMockAutoTrigNextTermPow", "Next KeyBlcokNumber", block.Header().Number)
	self.commitNewWork()

}
