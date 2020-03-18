package p2p

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cypherium/go-cypherium/common"
	"github.com/cypherium/go-cypherium/crypto"
	"github.com/cypherium/go-cypherium/crypto/bls"
	"github.com/cypherium/go-cypherium/event"
	"github.com/cypherium/go-cypherium/log"
	"github.com/cypherium/go-cypherium/rlp"
	uuid "gopkg.in/satori/go.uuid.v1"
)

var (
	ErrFailListen         = fmt.Errorf("transport fail to listen")
	ErrOversizeMsg        = fmt.Errorf("transport message size overflows uint24")
	ErrAuthHeaderMismatch = fmt.Errorf("transport auth header mis-match")
	ErrAuthInvalidPubKey  = fmt.Errorf("transport auth invalid publick key")
	ErrAuthSignature      = fmt.Errorf("transport auth verify signature failed")
	ErrAuthNoMatch        = fmt.Errorf("transport auth found no matching node")
	ErrNotConnected       = fmt.Errorf("transport not connected")
)

type Node struct {
	addr   string // ip + port
	pubKey *bls.PublicKey
}

func (n *Node) Address() string        { return n.addr }
func (n *Node) PubKey() *bls.PublicKey { return n.pubKey }
func NewpNode(addr string, pubKey *bls.PublicKey) *Node {
	n := &Node{addr: addr, pubKey: pubKey}
	return n
}

type connect struct {
	id        uuid.UUID
	fd        net.Conn
	rw        *streamRW
	peer      *Node
	createdAt time.Time
	remoteIP  string
}

type netEvent struct {
	evType int
	evConn *connect

	// for write message
	msg Msg

	// exception node when broadcasting
	ignore string

	nodes []*Node
}

const (
	EvConnected = iota
	EvDisconnected
	EvWrite
	EvBroadcast
	EvLocalLoop
	EvSetNode
	EvRetartListen
	//EvStopListen
)

const (
	MsgAuth = iota
	MsgUser
)

type TransportApp interface {
	Self() string
	SignString(string) (sign []byte, pubKey []byte)
	VerifySignature(pubKey *bls.PublicKey, s string, bSign []byte) error
	DHKeyExchange(pubKey *bls.PublicKey) bls.PublicKey

	OnMessage(node *Node, msg *Msg)
	OnConnected(node *Node)
	OnDisconnected(node *Node)

	AuthNode(pubKey []byte) bool
	AuthIP(ip string) bool
}

type Transport struct {
	listener    net.Listener
	listenAddr  string
	isListening bool

	feed   event.Feed
	netCh  chan netEvent
	netSub event.Subscription // Subscription for msg event

	nodes          map[string]*Node
	authenticating map[string]*connect
	connected      map[string]*connect
	connecting     map[string]*Node
	lock           sync.Mutex
	authTimeout    time.Duration

	exit bool

	app TransportApp

	numReader int32
	numDialer int32
}

func NewTransport(nodes []*Node, port int, a TransportApp) (*Transport, error) {
	address := fmt.Sprintf("0.0.0.0:%d", port)

	log.Info("Node ", "listen on", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, ErrFailListen
	}

	trans := &Transport{
		listener:       listener,
		listenAddr:     address,
		netCh:          make(chan netEvent, 20),
		nodes:          make(map[string]*Node),
		authenticating: make(map[string]*connect),
		connected:      make(map[string]*connect),
		connecting:     make(map[string]*Node),
		exit:           false,
		app:            a,
		authTimeout:    time.Second * 10,
	}

	trans.netSub = trans.feed.Subscribe(trans.netCh)

	go trans.acceptLoop()
	go trans.eventLoop()

	trans.isListening = true
	for _, n := range nodes {
		if n.addr == a.Self() {
			continue
		}

		trans.nodes[n.addr] = n
		trans.connecting[n.addr] = n

		go trans.dial(n)
	}

	return trans, nil
}

func (t *Transport) Close() {
	t.exit = true

	t.listener.Close()
	t.netSub.Unsubscribe()

	for _, n := range t.connected {
		n.fd.Close()
	}

	for _, n := range t.authenticating {
		n.fd.Close()
	}
}

func (t *Transport) isDupConnection(ip string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Bypass local test
	if matched, _ := regexp.MatchString(`192.168.0.*`, ip); matched {
		return false
	}
	if matched, _ := regexp.MatchString(`127.0.0.*`, ip); matched {
		return false
	}

	for _, c := range t.authenticating {
		if ip == c.remoteIP {
			return true
		}
	}

	for _, c := range t.connected {
		if ip == c.remoteIP {
			return true
		}
	}

	return false
}

func (t *Transport) acceptLoop() {
out:
	for {
		fd, err := t.listener.Accept()
		if err != nil {
			log.Debug("Listener returns error", "err", err)
			break out
		}

		remoteIP := strings.Split(fd.RemoteAddr().String(), ":")[0]
		if !t.app.AuthIP(remoteIP) {
			fd.Close()
			log.Debug("Refuse connection from unknown node", "ip", remoteIP)
			continue
		}

		if t.isDupConnection(remoteIP) {
			fd.Close()
			log.Debug("Close dup connection", "from", remoteIP)
			continue
		}

		conn := &connect{
			id:        uuid.NewV4(),
			fd:        fd,
			rw:        newStreamRW(fd),
			createdAt: time.Now(),
			remoteIP:  remoteIP,
		}

		log.Debug("accept connection", "local", t.app.Self(), "remote", conn.fd.RemoteAddr().String(), "fd", conn.fd)
		go t.readLoop(conn)

	}

	log.Debug("Quit accept loop", "local", t.app.Self())
}

func (t *Transport) dial(n *Node) {
	atomic.AddInt32(&t.numDialer, 1)
	defer atomic.AddInt32(&t.numDialer, -1)

	if n.addr == t.app.Self() {
		return
	}

	log.Debug("Setup Dialer", "remote", n.addr)

	for {
		if t.exit {
			log.Debug("Dialer quit", "local", t.app.Self(), "remote", n.addr)
			return
		}

		t.lock.Lock()
		if !t.isPeer(n) {
			t.deleteNode(n)
			t.lock.Unlock()
			return
		}
		t.lock.Unlock()

		d := net.Dialer{Timeout: time.Second * 10}
		fd, err := d.Dial("tcp", n.addr)
		if err != nil {
			time.Sleep(time.Second * 2)
			continue
		}

		remoteIP := strings.Split(fd.RemoteAddr().String(), ":")[0]
		conn := &connect{
			id:        uuid.NewV4(),
			fd:        fd,
			rw:        newStreamRW(fd),
			createdAt: time.Now(),
			remoteIP:  remoteIP,
		}

		log.Debug("Dialer exits with new connection", "local", t.app.Self(), "remote", conn.fd.RemoteAddr().String(), "fd", conn.fd)
		go t.readLoop(conn)

		return
	}
}

func (t *Transport) isConnected(n *Node) bool {
	if n == nil {
		return false
	}
	_, exist := t.connected[n.addr]
	return exist
}

func (t *Transport) isPeer(n *Node) bool {
	_, exist := t.nodes[n.addr]
	return exist
}

func (t *Transport) isConnecting(n *Node) bool {
	_, exist := t.connecting[n.addr]
	return exist
}

func (t *Transport) deleteNode(n *Node) {
	log.Debug("Delete node", "node", n.addr)

	delete(t.connecting, n.addr)
	delete(t.connected, n.addr)
	delete(t.nodes, n.addr)
}

func (t *Transport) handleNetEvent(ev *netEvent) {
	t.lock.Lock()
	defer t.lock.Unlock()

	conn := ev.evConn
	switch ev.evType {
	case EvConnected:
		delete(t.connecting, conn.peer.addr)
		delete(t.authenticating, conn.id.String())
		t.app.OnConnected(conn.peer)
		t.dumpConnected()

		//log.Debug("Establish connection ", "local", t.app.Self(), "remote", conn.peer.addr, "fd", conn.fd)
	case EvDisconnected:
		t.app.OnDisconnected(conn.peer)

		if t.isConnected(conn.peer) {
			delete(t.connected, conn.peer.addr)
			t.connecting[conn.peer.addr] = conn.peer

			go t.dial(conn.peer)
		}
		t.dumpConnected()
	case EvWrite:
		if t.isConnected(conn.peer) {
			err := t.connected[conn.peer.addr].rw.writeMsg(&ev.msg)
			if err != nil {
				log.Debug("Transport write message failed", "remote", conn.peer.addr)
				t.connected[conn.peer.addr].fd.Close()
			}
		}
	case EvBroadcast:
		t.broadcast(&ev.msg, ev.ignore)
	case EvLocalLoop:
		t.app.OnMessage(nil, &ev.msg)
	case EvSetNode:
		t.setNode(ev.nodes)
	case EvRetartListen:
		t.restartListen()
		//case EvStopListen:
		//	t.stopListen()
	}
}

func (t *Transport) restartListen() {
	if t.isListening {
		return
	}

	var err error
	t.listener, err = net.Listen("tcp", t.listenAddr)
	if err != nil {
		log.Error("transport fail to listen", "on", t.listenAddr)
		return
	}

	go t.acceptLoop()
	t.isListening = true

	log.Debug("Restart listening", "on", t.listenAddr)
}

func (t *Transport) stopListen() {
	if t.isListening {
		t.listener.Close()
		t.isListening = false
	}
}

func (t *Transport) removeNode(n *Node) {
	if t.isConnected(n) {
		t.connected[n.addr].fd.Close()
	}

	t.deleteNode(n)
}

func (t *Transport) contains(n *Node, nodes []*Node) bool {
	for _, node := range nodes {
		if node.addr == n.addr {
			return true
		}
	}
	return false
}

func (t *Transport) setNode(nodes []*Node) {
	if len(t.nodes) > 0 && len(nodes) == 0 {
		t.stopListen()
	}

	for _, n := range t.nodes {
		if t.contains(n, nodes) {
			continue
		}

		log.Debug("set node ", "remove", n.addr)
		t.removeNode(n)
	}

	for _, n := range nodes {
		if n.addr == t.app.Self() {
			continue
		}

		if t.isPeer(n) {
			log.Debug("set node ", "continue", n.addr)
			continue // already exists
		}

		log.Debug("set node ", "add", n.addr)
		t.nodes[n.addr] = n
		t.connecting[n.addr] = n
		go t.dial(n)
	}
}

func (t *Transport) UnconnectedNodes() []string {
	nodes := make([]string, 0)

	t.lock.Lock()
	defer t.lock.Unlock()

	for _, n := range t.connecting {
		nodes = append(nodes, n.addr)
	}

	return nodes
}

//func (t *Transport)StopListen()  {
//	t.feed.Send(netEvent{
//		evType: EvStopListen,
//	})
//}

func (t *Transport) RestartListen() {
	t.feed.Send(netEvent{
		evType: EvRetartListen,
	})
}

func (t *Transport) sendToNodes(buf []byte, errConn []*connect, ignore string) {
	index := 0

	for _, c := range t.connected {
		if errConn[index] != nil || c.peer.addr == ignore {
			index += 1
			continue
		}

		if _, err := c.rw.conn.Write(buf); err != nil {
			errConn[index] = c
		}

		index += 1
	}
}

func (t *Transport) broadcast(msg *Msg, ignore string) {
	errorConn := make([]*connect, len(t.connected))

	ptype, _ := rlp.EncodeToBytes(msg.Code)

	headBuf := make([]byte, headerSize)
	mSize := uint32(len(ptype)) + msg.Size
	if mSize > maxUint24 {
		log.Error("Message oversize, fail to broadcast")
		return
	}
	putInt24(mSize, headBuf)
	copy(headBuf[3:], headerMagic)

	t.sendToNodes(headBuf, errorConn, ignore)
	t.sendToNodes(ptype, errorConn, ignore)

	pBuf := make([]byte, msg.Size)
	io.ReadFull(msg.Payload, pBuf)
	t.sendToNodes(pBuf, errorConn, ignore)

	//hash := crypto.Keccak256Hash(pBuf)
	//log.Debug("write message...", "code", msg.Code, "size", msg.Size,"ignore", ignore, "hash", hash.String())

	for _, c := range errorConn {
		if c == nil {
			continue
		}
		log.Debug("Transport write message failed", "remote", c.peer.addr)
		c.fd.Close()
	}
}

func (t *Transport) clearTimeoutConnection() {
	t.lock.Lock()
	defer t.lock.Unlock()

	now := time.Now()

	for k, c := range t.authenticating {
		duration := now.Sub(c.createdAt).Seconds()

		if duration > t.authTimeout.Seconds() {
			c.fd.Close()
			delete(t.authenticating, k)
		}
	}

	log.Debug("Authenticating connection", "number", len(t.authenticating))
	log.Debug("Reader", "number", t.numReader)
	log.Debug("Dialer", "number", t.numDialer)
}

func (t *Transport) eventLoop() {
	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case ev := <-t.netCh:
			t.handleNetEvent(&ev)
		case <-ticker.C:
			t.clearTimeoutConnection()
		case <-t.netSub.Err():
			log.Debug("Node quit event loop", "node", t.app.Self())
			return
		}
	}
}

func (t *Transport) addToConnectPool(conn *connect) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	if !t.isPeer(conn.peer) {
		t.deleteNode(conn.peer)
		return false
	}

	if t.isConnected(conn.peer) {
		return false
	}

	t.connected[conn.peer.addr] = conn

	return true
}

func (t *Transport) readLoop(conn *connect) {
	atomic.AddInt32(&t.numReader, 1)
	defer atomic.AddInt32(&t.numReader, -1)

	log.Debug("Read loop start", "local", t.app.Self(), "remote", conn.fd.RemoteAddr().String(), "fd", conn.fd)

	t.lock.Lock()
	t.authenticating[conn.id.String()] = conn
	t.lock.Unlock()

	connected := false
	t.writeAuthMsg(conn.rw)

out:
	for {
		msg, err := conn.rw.readMsg()
		if err != nil {
			log.Debug("Read message fail", "node", t.app.Self(), "fd", conn.fd, "error", err)
			break out
		}

		if msg.Code == MsgAuth {
			if connected {
				log.Warn("Receive redundant auth message", "node", t.app.Self(), "from", conn.peer.addr)
				continue
			}

			var authMsg authPacket
			if err := msg.Decode(&authMsg); err != nil {
				log.Debug("Decode auth message failed", "connId", conn.id.String(), "error", err)
				continue
			}

			sec, err := t.authenticateConnect(conn, &authMsg)
			if err != nil {
				log.Debug("Auth connection failed", "remote", authMsg.Addr, "error", err)
				conn.fd.Close()
				continue
			}
			conn.rw.setSecret(sec)

			if !t.addToConnectPool(conn) {
				log.Debug("Add to connection pool failed", "local", t.app.Self(), "remote", conn.peer.addr, "fd", conn.fd)
				conn.fd.Close()
				return
			}

			log.Debug("Add to connection pool ok", "local", t.app.Self(), "remote", conn.peer.addr, "fd", conn.fd)

			t.feed.Send(netEvent{
				evType: EvConnected,
				evConn: conn,
			})

			connected = true
			continue
		}

		t.app.OnMessage(conn.peer, msg)
	}

	if !t.exit && connected {
		t.feed.Send(netEvent{
			evType: EvDisconnected,
			evConn: conn,
		})
	}
}

func (t *Transport) generateSecret(pubKey *bls.PublicKey) *secrets {
	dhSecret := t.app.DHKeyExchange(pubKey)
	sharedSecret := crypto.Keccak256(dhSecret.Serialize())
	return &secrets{AES: sharedSecret}
}

func (t *Transport) authenticateConnect(conn *connect, authMsg *authPacket) (*secrets, error) {
	pub := bls.GetPublicKey(authMsg.PubKey)
	if pub == nil {
		return nil, ErrAuthInvalidPubKey
	}

	if err := t.app.VerifySignature(pub, authMsg.Addr, authMsg.Signature); err != nil {
		return nil, ErrAuthSignature
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	for _, n := range t.nodes {
		if pub.IsEqual(n.pubKey) && n.addr == authMsg.Addr {
			conn.peer = n
			return t.generateSecret(pub), nil
		}
	}

	log.Debug("auth node not found in current nodes")

	// not found in current node pool, let the application judge
	if t.app.AuthNode(authMsg.PubKey) {
		conn.peer = &Node{addr: authMsg.Addr, pubKey: pub}
		t.nodes[authMsg.Addr] = conn.peer
		t.connecting[authMsg.Addr] = conn.peer

		log.Debug("auth node is accepted by the app", "node", conn.peer.addr)
		return t.generateSecret(pub), nil
	}

	return nil, ErrAuthNoMatch
}

type authPacket struct {
	Addr      string
	PubKey    []byte
	Signature []byte
}

const (
	headerSize = 6
)

var (
	headerMagic = []byte{0xC2, 0xFE, 0x55}
)

func headerCheck(h []byte) error {
	if h[3] == headerMagic[0] && h[4] == headerMagic[1] && h[5] == headerMagic[2] {
		return nil
	}

	return ErrAuthHeaderMismatch
}

func (t *Transport) writeAuthMsg(rw *streamRW) error {
	auth := &authPacket{
		Addr: t.app.Self(),
	}
	auth.Signature, auth.PubKey = t.app.SignString(auth.Addr)

	size, r, err := rlp.EncodeToReader(auth)
	if err != nil {
		return err
	}

	log.Debug("Write auth message", "from", t.app.Self())
	return rw.writeMsg(&Msg{Code: MsgAuth, Size: uint32(size), Payload: r})
}

func (t *Transport) dumpConnected() {
	log.Debug("Transport connected nodes")

	for i, n := range t.connected {
		log.Debug("node ", "index", i, "node", n.peer.addr)
	}
}

func (t *Transport) WriteMsg(n *Node, data interface{}) error {
	if !t.isConnected(n) {
		t.dumpConnected()
		return ErrNotConnected
	}

	size, r, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}

	t.feed.Send(netEvent{
		evType: EvWrite,
		evConn: &connect{peer: n},
		msg:    Msg{Code: MsgUser, Size: uint32(size), Payload: r},
	})

	return nil
}

func (t *Transport) WriteMsgC(r *common.Cnode, data interface{}) error {
	addr := r.Address
	if addr[:6] == "tcp://" {
		addr = addr[6:]
	}

	n, ok := t.nodes[addr]
	if ok && n != nil {
		return t.WriteMsg(n, data)
	}

	h, _ := hex.DecodeString(r.Public)
	return t.WriteMsg(&Node{addr: addr, pubKey: bls.GetPublicKey(h)}, data)
}

func (t *Transport) BroadcastMsg(data interface{}, ignoreLocal bool, ignoreAddr string) error {
	size, r1, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}

	t.feed.Send(netEvent{
		evType: EvBroadcast,
		evConn: nil,
		msg:    Msg{Code: MsgUser, Size: uint32(size), Payload: r1},
		ignore: ignoreAddr,
	})

	if ignoreLocal {
		return nil
	}

	_, r2, err := rlp.EncodeToReader(data)
	if err != nil {
		return err
	}

	t.feed.Send(netEvent{
		evType: EvLocalLoop,
		evConn: nil,
		msg:    Msg{Code: MsgUser, Size: uint32(size), Payload: r2},
	})

	return nil
}

func (t *Transport) SetNodes(nodes []*Node) {
	t.feed.Send(netEvent{
		evType: EvSetNode,
		nodes:  nodes,
	})
}

type streamRW struct {
	authenticated bool
	conn          io.ReadWriter
	secret        *secrets
	enc           cipher.Stream
	dec           cipher.Stream
}

func newStreamRW(c io.ReadWriter) *streamRW {
	return &streamRW{
		conn:          c,
		authenticated: false,
	}
}

func (s *streamRW) setSecret(sec *secrets) {
	s.secret = sec

	encc, err := aes.NewCipher(sec.AES)
	if err != nil {
		panic(err)
	}

	iv := make([]byte, encc.BlockSize())
	s.enc = cipher.NewCTR(encc, iv)
	s.dec = cipher.NewCTR(encc, iv)

	s.authenticated = true
}

func (s *streamRW) readMsg() (*Msg, error) {
	var msg Msg

out:
	for {
		headBuf := make([]byte, headerSize)
		if _, err := io.ReadFull(s.conn, headBuf); err != nil {
			//log.Debug("read message header fail", "error", err)
			return nil, err
		}

		// Check the message header
		mSize := readInt24(headBuf)
		if err := headerCheck(headBuf); err != nil {
			log.Debug("header check fail", "error", err)
			continue
		}

		// Read the message content
		mBuf := make([]byte, mSize)
		if _, err := io.ReadFull(s.conn, mBuf); err != nil {
			log.Debug("read message content fail", "error", err)
			return nil, err
		}

		content := bytes.NewReader(mBuf)
		if err := rlp.Decode(content, &msg.Code); err != nil {
			log.Debug("read message decode fail", "content len", content.Len(), "error", err, "content", content)
			continue
		}
		msg.Size = uint32(content.Len())
		msg.Payload = content

		//ptype, _ := rlp.EncodeToBytes(msg.Code)
		//hash := crypto.Keccak256Hash(mBuf[len(ptype):])
		//log.Debug("read message...", "code", msg.Code, "size", msg.Size, "hash", hash.String())

		break out
	}

	return &msg, nil
}

func (s *streamRW) writeMsg(msg *Msg) error {
	ptype, _ := rlp.EncodeToBytes(msg.Code)

	// Message header
	headBuf := make([]byte, headerSize)
	mSize := uint32(len(ptype)) + msg.Size
	if mSize > maxUint24 {
		return ErrOversizeMsg
	}
	putInt24(mSize, headBuf)
	copy(headBuf[3:], headerMagic)

	if _, err := s.conn.Write(headBuf); err != nil {
		return err
	}

	if _, err := s.conn.Write(ptype); err != nil {
		return err
	}

	pBuf := make([]byte, msg.Size)

	io.ReadFull(msg.Payload, pBuf)
	_, err := s.conn.Write(pBuf)

	//hash := crypto.Keccak256Hash(pBuf)
	//log.Debug("write message...", "code", msg.Code, "size", msg.Size, "hash", hash.String())
	return err
}
