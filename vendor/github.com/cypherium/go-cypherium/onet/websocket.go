package onet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cypherium/go-cypherium/onet/log"
	"github.com/cypherium/go-cypherium/onet/network"
	"github.com/dedis/protobuf"
	"github.com/gorilla/websocket"
	"gopkg.in/tylerb/graceful.v1"
)

// WebSocket handles incoming client-requests using the websocket
// protocol. When making a new WebSocket, it will listen one port above the
// ServerIdentity-port-#.
// The websocket protocol has been chosen as smallest common denominator
// for languages including JavaScript.
type WebSocket struct {
	services  map[string]Service
	server    *graceful.Server
	mux       *http.ServeMux
	startstop chan bool
	started   bool
	TLSConfig *tls.Config // can only be modified before Start is called
	sync.Mutex
}

// NewWebSocket opens a webservice-listener one port above the given
// ServerIdentity.
func NewWebSocket(si *network.ServerIdentity) *WebSocket {
	w := &WebSocket{
		services:  make(map[string]Service),
		startstop: make(chan bool),
	}
	webHost, err := getWebAddress(si, true)
	log.ErrFatal(err)
	w.mux = http.NewServeMux()
	w.mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		log.Lvl4("ok?", r.RemoteAddr)
		ok := []byte("ok\n")
		w.Write(ok)
	})

	// Add a catch-all handler (longest paths take precedence, so "/" takes
	// all non-registered paths) and correctly upgrade to a websocket and
	// throw an error.
	w.mux.HandleFunc("/", func(wr http.ResponseWriter, re *http.Request) {
		log.Error("request from ", re.RemoteAddr, "for invalid path ", re.URL.Path)

		u := websocket.Upgrader{
			EnableCompression: true,
			// As the website will not be served from ourselves, we
			// need to accept _all_ origins. Cross-site scripting is
			// required.
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		}
		ws, err := u.Upgrade(wr, re, http.Header{})
		if err != nil {
			log.Error(err)
			return
		}

		ws.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(4001, "This service doesn't exist"),
			time.Now().Add(time.Millisecond*500))
		ws.Close()
	})
	w.server = &graceful.Server{
		Timeout: 100 * time.Millisecond,
		Server: &http.Server{
			Addr:    webHost,
			Handler: w.mux,
		},
		NoSignalHandling: true,
	}
	return w
}

// Listening returns true if the server has been started and is
// listening on the ports for incoming connections.
func (w *WebSocket) Listening() bool {
	w.Lock()
	defer w.Unlock()
	return w.started
}

// start listening on the port.
func (w *WebSocket) start() {
	w.Lock()
	w.started = true
	w.server.Server.TLSConfig = w.TLSConfig
	log.Lvl2("Starting to listen on", w.server.Server.Addr)
	started := make(chan bool)
	go func() {
		// Check if server is configured for TLS
		started <- true
		if w.server.Server.TLSConfig != nil && len(w.server.Server.TLSConfig.Certificates) >= 1 {
			w.server.ListenAndServeTLS("", "")
		} else {
			w.server.ListenAndServe()
		}
	}()
	<-started
	w.Unlock()
	w.startstop <- true
}

// registerService stores a service to the given path. All requests to that
// path and it's sub-endpoints will be forwarded to ProcessClientRequest.
func (w *WebSocket) registerService(service string, s Service) error {
	if service == "ok" {
		return errors.New("service name \"ok\" is not allowed")
	}

	w.services[service] = s
	h := &wsHandler{
		service:     s,
		serviceName: service,
	}
	w.mux.Handle(fmt.Sprintf("/%s/", service), h)
	return nil
}

// stop the websocket and free the port.
func (w *WebSocket) stop() {
	w.Lock()
	defer w.Unlock()
	if !w.started {
		return
	}
	log.Lvl3("Stopping", w.server.Server.Addr)
	w.server.Stop(100 * time.Millisecond)
	<-w.startstop
	w.started = false
}

// Pass the request to the websocket.
type wsHandler struct {
	serviceName string
	service     Service
}

// Wrapper-function so that http.Requests get 'upgraded' to websockets
// and handled correctly.
func (t wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rx := 0
	tx := 0
	n := 0
	ok := false

	defer func() {
		log.Lvl2("ws close", r.RemoteAddr, "n", n, "rx", rx, "tx", tx, "ok", ok)
	}()

	u := websocket.Upgrader{
		EnableCompression: true,
		// As the website will not be served from ourselves, we
		// need to accept _all_ origins. Cross-site scripting is
		// required.
		CheckOrigin: func(*http.Request) bool {
			return true
		},
	}
	ws, err := u.Upgrade(w, r, http.Header{})
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		ws.Close()
	}()

	// Loop for each message
	for err == nil {
		mt, buf, rerr := ws.ReadMessage()
		if rerr != nil {
			err = rerr
			break
		}
		rx += len(buf)
		n++

		s := t.service
		var reply []byte
		path := strings.TrimPrefix(r.URL.Path, "/"+t.serviceName+"/")
		log.Lvlf2("ws request from %s: %s/%s", r.RemoteAddr, t.serviceName, path)
		reply, err = s.ProcessClientRequest(r, path, buf)
		if err == nil {
			tx += len(reply)
			err := ws.WriteMessage(mt, reply)
			if err != nil {
				log.Error(err)
				return
			}
		} else {
			log.Lvlf3("Got an error while executing %s/%s: %s", t.serviceName, path, err.Error())
		}
	}

	ws.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(4000, err.Error()),
		time.Now().Add(time.Millisecond*500))
	ok = true
	return
}

type destination struct {
	si   *network.ServerIdentity
	path string
}

// Client is a struct used to communicate with a remote Service running on a
// onet.Server. Using Send it can connect to multiple remote Servers.
type Client struct {
	service     string
	connections map[destination]*websocket.Conn
	suite       network.Suite
	// if not nil, use TLS
	TLSClientConfig *tls.Config
	// whether to keep the connection
	keep bool
	rx   uint64
	tx   uint64
	sync.Mutex
}

// NewClient returns a client using the service s. On the first Send, the
// connection will be started, until Close is called.
func NewClient(suite network.Suite, s string) *Client {
	return &Client{
		service:     s,
		connections: make(map[destination]*websocket.Conn),
		suite:       suite,
	}
}

// NewClientKeep returns a Client that doesn't close the connection between
// two messages if it's the same server.
func NewClientKeep(suite network.Suite, s string) *Client {
	return &Client{
		service:     s,
		keep:        true,
		connections: make(map[destination]*websocket.Conn),
		suite:       suite,
	}
}

// Suite returns the cryptographic suite in use on this connection.
func (c *Client) Suite() network.Suite {
	return c.suite
}

// Send will marshal the message into a ClientRequest message and send it.
func (c *Client) Send(dst *network.ServerIdentity, path string, buf []byte) ([]byte, error) {
	c.Lock()
	defer c.Unlock()
	dest := destination{dst, path}
	conn, ok := c.connections[dest]
	if !ok {
		// Open connection to service.
		url, err := getWebAddress(dst, false)
		if err != nil {
			return nil, err
		}
		log.Lvlf4("Sending %x to %s/%s/%s", buf, url, c.service, path)
		d := &websocket.Dialer{}
		d.TLSClientConfig = c.TLSClientConfig

		var wsProtocol string
		var protocol string
		if c.TLSClientConfig != nil {
			wsProtocol = "wss"
			protocol = "https"
		} else {
			wsProtocol = "ws"
			protocol = "http"
		}
		serverURL := fmt.Sprintf("%s://%s/%s/%s", wsProtocol, url, c.service, path)
		headers := http.Header{"Origin": []string{protocol + "://" + url}}
		// Re-try to connect in case the websocket is just about to start
		for a := 0; a < network.MaxRetryConnect; a++ {
			conn, _, err = d.Dial(serverURL, headers)
			if err == nil {
				break
			}
			time.Sleep(network.WaitRetry)
		}
		if err != nil {
			return nil, err
		}
		c.connections[dest] = conn
	}
	defer func() {
		if !c.keep {
			if err := c.closeConn(dest); err != nil {
				log.Errorf("error while closing the connection to %v : %v\n", dest, err)
			}
		}
	}()
	if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		return nil, err
	}
	c.tx += uint64(len(buf))

	_, rcv, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	log.Lvlf4("Received %x", rcv)
	c.rx += uint64(len(rcv))
	return rcv, nil
}

// SendProtobuf wraps protobuf.(En|De)code over the Client.Send-function. It
// takes the destination, a pointer to a msg-structure that will be
// protobuf-encoded and sent over the websocket. If ret is non-nil, it
// has to be a pointer to the struct that is sent back to the
// client. If there is no error, the ret-structure is filled with the
// data from the service.
func (c *Client) SendProtobuf(dst *network.ServerIdentity, msg interface{}, ret interface{}) error {
	buf, err := protobuf.Encode(msg)
	if err != nil {
		return err
	}
	path := strings.Split(reflect.TypeOf(msg).String(), ".")[1]
	reply, err := c.Send(dst, path, buf)
	if err != nil {
		return err
	}
	if ret != nil {
		return protobuf.DecodeWithConstructors(reply, ret,
			network.DefaultConstructors(c.suite))
	}
	return nil
}

// SendToAll sends a message to all ServerIdentities of the Roster and returns
// all errors encountered concatenated together as a string.
func (c *Client) SendToAll(dst *Roster, path string, buf []byte) ([][]byte, error) {
	msgs := make([][]byte, len(dst.List))
	var errstrs []string
	for i, e := range dst.List {
		var err error
		msgs[i], err = c.Send(e, path, buf)
		if err != nil {
			errstrs = append(errstrs, fmt.Sprint(e.String(), err.Error()))
		}
	}
	var err error
	if len(errstrs) > 0 {
		err = errors.New(strings.Join(errstrs, "\n"))
	}
	return msgs, err
}

// Close sends a close-command to all open connections and returns nil if no
// errors occurred or all errors encountered concatenated together as a string.
func (c *Client) Close() error {
	c.Lock()
	defer c.Unlock()
	var errstrs []string
	for dest := range c.connections {
		if err := c.closeConn(dest); err != nil {
			errstrs = append(errstrs, err.Error())
		}
	}
	var err error
	if len(errstrs) > 0 {
		err = errors.New(strings.Join(errstrs, "\n"))
	}
	return err
}

// closeConn sends a close-command to the connection.
func (c *Client) closeConn(dst destination) error {
	conn, ok := c.connections[dst]
	if ok {
		delete(c.connections, dst)
		conn.WriteMessage(websocket.CloseMessage, nil)
		return conn.Close()
	}
	return nil
}

// Tx returns the number of bytes transmitted by this Client. It implements
// the monitor.CounterIOMeasure interface.
func (c *Client) Tx() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.tx
}

// Rx returns the number of bytes read by this Client. It implements
// the monitor.CounterIOMeasure interface.
func (c *Client) Rx() uint64 {
	c.Lock()
	defer c.Unlock()
	return c.rx
}

// getWebAddress returns the host:port+1 of the serverIdentity. If
// global is true, the address is set to the unspecified 0.0.0.0-address.
func getWebAddress(si *network.ServerIdentity, global bool) (string, error) {
	p, err := strconv.Atoi(si.Address.Port())
	if err != nil {
		return "", err
	}
	host := si.Address.Host()
	if global {
		host = "0.0.0.0"
	}
	return net.JoinHostPort(host, strconv.Itoa(p+1)), nil
}
