package socket

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-stomp/stomp"
	"golang.org/x/net/websocket"
	"net/url"
	"nexsoft.co.id/nexchief2/config"
	"nexsoft.co.id/nexchief2/constanta"
	"strings"
	"sync"
	"time"
)

const heartbeat = "25000"
const pingPeriode = 10 * time.Second
const defaultRetryPeriode = 10 * time.Second

var SocketClient *WebSocketClient

type WebSocketClient struct {
	url                 url.URL
	token, cookie       string
	sendUser, sendGroup chan []byte
	IsConnected         bool
	ctx                 context.Context
	ctxCancel           context.CancelFunc
	stomp               *stomp.Conn
	stompHeader         []func(*stomp.Conn) error
	wsconn              *websocket.Conn
	mu                  sync.RWMutex
}

// Fungsi ini Digunakan untuk mengisi struct WebsocketClient
// tujuannya adalah untuk menjalankan fungsi2 lain pada socket
// seperti : subscribe, send dan listen
// dipanggil di file main()
// --
func GenerateSocketClient(scheme string, host string, path string) () {
	SocketClient = &WebSocketClient{
		sendUser:  make(chan []byte, 1),
		sendGroup: make(chan []byte, 1),
		url: url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   path,
		},
	}

	SocketClient.ctx, SocketClient.ctxCancel = context.WithCancel(context.Background())

	err := SocketClient.login()
	if err != nil {
		log("Failed Logged-in to Socket Server", "SocketConnection.go", "GenerateSocketClient", constanta.LogLevelError, 400)
	}

	go SocketClient.subscribe()
	go SocketClient.listenSendUser()
	//go conn.listenSendGroup()
	//go SocketClient.PingConnection()
}

// Fungsi yang digunakan untuk open connection ke socket server
// jika terjadi kegagalan untuk open koneksi ke socket server, maka akan melakukan percobaan terus menerus sampai berhasil connect
// step :
// 1. melakukan dial ke socket server --> mendapat keluaran berupa websocket conn
// 2. menggunakan websocket conn diatas untuk di koneksikan diatas stomp client --> mendapat output berupa Stomp conn
// 3. Stomp Conn ini yang Kemudian digunakan untuk melakukan send dan subscribe ke socket server
// --
func (conn *WebSocketClient) Connect() *stomp.Conn {
	fileName := "SocketConnection"
	funcName := "Connect"
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.stomp != nil {
		return conn.stomp
	}

	//conn.reset()
	ticker := time.NewTicker(defaultRetryPeriode)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		select {
		case <-conn.ctx.Done():
			log("Connect Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
			return nil
		default:
			if conn.token == "" {
				err := conn.login()
				if err != nil {
					log("Failed Logged-in to Socket Server", fileName, funcName, constanta.LogLevelError, 400)
					continue
				}
			}

			//certPath := config.ApplicationConfiguration.GetSocketServer().CertPath
			//certFile, err := ioutil.ReadFile(certPath)
			//if err != nil {
			//	fmt.Println("path --> ", certPath)
			//	log("failed to read certs file : "+err.Error(), "SocketConnection.go", "GenerateSocketClient", constanta.LogLevelError, 400)
			//	continue
			//}
			//
			//caCert := x509.NewCertPool()
			//caCert.AppendCertsFromPEM(certFile)
			tlsConfig := &tls.Config{
				//RootCAs:            caCert,
				InsecureSkipVerify: true,
			}

			loc, err := url.ParseRequestURI(conn.url.String())
			if err != nil {
				log("Error Parse URL", fileName, funcName, constanta.LogLevelError, 400)
				continue
			}

			origin, err := url.ParseRequestURI("https://" + conn.url.Host)
			if err != nil {
				log("Error Parse URL", fileName, funcName, constanta.LogLevelError, 400)
				continue
			}

			wsConfig := new(websocket.Config)
			wsConfig.Location = loc
			wsConfig.Origin = origin
			wsConfig.Protocol = []string{"tcp"}
			wsConfig.Version = websocket.ProtocolVersionHybi13
			wsConfig.TlsConfig = tlsConfig
			wsConfig.Header = make(map[string][]string)

			ws, err := websocket.DialConfig(wsConfig)
			if err != nil {
				log(fmt.Sprintf("cannot connect to websocket : %s", conn.url.String()), "SocketConnection", "Connect", constanta.LogLevelError, 500)
				continue
			}

			conn.stompHeader = []func(*stomp.Conn) error{
				stomp.ConnOpt.Header(constanta.TokenHeaderSocket, conn.token),
				stomp.ConnOpt.Header("heart-beat", heartbeat+","+heartbeat),
				stomp.ConnOpt.Header(constanta.CookieSocket, "SOCKETSERVER="+conn.cookie),
				stomp.ConnOpt.HeartBeatError(360 * time.Second),
			}

			sc, err := stomp.Connect(ws, conn.stompHeader...)
			if err != nil {
				log(fmt.Sprintf("cannot connect to websocket : %s", conn.url.String()), fileName, funcName, constanta.LogLevelError, 500)
				continue
			}

			conn.wsconn = ws
			conn.stomp = sc
			conn.IsConnected = true

			log(fmt.Sprintf("Connected to websocket : %s", conn.url.String()), fileName, funcName, constanta.LogLevelInfo, 200)
			return conn.stomp
		}
	}
}

// fungsi yang digunakan untuk melakukan ping ke socket server
// ditujukan untuk mengecek apakah sudah terkoneksi dengan socket server atau belum
// --
func (conn *WebSocketClient) Ping() {
	ticker := time.NewTicker(pingPeriode)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ws := conn.Connect()
			if ws == nil {
				continue
			}
			err := ws.Send("/nexsocket/ping", "text/plain", nil)
			if err != nil {
				log("Failed Send Ping To Websocket : "+err.Error(), "SocketConnection", "Ping", constanta.LogLevelError, 400)
			}
			fmt.Println("Send Ping To Socket Server....")
		case <-conn.ctx.Done():
			return
		}
	}
}

func (conn *WebSocketClient) PingConnection() (isConnect bool) {
	host := config.ApplicationConfiguration.GetSocketServer()
	fileName := "SocketConnection.go"
	funcName := "PingConnection"

	ws := conn.Connect()
	if ws == nil {
		return
	}

	err := ws.Send("/nexsocket/ping", "text/plain", nil)
	if err != nil {
		log("Failed Send Ping To Websocket : "+err.Error(), "SocketConnection", "Ping", constanta.LogLevelError, 400)
		fmt.Errorf("cannot send ping to socket server")
		return
	}
	fmt.Println("Send Ping To Socket Server....")

	//subscribe utuk menerima dari socket server (PONG)
	sub, err := ws.Subscribe(host.Destination.Pong, stomp.AckClient)
	if err != nil {
		log("Failed To Subscribe destination /user/pong", fileName, funcName, constanta.LogLevelError, 400)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

loop:
	for {
		select {
		case msg := <-sub.C:
			if msg.Err != nil {
				conn.reset()
				break loop
			}
			fmt.Println("message pong : ", string(msg.Body))
			if strings.Contains(string(msg.Body), "pong") {
				return true
			}
		case <-ctx.Done():
			log("Subscribe Destination Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
			return false
		}
	}

	return
}

// fungsi ini akan memanggil context cancel untuk menghentikan seluruh orphan goroutines
// jika terjadi error , panik ataupun lainnya yang membuat goroutines dengan infinite loop menjadi orphan
//--
func (conn *WebSocketClient) CloseConnection() {
	conn.ctxCancel()
	conn.reset()
}

//fungsi untuk mengosongkan isi dari struct webscoketclient
func (conn *WebSocketClient) reset() {
	conn.IsConnected = false
	conn.wsconn = nil
	conn.stomp = nil
	conn.token = ""
	conn.cookie = ""
}
