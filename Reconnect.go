package socket

import (
	"fmt"
	"github.com/go-stomp/stomp"
	"golang.org/x/net/websocket"
	"nexsoft.co.id/nexchief2/constanta"
	"time"
)

// fungsi ini digunakan / dipanggil saat akan melakukan koneksi ulang
// case : jika koneksi mati, socket server mati etc.
func (conn *WebSocketClient) reconnect() bool {
	var err error
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.reset()

	conn.wsconn, err = websocket.Dial(conn.url.String(), "tcp", conn.url.Host)
	if err != nil {
		log(fmt.Sprintf("cannot connect to websocket : %s", conn.url.String()), "SocketConnection", "Connect", constanta.LogLevelError, 500)
		return false
	}

	if conn.token == "" {
		err = conn.login()
		if err != nil {
			return false
		}
	}

	conn.stompHeader = []func(*stomp.Conn) error{
		stomp.ConnOpt.Header(constanta.TokenHeaderSocket, conn.token),
		stomp.ConnOpt.Header("heart-beat", heartbeat+","+heartbeat),
		stomp.ConnOpt.Header(constanta.CookieSocket, "SOCKETSERVER="+conn.cookie),
		stomp.ConnOpt.HeartBeatError(360 * time.Second),
	}

	conn.stomp, err = stomp.Connect(conn.wsconn, conn.stompHeader...)
	if err != nil {
		log(fmt.Sprintf("cannot connect to websocket : %s", conn.url.String()), "SocketConnection", "Connect", constanta.LogLevelError, 500)
		return false
	}

	log(fmt.Sprintf("Connected to websocket : %s", conn.url.String()), "SocketConnection", "Connect", constanta.LogLevelInfo, 200)
	return true
}
