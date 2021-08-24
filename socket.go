package lavago

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Socket struct {
	cfg                *Config
	connectionAttempts int
	reconnectInterval  time.Duration
	dialer             *websocket.Dialer
	conn               *websocket.Conn
	connected          bool
	sendChan           chan wsData
	DataReceived       func([]byte)
	OnOpen             func()
	ErrorReceived      func(error)
	sync.RWMutex
}

type wsData struct {
	data    []byte
	errChan chan error
}

func NewSocket(cfg *Config) *Socket {
	s := &Socket{
		cfg: cfg,
		// reconnectInterval: cfg.ReconnectDelay, start at 0 because we add up duration everytime we call Connect(), ie. on each retry we += cfg.ReconnectDelay
		dialer: &websocket.Dialer{
			ReadBufferSize:   cfg.BufferSize,
			WriteBufferSize:  cfg.BufferSize,
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
		sendChan:      make(chan wsData),
		DataReceived:  func(b []byte) {},
		OnOpen:        func() {},
		ErrorReceived: func(err error) {},
	}

	return s
}

func (s *Socket) Connect(headers http.Header) error {
	if s.conn != nil {
		return errors.New("websocket is already in open state")
	}
	conn, res, err := s.dialer.Dial(s.cfg.socketEndpoint(), headers)
	if err != nil {
		if s.connectionAttempts < s.cfg.ReconnectAttempts {
			s.connectionAttempts++
			s.reconnectInterval += s.cfg.ReconnectDelay
			time.Sleep(s.reconnectInterval)
			return s.Connect(headers)
		}
		return err
	}
	lverS := res.Header.Get("Lavalink-Api-Version")
	lver, err := strconv.Atoi(lverS)
	if err != nil {
		return err
	}
	if lver != 3 {
		return errors.New("this version of lavago only supports Lavalink v3.x")
	}
	s.conn = conn
	s.connected = true
	go s.OnOpen()
	go s.sendListener()
	go s.readListener()
	return nil
}

func (s *Socket) sendListener() {
	for data := range s.sendChan {
		data.errChan <- s.conn.WriteMessage(websocket.TextMessage, data.data)
	}
}

func (s *Socket) readListener() {
	for {
		msgType, data, err := s.conn.ReadMessage()
		if msgType != websocket.CloseMessage {
			return
		}
		if err != nil {
			go s.ErrorReceived(err)
		}
		go s.DataReceived(data)
	}
}

func (s *Socket) Send(data []byte) error {
	if !s.connected {
		return errors.New("can't send, no connection open")
	}
	if len(data) == 0 {
		return errors.New("can't send no data")
	}
	errChan := make(chan error, 1)
	s.sendChan <- wsData{data, errChan}
	err := <-errChan
	close(errChan)
	return err
}

func (s *Socket) SendJSON(value interface{}) error {
	if !s.connected {
		return errors.New("can't send, no connection open")
	}
	if value == nil {
		return errors.New("can't send nil value")
	}
	errChan := make(chan error, 1)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.sendChan <- wsData{data, errChan}
	err = <-errChan
	close(errChan)
	return err
}

func (s *Socket) Close() error {
	s.Lock()
	s.connected = false
	s.Unlock()
	close(s.sendChan)
	return s.conn.Close()
}
