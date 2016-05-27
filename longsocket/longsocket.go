package longsocket

import (
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	//"github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
)

const (
	SHAKE_HANDS_MSG       = "{\"Status\":\"OK\", \"Message\":\"This is shake hands message.\"}"
	SHAKE_HANDS_FREQUENCY = 60
)

const CHAN_SIZE = 1

const (
	STATUS_INIT = iota
	STATUS_CONNECT
	STATUS_CLOSE
)

type Longsocket struct {
	Ws         *websocket.Conn
	writeCh    chan []byte
	readCh     chan []byte
	ReadBuf    []byte
	ShakeHand  bool
	Url        string
	Protocol   string
	Origin     string
	BufferSize int
	Status     int
	mu         sync.Mutex
}

func NewConn(url, protocol, origin string, shankhand bool, buffersize int) *Longsocket {
	return &Longsocket{
		Ws:         nil,
		writeCh:    make(chan []byte, CHAN_SIZE),
		readCh:     make(chan []byte, CHAN_SIZE),
		ReadBuf:    make([]byte, buffersize),
		ShakeHand:  shankhand,
		Url:        url,
		Protocol:   protocol,
		Origin:     origin,
		BufferSize: buffersize,
		Status:     STATUS_INIT,
	}
}

func (l *Longsocket) Dial(ssl bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var err error
	if ssl {
		l.Ws, err = DialSsl(l.Url, l.Protocol, l.Origin)
	} else {
		l.Ws, err = websocket.Dial(l.Url, l.Protocol, l.Origin)
	}

	if err == nil {
		l.Status = STATUS_CONNECT
	}

	return err
}

func DialSsl(url_ string, protocol string, origin string) (ws *websocket.Conn, err error) {
	config, err := websocket.NewConfig(url_, origin)
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	if err != nil {
		return nil, err
	}
	if protocol != "" {
		config.Protocol = []string{protocol}
	}
	return websocket.DialConfig(config)
}

func (l *Longsocket) SetSocket(ws *websocket.Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Ws != nil {
		l.Ws.Close()
	}

	l.Ws = ws
	l.Status = STATUS_CONNECT
}

func (l *Longsocket) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Status == STATUS_CLOSE {
		return
	}

	if l.Ws != nil {
		l.Ws.Close()
		l.Ws = nil
	}

	close(l.writeCh)
	close(l.readCh)

	l.Status = STATUS_CLOSE
}

func (l *Longsocket) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	for {
		errCount := 0
		if l.Status != STATUS_CONNECT {
			break
		}
		select {
		case <-time.After(time.Second * time.Duration(SHAKE_HANDS_FREQUENCY)):
			if l.ShakeHand {
				_, err := l.Ws.Write([]byte(SHAKE_HANDS_MSG))
				if err != nil {
					errCount++
				}
			}
		case msg := <-l.writeCh:
			_, err := l.Ws.Write(msg)
			if err != nil {
				errCount++
			}
		}

		if errCount != 0 {
			break
		}
	}
	l.Close()
}

func (l *Longsocket) Write(buf []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.Status == STATUS_CONNECT {
		l.writeCh <- buf
		return nil
	} else {
		return errors.New("connection closed")
	}
}

func (l *Longsocket) ReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	for {
		if l.Status != STATUS_CONNECT {
			break
		}

		n, err := l.Ws.Read(l.ReadBuf)
		if err != nil {
			break
		}

		if n > 0 && string(l.ReadBuf[0:n]) != SHAKE_HANDS_MSG {
			l.readCh <- l.ReadBuf[0:n]
		}
	}
	l.Close()
}

type dealmsg func([]byte) error

func (l *Longsocket) Read(f dealmsg) {
	for {
		select {
		case msg := <-l.readCh:
			err := f(msg)
			if err != nil {
				l.Write([]byte(err.Error()))
			}
		}
	}
}