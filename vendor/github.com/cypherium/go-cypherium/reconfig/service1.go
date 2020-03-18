package reconfig

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/core"
	"github.com/cypherium/go-cypherium/core/types"
	"github.com/cypherium/go-cypherium/crypto"
	"github.com/cypherium/go-cypherium/crypto/bls"
	"github.com/cypherium/go-cypherium/event"
	"github.com/cypherium/go-cypherium/hotstuff"
	"github.com/cypherium/go-cypherium/log"
	"github.com/cypherium/go-cypherium/p2p"
	"github.com/cypherium/go-cypherium/params"
	"github.com/cypherium/go-cypherium/reconfig/bftview"
	"github.com/cypherium/go-cypherium/rlp"
)

type committeeInfo1 struct {
	KeyHash   common.Hash
	KeyNumber uint64
	Committee *bftview.Committee `rlp:"nil"`
}

/*
func (c *committeeInfo1) DecodeRLP(s *rlp.Stream) error {
	type committeeInfoEx struct {
		Committee *uint //bftview.Committee
		KeyHash   common.Hash
		KeyNumber uint64
	}
	log.Info("committeeInfo1.DecodeRLP")
	var cex committeeInfoEx
	raw, _ := s.Raw()
	err := rlp.DecodeBytes(raw, &cex)
	//	c.Committee = cex.Committee
	c.KeyHash = cex.KeyHash
	c.KeyNumber = cex.KeyNumber

	return err
}
*/
type bestCandidateInfo1 struct {
	KeyNumber uint64
	Node      *common.Cnode `rlp:"nil"`
	KeyHash   common.Hash
}
type cachedCommitteeInfo1 struct {
	keyHash   common.Hash
	keyNumber uint64
	committee *bftview.Committee `rlp:"nil"`
	node      *common.Cnode      `rlp:"nil"`
}
type committeeMsg1 struct {
	node  *p2p.Node           `rlp:"nil"`
	cinfo *committeeInfo1     `rlp:"nil"`
	best  *bestCandidateInfo1 `rlp:"nil"`
}
type hotstuffMsg1 struct {
	node  *p2p.Node `rlp:"nil"`
	lastN uint64
	hMsg  *hotstuff.HotstuffMessage `rlp:"nil"`
}
type networkMsg1 struct {
	ID   uint64
	Hmsg *hotstuff.HotstuffMessage `rlp:"nil"`
	Cmsg *committeeInfo1           `rlp:"nil"`
	Bmsg *bestCandidateInfo1       `rlp:"nil"`
}
type extraData struct {
	BestCandidate *types.Candidate `rlp:"nil"`
	UnConnected   []string
}

//Service1 work for protcol
type Service1 struct {
	secretKey     *bls.SecretKey
	publicKey     []byte
	transport     *p2p.Transport
	serverID      string
	serverAddress string

	bc         *core.BlockChain
	txService  *txService
	kbc        *core.KeyBlockChain
	keyService *keyService

	protocolMng *hotstuff.HotstuffProtocolManager

	lastCmInfoMap   map[common.Hash]*cachedCommitteeInfo1
	muCommitteeInfo sync.Mutex

	currentView     bftview.View
	waittingView    bftview.View
	lastReqCmNumber uint64
	muCurrentView   sync.Mutex

	replicaView      *bftview.View
	runningState     int32
	lastProposeTime  time.Time
	lastNodesChanged bool
	pacetMakerTimer  *paceMakerTimer

	hotstuffMsgs  []*hotstuffMsg1
	muHotstuffMsg sync.Mutex
	/*
		feed   event.Feed
		msgCh  chan hotstuffMsg1
		msgSub event.Subscription // Subscription for msg event
	*/
	feed1   event.Feed
	msgCh1  chan committeeMsg1
	msgSub1 event.Subscription // Subscription for msg event

	/*
		blockCh     chan *types.Block
		muInserting sync.Mutex
		isInserting bool
	*/
}

func newService1(sName string, re *Reconfig) *Service1 {
	s := new(Service1)
	s.txService = newTxService(s, re.cph, re.config)
	s.keyService = newKeyService(s, re.cph, re.config)

	s.bc = re.cph.BlockChain()
	s.kbc = re.cph.KeyBlockChain()
	s.lastCmInfoMap = make(map[common.Hash]*cachedCommitteeInfo1)

	//	s.msgCh = make(chan hotstuffMsg1, 1024)
	//	s.msgSub = s.feed.Subscribe(s.msgCh)
	s.msgCh1 = make(chan committeeMsg1, 10)
	s.msgSub1 = s.feed1.Subscribe(s.msgCh1)

	//	s.blockCh = make(chan *types.Block, 5)
	s.protocolMng = hotstuff.NewHotstuffProtocolManager(s, nil, nil, params.PaceMakerTimeout*2)

	go s.handleHotStuffMsg()
	go s.handleCommitteeMsg()
	return s
}

//OnNewView --------------------------------------------------------------------------
func (s *Service1) OnNewView(data []byte, extraes [][]byte) error { //buf is snapshot, //verify repla' block before newview
	view := bftview.DecodeToView(data)
	log.Info("OnNewView..", "txNumber", view.TxNumber, "keyNumber", view.KeyNumber)

	s.muCurrentView.Lock()
	s.replicaView = view
	if view.EqualNoIndex(&s.currentView) {
		s.currentView.LeaderIndex = view.LeaderIndex
		s.currentView.ReconfigType = view.ReconfigType
	}
	s.muCurrentView.Unlock()

	var bestCandidates []*types.Candidate
	var unConnected []string
	for _, extraD := range extraes {
		if extraD == nil {
			continue
		}
		extra := DecodeToExtraData(extraD)
		if extra.BestCandidate != nil {
			bestCandidates = append(bestCandidates, extra.BestCandidate)
		}
		if extra.UnConnected != nil {
			unConnected = append(unConnected, extra.UnConnected...)
		}
	}
	s.keyService.setBestCandidateAndBadAddress(bestCandidates, unConnected)

	return nil
}

//CurrentState call by hotstuff
func (s *Service1) CurrentState() ([]byte, string) { //recv by onnewview
	curView := s.getCurrentView()
	leaderID := ""
	mb := bftview.GetCurrentMember()
	if mb != nil {
		leader := mb.List[curView.LeaderIndex]
		//leader := mb.List[0]
		log.Info("CurrentState", "leader index", curView.LeaderIndex, "ip", leader.Address)
		leaderID = bftview.GetNodeID(leader.Address, leader.Public)
	} else {
		log.Warn("CurrentState: can't get current committee!")
		s.Committee_Request(curView.KeyNumber, curView.KeyHash)
	}

	log.Info("CurrentState", "TxNumber", curView.TxNumber, "KeyNumber", curView.KeyNumber, "LeaderIndex", curView.LeaderIndex, "ReconfigType", curView.ReconfigType)

	return curView.EncodeToBytes(), leaderID
}

//GetExtra call by hotstuff
func (s *Service1) GetExtra() []byte {
	s.muCurrentView.Lock()
	reconfigType := s.currentView.ReconfigType
	s.muCurrentView.Unlock()
	if reconfigType <= 0 {
		return nil
	}
	var unConnected []string
	if reconfigType == 1 {
		unConnected = s.transport.UnconnectedNodes()
		s.keyService.setUnconnectedNodes(unConnected)
	}

	return EncodeExtraData(s.keyService.getBestCandidate(true), unConnected)
}

//GetPublicKey call by hotstuff
func (s *Service1) GetPublicKey() []*bls.PublicKey {
	keyblock := s.kbc.CurrentBlock()
	keyNumber := keyblock.NumberU64()
	c := bftview.LoadMember(keyNumber, keyblock.Hash(), false)
	if c == nil {
		return nil
	}
	return c.ToBlsPublicKeys(keyNumber)
}

//Self call by hotstuff
func (s *Service1) Self() string {
	return s.serverID
}

//CheckView call by hotstuff
func (s *Service1) CheckView(data []byte) error {
	if !s.isRunning() {
		return types.ErrNotRunning
	}
	view := bftview.DecodeToView(data)
	knumber := s.kbc.CurrentBlock().NumberU64()
	txnumber := s.bc.CurrentBlock().NumberU64()
	log.Info("CheckView..", "txNumber", view.TxNumber, "keyNumber", view.KeyNumber, "local key number", knumber, "tx number", txnumber)
	if view.KeyNumber < knumber {
		return hotstuff.ErrOldState
	} else if view.KeyNumber > knumber {
		return hotstuff.ErrFutureState
	}
	if view.TxNumber < txnumber {
		return hotstuff.ErrOldState
	} else if view.TxNumber > txnumber {
		return hotstuff.ErrFutureState
	}

	return nil
}

//OnPropose call by hotstuff
func (s *Service1) OnPropose(kState []byte, tState []byte, extra []byte) error { //verify new block
	log.Info("OnPropose..")
	if !s.isRunning() {
		return types.ErrNotRunning
	}

	var err error
	var block *types.Block
	var kblock *types.KeyBlock
	if kState != nil {
		kblock = types.DecodeToKeyBlock(kState)
		log.Info("OnPropose", "keyNumber", kblock.NumberU64())
	}
	if tState != nil {
		block = types.DecodeToBlock(tState)
		log.Info("OnPropose", "txNumber", block.NumberU64())
	}
	if kblock != nil {
		extraD := DecodeToExtraData(extra)
		err = s.keyService.verifyKeyBlock(kblock, extraD.BestCandidate, extraD.UnConnected)
		if err != nil {
			log.Error("verify keyblock", "number", kblock.NumberU64(), "err", err)
			return err
		}
	}
	if block != nil {
		err = s.txService.verifyTxBlock(block)
		if err != nil {
			log.Error("verify txblock", "number", block.NumberU64(), "err", err)
			return err
		}
	}
	s.pacetMakerTimer.start()
	return nil
}

//Propose call by hotstuff
func (s *Service1) Propose() (e error, kState []byte, tState []byte, extra []byte) { //buf recv by onpropose, onviewdown
	log.Info("Propose..")

	proposeOK := false
	defer func() {
		if !proposeOK {
			go func() {
				time.Sleep(2 * time.Second)
				curView := s.getCurrentView()
				if bftview.IamLeader(curView.LeaderIndex) {
					s.addHotstuffMsg(&hotstuffMsg1{lastN: s.bc.CurrentBlock().NumberU64(), hMsg: s.protocolMng.TryProposeMessage()})
					//s.feed.Send(hotstuffMsg1{sid: nil, lastN: s.bc.CurrentBlock().NumberU64(), hMsg: s.protocolMng.TryProposeMessage()})
				}
			}()
		} else {
			s.lastProposeTime = time.Now()
		}
	}()

	if !s.isRunning() {
		err := fmt.Errorf("not running for propose")
		return err, nil, nil, nil
	}

	s.muCurrentView.Lock()
	leaderIndex := s.currentView.LeaderIndex
	reconfigType := s.currentView.ReconfigType
	if !s.replicaView.EqualAll(&s.currentView) {
		log.Error("Propose", "replica view not equal to local current view txNumber", s.currentView.TxNumber, "keyNumber", s.currentView.KeyNumber, "LeaderIndex", leaderIndex, "ReconfigType",
			reconfigType, "replica txNumber", s.replicaView.TxNumber, "keyNumber", s.replicaView.KeyNumber, "LeaderIndex", s.replicaView.LeaderIndex, "ReconfigType", s.replicaView.ReconfigType)
		s.muCurrentView.Unlock()
		return fmt.Errorf("replica view not equal to local current view"), nil, nil, nil
	}
	if !bftview.IamLeader(leaderIndex) {
		proposeOK = true
		err := fmt.Errorf("not leader for propose")
		log.Error("Propose", "error", err)
		s.muCurrentView.Unlock()
		return err, nil, nil, nil
	}
	s.muCurrentView.Unlock()

	if reconfigType > 0 {
		block, keyblock, mb, bestCandi, badAddress, err := s.keyService.tryProposalChangeCommittee(s.bc.CurrentBlock(), reconfigType, leaderIndex)
		if err == nil && block != nil && keyblock != nil && mb != nil {
			tbuf := block.EncodeToBytes()
			kbuf := keyblock.EncodeToBytes()
			extra = EncodeExtraData(bestCandi, []string{badAddress})
			proposeOK = true
			return nil, kbuf, tbuf, extra
		}
		log.Warn("tryProposalChangeCommittee error and tryProposalNewBlock", "error", err)
		data, err := s.txService.tryProposalNewBlock(types.IsKeyBlockSkipType)
		if err != nil {
			log.Error("tryProposalNewBlock.1", "error", err)
			return err, nil, nil, nil
		}
		proposeOK = true
		return nil, nil, data, nil
	}
	data, err := s.txService.tryProposalNewBlock(types.IsTxBlockType)
	if err != nil {
		log.Error("tryProposalNewBlock", "error", err)
		return err, nil, nil, nil
	}
	proposeOK = true
	return nil, nil, data, nil
}

//OnViewDone call by hotstuff
func (s *Service1) OnViewDone(e error, phase uint64, kSign *hotstuff.SignedState, tSign *hotstuff.SignedState) error {
	log.Info("OnViewDone", "phase", phase)
	if !s.isRunning() {
		return types.ErrNotRunning
	}
	if e != nil {
		log.Info("OnViewDone", "error", e)
		return e
	}
	if tSign != nil {
		block := types.DecodeToBlock(tSign.State)
		err := s.txService.decideNewBlock(block, tSign.Sign, tSign.Mask)
		if err != nil {
			return err
		}
	}
	if kSign != nil {
		block := types.DecodeToKeyBlock(kSign.State)
		err := s.keyService.decideNewKeyBlock(block, kSign.Sign, kSign.Mask)
		if err != nil {
			return err
		}
	}

	return nil
}

//Write call by hotstuff------------------------------------------------------------------------------------------------
func (s *Service1) Write(id string, data *hotstuff.HotstuffMessage) error {
	if data.Code != hotstuff.MsgCollectTimeoutView {
		log.Info("Write", "self", s.Self(), "to id", id, "code", hotstuff.ReadableMsgType(data.Code), "ViewId", data.ViewId)
	}

	if id == s.Self() {
		s.addHotstuffMsg(&hotstuffMsg1{hMsg: data})
		//s.feed.Send(hotstuffMsg1{sid: nil, hMsg: data})
		return nil
	}

	mb := bftview.GetCurrentMember()
	if mb == nil {
		return fmt.Errorf("can't find current committee,id %s", id)
	}
	node, _ := mb.Get(id, bftview.ID)
	if node == nil || len(node.Address) < 7 { //1.1.1.1
		err := fmt.Errorf("can't find id %s in current committee", id)
		log.Error("Couldn't send", "err", err)
		return err
	}
	if s.transport != nil {
		err := s.transport.WriteMsgC(node, &networkMsg1{Hmsg: data})
		return err
	}
	return nil
}

//Broadcast call by hotstuff
func (s *Service1) Broadcast(data *hotstuff.HotstuffMessage) []error {
	log.Info("Broadcast", "code", hotstuff.ReadableMsgType(data.Code), "ViewId", data.ViewId)
	if s.transport != nil {
		s.transport.BroadcastMsg(&networkMsg1{Hmsg: data}, false, "")
	}
	return nil //return arr
}

func (s *Service1) addHotstuffMsg(msg *hotstuffMsg1) {
	s.muHotstuffMsg.Lock()
	defer s.muHotstuffMsg.Unlock()
	s.hotstuffMsgs = append(s.hotstuffMsgs, msg)
}

func (s *Service1) getHotstuffMsg() *hotstuffMsg1 {
	s.muHotstuffMsg.Lock()
	defer s.muHotstuffMsg.Unlock()
	if len(s.hotstuffMsgs) > 0 {
		msg := s.hotstuffMsgs[0]
		s.hotstuffMsgs = s.hotstuffMsgs[1:]
		return msg
	}
	return nil
}

func (s *Service1) OnMessage(n *p2p.Node, m *p2p.Msg) {
	var msg networkMsg1
	if err := m.Decode(&msg); err != nil {
		log.Debug("Receive message failed", "from", n.Address(), "code", m.Code, "error", err)
	}

	if msg.Hmsg != nil {
		s.addHotstuffMsg(&hotstuffMsg1{node: n, hMsg: msg.Hmsg})
		//s.feed.Send(hotstuffMsg1{sid: si, hMsg: msg.Hmsg})
		return
	}
	s.feed1.Send(committeeMsg1{node: n, cinfo: msg.Cmsg, best: msg.Bmsg})

	//log.Info("Receive message", "from", n.addr, "code", m.Code)
	//s.received += 1
}

func (s *Service1) handleHotStuffMsg() {
	for {
		msg := s.getHotstuffMsg()
		if msg == nil {
			time.Sleep(2 * time.Millisecond)
			continue
		}
		msgCode := msg.hMsg.Code
		if msgCode != hotstuff.MsgCollectTimeoutView {
			log.Info("handleHotStuffMsg", "id", msg.hMsg.Id, "code", hotstuff.ReadableMsgType(msgCode), "ViewId", msg.hMsg.ViewId)
		}
		var curN uint64
		if msgCode == hotstuff.MsgTryPropose || msgCode == hotstuff.MsgStartNewView {
			curN = s.bc.CurrentBlock().NumberU64()
			if msg.lastN < curN {
				log.Info("handleHotStuffMsg", "code", hotstuff.ReadableMsgType(msgCode), "lastN", msg.lastN, "curN", curN)
				continue
			}
		}
		err := s.protocolMng.HandleMessage(msg.hMsg)
		if err != nil && msgCode == hotstuff.MsgStartNewView {
			log.Warn("handleHotStuffMsg", "MsgStartNewView error", err)
			go func(curN uint64) {
				time.Sleep(1 * time.Second)
				s.sendNewViewMsg(curN)
			}(curN)
		}
	}
}

//-------------------------------------------------------------------------------------------------------------------------
func (s *Service1) syncCommittee(mb *bftview.Committee, keyblock *types.KeyBlock) {
	if !keyblock.HasNewNode() {
		return
	}
	if s.transport != nil {
		msg := &bestCandidateInfo1{Node: mb.In(), KeyHash: keyblock.Hash(), KeyNumber: keyblock.NumberU64()}
		s.transport.BroadcastMsg(&networkMsg1{Bmsg: msg}, true, mb.Leader().Address)
	}
}

func (s *Service1) storeCommitteeInCache(cmInfo *committeeInfo1, best *bestCandidateInfo1) {
	s.muCommitteeInfo.Lock()
	defer s.muCommitteeInfo.Unlock()
	var (
		keyHash   common.Hash
		keyNumber uint64
		committee *bftview.Committee
		node      *common.Cnode
	)
	if cmInfo != nil {
		keyHash = cmInfo.KeyHash
		keyNumber = cmInfo.KeyNumber
		committee = cmInfo.Committee
	} else if best != nil {
		keyHash = best.KeyHash
		keyNumber = best.KeyNumber
		node = best.Node
	}

	ac, ok := s.lastCmInfoMap[keyHash]
	if ok {
		if cmInfo != nil {
			ac.committee = cmInfo.Committee
		}
		if best != nil {
			ac.node = best.Node
		}
		return
	}
	//clear prev map
	maxNumber := s.kbc.CurrentBlock().NumberU64()
	for hash, ac := range s.lastCmInfoMap {
		if ac.keyNumber < maxNumber-9 {
			delete(s.lastCmInfoMap, hash)
		}
	}
	log.Info("@@storeCommitteeInCache", "key number", keyNumber)

	s.lastCmInfoMap[keyHash] = &cachedCommitteeInfo1{keyHash: keyHash, keyNumber: keyNumber, committee: committee, node: node}
}

func (s *Service1) handleCommitteeMsg() {
	for {
		select {
		case msg := <-s.msgCh1:
			if msg.best != nil {
				if bftview.LoadMember(msg.best.KeyNumber, msg.best.KeyHash, true) != nil {
					continue
				}
				log.Info("got bestCandidate", "best KeyNumber", msg.best.KeyNumber)
				s.storeCommitteeInCache(nil, msg.best)
				continue
			}

			cInfo := msg.cinfo
			if cInfo == nil {
				continue
			}
			if cInfo.Committee == nil {
				if msg.node == nil {
					continue
				}
				mb := bftview.LoadMember(cInfo.KeyNumber, cInfo.KeyHash, true)
				if mb == nil {
					continue
				}
				log.Info("committeeInfo1 answer", "number", cInfo.KeyNumber, "adddress", msg.node.Address())
				r, _ := mb.Get(msg.node.Address(), bftview.Address)
				if r != nil {
					log.Info("committeeInfo1 answer..ok", "number", cInfo.KeyNumber)
					s.transport.WriteMsg(msg.node, &networkMsg1{Cmsg: &committeeInfo1{Committee: mb, KeyHash: cInfo.KeyHash, KeyNumber: cInfo.KeyNumber}})
				}
				continue
			}

			if bftview.LoadMember(cInfo.KeyNumber, cInfo.KeyHash, true) != nil {
				continue
			}
			log.Info("committeeInfo1", "number", cInfo.KeyNumber, "adddress", msg.node.Address())
			keyblock := s.kbc.GetBlock(cInfo.KeyHash, cInfo.KeyNumber)
			if keyblock != nil {
				if cInfo.Committee.RlpHash() == keyblock.CommitteeHash() {
					cInfo.Committee.Store(keyblock)
				} else {
					log.Error("handleCommitteeMsg.committeeInfo1", "not the committee hash keyNumber", cInfo.KeyNumber)
				}
			} else {
				s.storeCommitteeInCache(cInfo, nil)
			}

		case <-s.msgSub1.Err():
			log.Error("handleHotStuffMsg Feed error")
			return
		}
	}
}

func (s *Service1) updateCommittee(keyBlock *types.KeyBlock) bool {
	bStore := false
	curKeyBlock := keyBlock
	if curKeyBlock == nil {
		curKeyBlock = s.kbc.CurrentBlock()
	}
	mb := bftview.LoadMember(curKeyBlock.NumberU64(), curKeyBlock.Hash(), true)
	if mb != nil {
		return bStore
	}

	s.muCommitteeInfo.Lock()
	ac, ok := s.lastCmInfoMap[curKeyBlock.Hash()]
	if ok {
		if ac.committee != nil {
			mb = ac.committee
		} else if ac.node != nil {
			mb, _ = bftview.GetCommittee(ac.node, curKeyBlock, true)
		}
	}
	s.muCommitteeInfo.Unlock()

	if mb == nil && !curKeyBlock.HasNewNode() {
		mb, _ = bftview.GetCommittee(nil, curKeyBlock, true)
	}

	if mb != nil {
		if mb.RlpHash() != curKeyBlock.CommitteeHash() {
			log.Error("updateCommittee from cache", "committee.RlpHash != keyblock.CommitteeHash keyblock number", curKeyBlock.NumberU64())
			return bStore
		}
		mb.Store(curKeyBlock)
		bStore = true
		log.Info("updateCommittee from cache", "txNumber", s.bc.CurrentBlock().NumberU64(), "keyNumber", curKeyBlock.NumberU64(), "m0", mb.List[0].Address, "m1", mb.List[1].Address)
	} else {
		log.Info("updateCommittee can't found committee", "txNumber", s.bc.CurrentBlock().NumberU64(), "keyNumber", curKeyBlock.NumberU64())
	}

	return bStore
}

func (s *Service1) Committee_OnStored(keyblock *types.KeyBlock, mb *bftview.Committee) {
	log.Info("store committee", "keyNumber", keyblock.NumberU64(), "ip0", mb.List[0].Address, "ipn", mb.List[len(mb.List)-1].Address)
	if keyblock.HasNewNode() && keyblock.NumberU64() == s.kbc.CurrentBlock().NumberU64() && !s.lastNodesChanged {
		r, _ := mb.Get(s.serverAddress, bftview.Address)
		if r != nil {
			s.adjustConnect(mb)
		}
	}
}

func (s *Service1) Committee_Request(kNumber uint64, hash common.Hash) {
	if kNumber <= s.lastReqCmNumber || !bftview.IamMemberByNumber(kNumber, hash) {
		return
	}
	if s.transport != nil {
		s.transport.BroadcastMsg(&networkMsg1{Cmsg: &committeeInfo1{Committee: nil, KeyHash: hash, KeyNumber: kNumber}}, true, "")
	}
	s.lastReqCmNumber = kNumber
}

func (s *Service1) updateCurrentView(fromKeyBlock bool) { //call by keyblock done
	s.muCurrentView.Lock()
	defer s.muCurrentView.Unlock()

	curBlock := s.bc.CurrentBlock()
	curKeyBlock := s.kbc.CurrentBlock()

	s.currentView.TxNumber = curBlock.NumberU64()
	s.currentView.TxHash = curBlock.Hash()
	s.currentView.KeyNumber = curKeyBlock.NumberU64()
	s.currentView.KeyHash = curKeyBlock.Hash()
	s.currentView.CommitteeHash = curKeyBlock.CommitteeHash()

	if fromKeyBlock || curBlock.NumberU64() >= curKeyBlock.T_Number() {
		s.currentView.LeaderIndex = 0
		s.currentView.ReconfigType = 0
	}
	log.Info("updateCurrentView", "TxNumber", s.currentView.TxNumber, "KeyNumber", s.currentView.KeyNumber, "LeaderIndex", s.currentView.LeaderIndex, "ReconfigType", s.currentView.ReconfigType)
	if (s.currentView.TxNumber >= s.waittingView.TxNumber && s.currentView.KeyNumber >= s.waittingView.KeyNumber) || curBlock.BlockType() == types.IsKeyBlockSkipType {
		s.sendNewViewMsg(s.currentView.TxNumber)
		s.waittingView.KeyNumber = s.currentView.KeyNumber
		s.waittingView.TxNumber = s.currentView.TxNumber
	}
}

func (s *Service1) getCurrentView() *bftview.View {
	s.muCurrentView.Lock()
	defer s.muCurrentView.Unlock()
	v := &s.currentView
	return v
}

func (s *Service1) getBestCandidate(refresh bool) *types.Candidate {
	return s.keyService.getBestCandidate(refresh)
}

func (s *Service1) sendNewViewMsg(curN uint64) {
	if bftview.IamMember() >= 0 && curN >= s.kbc.CurrentBlock().T_Number() {
		s.addHotstuffMsg(&hotstuffMsg1{lastN: curN, hMsg: s.protocolMng.NewViewMessage()})
		//s.feed.Send(hotstuffMsg1{sid: nil, lastN: curN, hMsg: s.protocolMng.NewViewMessage()})
	}
}

func (s *Service1) setNextLeader(reconfigType uint8) {
	s.muCurrentView.Lock()
	defer s.muCurrentView.Unlock()

	unConnected := s.transport.UnconnectedNodes()
	if reconfigType == types.PowReconfig {
		s.currentView.LeaderIndex = s.kbc.GetNextLeaderIndex(0, unConnected)
	} else {
		s.currentView.LeaderIndex = s.kbc.GetNextLeaderIndex(s.currentView.LeaderIndex, unConnected)
	}

	s.currentView.ReconfigType = reconfigType
	log.Info("setNextLeader", "type", reconfigType, "index", s.currentView.LeaderIndex)

	s.waittingView.TxNumber = s.currentView.TxNumber + 1
	s.waittingView.KeyNumber = s.currentView.KeyNumber + 1
}

func (s *Service1) SignString(buf string) (sign []byte, pubKey []byte) {
	return s.secretKey.SignHash(crypto.Keccak256([]byte(buf))).Serialize(), s.publicKey
}

func (s *Service1) VerifySignature(pubKey *bls.PublicKey, buf string, bSign []byte) error {
	var sign bls.Sign
	if err := sign.Deserialize(bSign); err != nil {
		return err
	}

	if !sign.VerifyHash(pubKey, crypto.Keccak256([]byte(buf))) {
		return p2p.ErrAuthSignature
	}

	return nil
}

func (s *Service1) DHKeyExchange(pubKey *bls.PublicKey) bls.PublicKey {
	return bls.DHKeyExchange(s.secretKey, pubKey)
}

func (s *Service1) AuthNode(pubKey []byte) bool {
	if bftview.IamMember() < 0 {
		return false
	}
	curKeyBlock := s.kbc.CurrentBlock()
	mb := bftview.LoadMember(curKeyBlock.NumberU64(), curKeyBlock.Hash(), false)
	if mb == nil {
		return false
	}
	pub := hex.EncodeToString(pubKey)
	n, _ := mb.Get(pub, bftview.PublicKey)
	if n != nil {
		return true
	}
	return false
}

func (s *Service1) AuthIP(ip string) bool {
	if bftview.IamMember() < 0 {
		return true
	}
	mb := bftview.GetCurrentMember()
	if mb != nil { //prevent attacks only for normal member
		for _, r := range mb.List {
			addr := strings.Split(r.Address, ":")
			if len(addr) > 1 && addr[0] == ip {
				return true
			}
		}
		return false
	}

	return true
}

func (s *Service1) OnConnected(n *p2p.Node) {
	log.Info("OnConnected", "node addr", n.Address())

	curKeyBlock := s.kbc.CurrentBlock()
	mb := bftview.LoadMember(curKeyBlock.NumberU64(), curKeyBlock.Hash(), true)
	if mb != nil {
		return
	}

	s.transport.WriteMsg(n, &networkMsg1{Cmsg: &committeeInfo1{Committee: nil, KeyHash: curKeyBlock.Hash(), KeyNumber: curKeyBlock.NumberU64()}})

}

func (s *Service1) OnDisconnected(n *p2p.Node) {
	log.Info("OnDisconnected", "node addr", n.Address())
}

func (s *Service1) procBlockDone(txBlock *types.Block, keyblock *types.KeyBlock) {
	if keyblock == nil {
		keyblock = s.kbc.CurrentBlock()
	} else {
		s.lastNodesChanged = false
		if bftview.IamMember() < 0 {
			s.adjustConnect(nil)
			s.lastNodesChanged = true
		} else {
			mb := bftview.LoadMember(keyblock.NumberU64(), keyblock.Hash(), true)
			if mb != nil {
				s.adjustConnect(mb)
				s.lastNodesChanged = true
			} else {
				s.transport.RestartListen()
			}
		}
	}

	s.pacetMakerTimer.procBlockDone(txBlock, keyblock)
}

func (s *Service1) adjustConnect(mb *bftview.Committee) {
	if s.transport == nil {
		return
	}
	var nodes []*p2p.Node
	if mb == nil {
		s.transport.SetNodes(nodes)
		return
	}

	for _, r := range mb.List {
		nodes = append(nodes, p2p.NewpNode(r.Address, bftview.StrToBlsPubKey(r.Public)))
	}
	s.transport.SetNodes(nodes)
}

func (s *Service1) start(config *common.NodeConfig) {
	if !s.isRunning() {
		s.serverAddress = config.Ip + ":" + config.Port
		s.serverID = bftview.GetNodeID(s.serverAddress, config.Public)
		s.secretKey = bftview.StrToBlsPrivKey(config.Private)
		s.publicKey, _ = hex.DecodeString(config.Public)
		s.protocolMng.UpdateKeyPair(s.secretKey)
		bftview.SetServerInfo(s.serverAddress, config.Public)

		if s.transport == nil {
			nPort, _ := strconv.Atoi(config.Port)
			t, err := p2p.NewTransport(nil, nPort, s)
			if err != nil {
				log.Error("Service1.NewTransport", "error", err)
				return
			}
			s.transport = t
		}

		if bftview.IamMember() >= 0 {
			bStore := s.updateCommittee(nil)
			if !bStore {
				var nodes []*p2p.Node
				mb := bftview.GetCurrentMember()
				for _, r := range mb.List {
					nodes = append(nodes, p2p.NewpNode(r.Address, bftview.StrToBlsPubKey(r.Public)))
				}
				s.transport.SetNodes(nodes)
			}
			s.pacetMakerTimer.start()
		}
		s.updateCurrentView(false)
		s.setRunState(1)
	}
}

func (s *Service1) stop() {
	if s.isRunning() {
		//s.server.Close()
		s.pacetMakerTimer.stop()
		s.setRunState(0)
	}
}

func (s *Service1) isRunning() bool {
	return atomic.LoadInt32(&s.runningState) == 1
}
func (s *Service1) setRunState(state int32) {
	atomic.StoreInt32(&s.runningState, state)
}

func DecodeToExtraData(data []byte) *extraData {
	nilExtra := &extraData{nil, nil}
	if data == nil {
		return nilExtra
	}
	extra := extraData{}
	buff := bytes.NewBuffer(data)
	c := rlp.NewStream(buff, 0)
	//_, size, _ := c.Kind()
	if err := c.Decode(&extra); err != nil {
		//log.Error("DecodeToExtraData", "error", err)
		return nilExtra
	}
	return &extra
}
func EncodeExtraData(bestCandidate *types.Candidate, unConnected []string) []byte {
	if bestCandidate == nil && unConnected == nil {
		return nil
	}
	m := make([]byte, 0)
	buff := bytes.NewBuffer(m)
	err := rlp.Encode(buff, extraData{BestCandidate: bestCandidate, UnConnected: unConnected})
	if err != nil {
		log.Error("extraData", "encode error", err)
		return nil
	}
	return buff.Bytes()
}
