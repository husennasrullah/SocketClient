package socket

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"nexsoft.co.id/nexchief2/config"
	"nexsoft.co.id/nexchief2/constanta"
	"nexsoft.co.id/nexchief2/model/applicationModel"
	"nexsoft.co.id/nexchief2/model/errorModel"
	"nexsoft.co.id/nexchief2/resource_common_service/common"
	"nexsoft.co.id/nexchief2/socket/dtosocket"
	"nexsoft.co.id/nexchief2/util/logUtil"
	util2 "nexsoft.co.id/nexcommon/util"
	"strings"
)

// Fungsi ini digunakan untuk Hit Api login dari socket server
// mendapat keluaran berupa token dan cookie
// token dan cookie digunakan pada header saat melakukan ke koneksi ke socket server
func (conn *WebSocketClient) login() error {
	funcName := "login"
	fileName := "SocketAPI.go"
	var err errorModel.ErrorModel

	socketServer := config.ApplicationConfiguration.GetSocketServer()
	address := url.URL{
		Scheme: socketServer.Protocol[0],
		Host:   socketServer.Host,
		Path:   socketServer.PathRedirect.Login,
	}

	// todo rohmat remove this code
	// cuma buat testing, belum nambah di env
	if strings.Contains(socketServer.UserID, "SOCKET_USERNAME") {
		socketServer.UserID = "NexchiefDev2022"
	}
	if strings.Contains(socketServer.Password, "SOCKET_PASSWORD") {
		socketServer.Password = "Nexchief2022"
	}
	// end remove code

	bodyRequest := dtosocket.SocketLogin{
		UserID:   socketServer.UserID,
		Password: socketServer.Password,
	}
	logger := applicationModel.LoggerModel{Class: "[SocketAPI.go,login]", Message: fmt.Sprintf("Start login socket username: %s, password:%s", socketServer.UserID, socketServer.Password)}
	logUtil.PrintLogWithLevel(constanta.LogLevelDebug, logger)

	headerRequest := make(map[string][]string)
	headerRequest["Content-Type"] = []string{"application/json"}

	//set tls configuration
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	statusCode, _, bodyResult, errorS := common.HitAPI(address.String(), headerRequest, util2.StructToJSON(bodyRequest), "POST", applicationModel.ContextModel{})
	if errorS != nil {
		err = errorModel.GenerateUnknownError(fileName, funcName, errorS)
		return errorS
	}

	var loginResponse map[string]interface{}
	_ = json.Unmarshal([]byte(bodyResult), &loginResponse)

	if statusCode == 200 {
		conn.token = loginResponse["token"].(string)
		conn.cookie = loginResponse["cookie"].(string)
		return errorS
	} else {
		err = errorModel.GenerateAuthenticationServerError(fileName, funcName, statusCode, string(rune(statusCode)), nil)
		return err.Error
	}
}
