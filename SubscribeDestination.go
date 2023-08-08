package socket

import (
	"encoding/json"
	"fmt"
	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
	"nexsoft.co.id/nexchief2/config"
	"nexsoft.co.id/nexchief2/constanta"
	"nexsoft.co.id/nexchief2/socket/dtosocket"
)

func (conn *WebSocketClient) subscribe() {
	fileName := "SubscribeDestination.go"
	go conn.subscribeReceiveMessage(fileName)
	go conn.subscribeNotificationSend(fileName)
	go conn.subscribeReceivedDestination(fileName)
	//go conn.subscribePong(fileName)
}

// untuk memperoleh message yang dikirimkan dari sisi client ke nexchief
// message kemudian di handle --> kemudian di proses
// setelah itu nexchief akan memberitahukan bahwa pesan diterima
//
func (conn *WebSocketClient) subscribeReceiveMessage(fileName string) {
	funcName := "subscribeReceiveMessage"
	host := config.ApplicationConfiguration.GetSocketServer()
	for {
		ws := conn.Connect()
		if ws == nil {
			continue
		}

		sub, err := ws.Subscribe(host.Destination.Message, stomp.AckClient)
		if err != nil {
			log("Failed To Subscribe destination /user/nexsoft/message", fileName, funcName, constanta.LogLevelError, 400)
			continue
		}

	loopReceive:
		for {
			select {
			case msg := <-sub.C:
				if msg.Err != nil {
					conn.reset()
					break loopReceive
				}
				//soon deleted
				log("Received Message From Socket : "+string(msg.Body), fileName, funcName, constanta.LogLevelInfo, 200)

				//unmarshal message ke model
				var message dtosocket.ReceiveMessage
				_ = json.Unmarshal(msg.Body, &message)

				//send delivered messaged to socket to inform that the message already received by nexchief
				var sendOpt = []func(*frame.Frame) error{
					stomp.SendOpt.Header("X-Token-Nexsoft", conn.token),
					stomp.SendOpt.Header("heart-beat", "25000,25000"),
					stomp.SendOpt.Header("Cookie", "SOCKETSERVER="+conn.cookie),
				}

				bodyMessage := dtosocket.DeliveredMessage{
					SmsgID: message.SmsgID,
				}

				sendBody, _ := json.Marshal(bodyMessage)
				err = ws.Send(host.Destination.SendDelivered, "application/json;charset=UTF-8", sendBody, sendOpt...)
				if err != nil {
					log("Failed Send Message To Websocket", fileName, funcName, constanta.LogLevelError, 400)
					//break loopReceive
				}

				//call function untuk handle message
				errs, payload := receiveHandler(message)
				if errs.Error != nil {
					if payload != nil {
						sendErrorReceive, _ := json.Marshal(payload)
						_ = ws.Send(host.Destination.SendUser, "application/json;charset=UTF-8", sendErrorReceive, sendOpt...)
						if err != nil {
							log("Failed To Send Error Message To Websocket", fileName, funcName, constanta.LogLevelError, 400)
							//break loopReceive
						}
					} else {
						log("Error Handling Socket Message :"+errs.CausedBy.Error(), errs.FileName, errs.FuncName, constanta.LogLevelError, 400)
					}
					break loopReceive
				}

			case <-conn.ctx.Done():
				log("Subscribe Destination Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
				return
			}
		}
	}
}

// untuk mendapatkan status pengiriman berhasil atau tidak dari socket server
// notif didapat langsung saat itu juga
// menandakan apakah terdapat error dalam pengiriman data ke socker server
// status berupa 200, 400 atau lainnya
// unmarshall message ke mode -->
//
func (conn *WebSocketClient) subscribeNotificationSend(fileName string) {
	funcName := "subscribeNotificationSend"
	host := config.ApplicationConfiguration.GetSocketServer()
	for {
		ws := conn.Connect()
		if ws == nil {
			continue
		}
		//subscribe utuk menerima dari socket server (PONG)
		sub, err := ws.Subscribe(host.Destination.NotifSend, stomp.AckClient)
		if err != nil {
			log("Failed To Subscribe destination /user/nexsoft/notifsend", fileName, funcName, constanta.LogLevelError, 400)
			continue
		}

	loop:
		for {
			select {
			case msg := <-sub.C:
				if msg.Err != nil {
					conn.reset()
					break loop
				}
				log("Received Notif_Send From Socket : "+string(msg.Body), fileName, funcName, constanta.LogLevelInfo, 200)

				//unmarshal message ke model
				var message dtosocket.NotifReceive
				_ = json.Unmarshal(msg.Body, &message)

				//buat function untuk notificationsend handler
				if message.Code == 200 {
					//todo action untuk message yang berhasil disimpan apa ?
					// apa perlu ditambahkan field atau tabel tersendiri untuk menyimpan status telah diterima client tujuan atau belum
				} else {
					//todo action apa jika data gagal diterima oleh socketserver  ?
					//
				}

			case <-conn.ctx.Done():
				log("Subscribe Destination Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
				return
			}
		}
	}
}

// untuk mendapatkan informasi apakah data yang dikirim ke socket server sudah diterima atau belum
// informasi status belum akan diterima jika message belum sampai pada client tujuan
// jika client tujuan dalam kondisi mati, maka nexchief akan mendapatkan messaga status ini
// ketika client tujuan sudah terhubung kembali dgn socket server dan pesan antrian tersebut sudah masuk disisi client
// message di unmarshal ke model -> socketDTO.ReceiveDelivered
func (conn *WebSocketClient) subscribeReceivedDestination(fileName string) {
	funcName := "subscribeReceivedDestination"
	host := config.ApplicationConfiguration.GetSocketServer()
	for {
		ws := conn.Connect()
		if ws == nil {
			continue
		}

		//subscribe utuk menerima dari socket server (PONG)
		sub, err := ws.Subscribe(host.Destination.Status, stomp.AckClient)
		if err != nil {
			log("Failed To Subscribe destination /user/nexsoft/status", fileName, funcName, constanta.LogLevelError, 400)
			continue
		}

	loopReceive:
		for {
			select {
			case msg := <-sub.C:
				if msg.Err != nil {
					conn.reset()
					break loopReceive
				}
				log("Received Status From Socket : "+string(msg.Body), fileName, funcName, constanta.LogLevelInfo, 200)

				var message dtosocket.ReceiveDelivered
				_ = json.Unmarshal(msg.Body, &message)

				//send delivered messaged to socket to inform that the message already received by nexchief
				var sendOpt = []func(*frame.Frame) error{
					stomp.SendOpt.Header("X-Token-Nexsoft", conn.token),
					stomp.SendOpt.Header("heart-beat", "25000,25000"),
					stomp.SendOpt.Header("Cookie", "SOCKETSERVER="+conn.cookie),
				}

				bodyMessage := dtosocket.DeliveredMessage{SmsgID: message.MsgId}
				sendBody, _ := json.Marshal(bodyMessage)

				err = ws.Send("/nexsocket/deliveredreceive", "application/json;charset=UTF-8", sendBody, sendOpt...)
				if err != nil {
					log("Failed Send Message To Websocket", fileName, funcName, constanta.LogLevelError, 400)
					break loopReceive
				}
			case <-conn.ctx.Done():
				log("Subscribe Destination Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
				return
			}
		}
	}
}

// untuk memperoleh respon berupa "pong" dari socket server
// yang menandakan bahwa nexchief sedang dan sudah terkoneksi dengan socket server
// message akan diperoleh sesaat setelah nexchief mengirimkan "ping" ke socket server
// message di unmarshal ke model -> socketDTO.Pong
func (conn *WebSocketClient) subscribePong(fileName string) {
	funcName := "subscribePong"
	host := config.ApplicationConfiguration.GetSocketServer()
	for {
		ws := conn.Connect()
		if ws == nil {
			continue
		}
		//subscribe utuk menerima dari socket server (PONG)
		sub, err := ws.Subscribe(host.Destination.Pong, stomp.AckClient)
		if err != nil {
			log("Failed To Subscribe destination /user/pong", fileName, funcName, constanta.LogLevelError, 400)
			continue
		}
	loop:
		for {
			select {
			case msg := <-sub.C:
				if msg.Err != nil {
					conn.reset()
					break loop
				}
				fmt.Println("message pong : ", string(msg.Body))
			case <-conn.ctx.Done():
				log("Subscribe Destination Stopped", fileName, funcName, constanta.LogLevelInfo, 200)
				return
			}
		}
	}
}
