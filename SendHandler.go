package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
	"nexsoft.co.id/nexchief2/constanta"
	"nexsoft.co.id/nexchief2/model/errorModel"
	"nexsoft.co.id/nexchief2/socket/dtosocket"
	"runtime/debug"
	"time"
)

// Fungsi yang digunakan untuk mengirim data ke client tujuan dari service lain di nexchief
// Fungsi ini dipanggil dari berbagai service lainnya
// Data kemudian dimasukkan kedalam channel conn.sendUser
// Kemudian data tersebut akan di proses di fungsi listenSendUser
func (conn *WebSocketClient) SendToUser(payload dtosocket.SendMessage) (err errorModel.ErrorModel) {
	fileName := "SendHandler.go"
	funcName := "SendToUser"

	defer func() {
		if r := recover(); r != nil {
			err = errorModel.GenerateRecoverError()
			err.CausedBy = fmt.Errorf(string(debug.Stack()))
			return
		}
	}()

	if payload.Touid == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "TO_UID")
		return
	} else if payload.SrcSender == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "SRC_SENDER")
		return
	} else if payload.Type == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "TYPE")
		return
	} else if payload.Msg == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "MESSAGE")
		return
	} else if payload.MsgSeq == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "MSG_SEQ")
		return
	} else if payload.MsgID == "" {
		err = errorModel.GenerateEmptyFieldError(fileName, funcName, "MSG_ID")
		return
	}

	data, errs := json.Marshal(payload)
	if errs != nil {
		err = errorModel.GenerateUnknownError(fileName, "SendToUser", errs)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	for {
		select {
		case conn.sendUser <- data:
			return errorModel.GenerateNonErrorModel()
		case <-ctx.Done():
			return errorModel.GenerateErrorModel(400, "context canceled", "sendHandler.go", funcName, nil)
		}
	}
}

// fungsi ini digunakan untuk meneruskan message yang akan dikirim dari service lain
// nexchief akan mengirim pada destinasi tujuan
func (conn *WebSocketClient) listenSendUser() {
	fileName := "SendHandler.go"
	funcName := "listenSendUser"

	for {
		select {
		case data := <-conn.sendUser:
			ws := conn.Connect()
			if ws == nil {
				continue
			}

			var sendOpt = []func(*frame.Frame) error{
				stomp.SendOpt.Header("X-Token-Nexsoft", conn.token),
				stomp.SendOpt.Header("heart-beat", "25000,25000"),
				stomp.SendOpt.Header("Cookie", "SOCKETSERVER="+conn.cookie),
			}

			log("ini token socket -->"+conn.token, fileName, funcName, 2, 200)
			log("ini cookie socket -->"+conn.cookie, fileName, funcName, 2, 200)

			err := ws.Send("/nexsocket/senduser", "application/json", data, sendOpt...)
			if err != nil {
				log("Failed To Send Message to Socket : "+err.Error(), "SocketConnection.go", "listenSendUser", constanta.LogLevelError, 400)
			}

		case <-conn.ctx.Done():
			log("Listen SendUser Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
			return
		}
	}

}

// Fungsi yang digunakan untuk mengirim data ke grup tujuan dari service lain di nexchief
// Fungsi ini dipanggil dari berbagai service lainnya
// Data kemudian dimasukkan kedalam channel conn.sendGroup
// Kemudian data tersebut akan di proses di fungsi listenSendGroup
func (conn *WebSocketClient) SendToGroup(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	for {
		select {
		case conn.sendGroup <- data:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		}
	}
}

// funsgi ini digunakan untuk meneruskan message yang akan dikirim dari service lain
// nexchief akan mengirim pada group tujuan
func (conn *WebSocketClient) listenSendGroup() {
	for data := range conn.sendGroup {
		ws := conn.Connect()
		if ws == nil {
			continue
		}

		var sendOpt = []func(*frame.Frame) error{
			stomp.SendOpt.Header("X-Token-Nexsoft", conn.token),
			stomp.SendOpt.Header("heart-beat", "25000,25000"),
			stomp.SendOpt.Header("Cookie", "SOCKETSERVER="+conn.cookie),
		}

		err := ws.Send("/nexsocket/sendgroup", "application/json", data, sendOpt...)
		if err != nil {
			log("Failed To Send Message to Socket :"+err.Error(), "SocketConnection.go", "listenSendGroup", constanta.LogLevelError, 400)
		}
	}
}
