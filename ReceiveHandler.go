package socket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//"nexsoft.co.id/nexchief2/constanta"
	"nexsoft.co.id/nexchief2/model/applicationModel"
	"nexsoft.co.id/nexchief2/model/errorModel"
	//"nexsoft.co.id/nexchief2/resource_common_service/model"
	"nexsoft.co.id/nexchief2/constanta"
	"nexsoft.co.id/nexchief2/resource_common_service/model"
	"nexsoft.co.id/nexchief2/serverconfig"
	"nexsoft.co.id/nexchief2/service/JBControlSocket/BackDateToService/RecieveBackDateFromSocket"
	"nexsoft.co.id/nexchief2/service/JBControlSocket/ChangeND6SysadminPasswordService/ReceiveChangeND6SysadminPasswordFromSocket"
	"nexsoft.co.id/nexchief2/service/JBControlSocket/EditND6UserService/ReceiveND6UserFromSocket"

	//"nexsoft.co.id/nexchief2/service/ProcessSocketDataService"
	"nexsoft.co.id/nexchief2/socket/dtosocket"
	"runtime/debug"
)

// fungsi untuk handle message yang masuk dari socket server
// fungsi ini memanggil service untuk memproses data
// data akan diproses sesuai tipe dari message yang diterima
func receiveHandler(msg dtosocket.ReceiveMessage) (err errorModel.ErrorModel, payloadError *dtosocket.SendMessage) {
	var nexsoftMessage dtosocket.SocketMessage
	_ = json.Unmarshal([]byte(msg.Msg), &nexsoftMessage)

	//create contextModel
	contextModel := &applicationModel.ContextModel{
		AuthAccessTokenModel: model.AuthAccessTokenModel{
			RedisAuthAccessTokenModel: model.RedisAuthAccessTokenModel{
				ResourceUserID: constanta.ResourceIdSystem,
			},
			ClientID: constanta.ClientIdSystem,
		},
	}

	//validate message
	err = validateReceiveMessage(msg)
	if err.Error != nil {
		return
	}

	//
	//switch msg.Type {
	//case "POST_VENDOR_DATA":
	//	//call function to process data nexseller vendor
	//	err, payloadError = processToDB(nexsoftMessage, contextModel, ProcessSocketDataService.ProcessSocketDataService.MappingDataVendor)
	//case "POST_SALESMAN_DATA":
	//	//call function to process data salesman
	//	err, payloadError = processToDB(nexsoftMessage, contextModel, ProcessSocketDataService.ProcessSocketDataService.MappingDataSalesman)
	//case "POST_CUSTOMER_DATA":
	//	//call function to process data nexseller customer
	//	err, payloadError = processToDB(nexsoftMessage, contextModel, ProcessSocketDataService.ProcessSocketDataService.MappingDataCustomer)
	//case "POST_PRODUCT_DATA":
	//	//call function to process nexseller product
	//	err, payloadError = processToDB(nexsoftMessage, contextModel, ProcessSocketDataService.ProcessSocketDataService.MappingDataProduct)
	//}
	//
	switch msg.Type {
	case "RESP_REQ_SYSADMIN_DATA":
		err, payloadError = processToDB(nexsoftMessage, contextModel, ReceiveChangeND6SysadminPasswordFromSocket.UpdateRequestSysadminToJobProcess)
	case "RESP_REQ_RESET_SYSADMIN_PASS":
		err, payloadError = processToDB(nexsoftMessage, contextModel, ReceiveChangeND6SysadminPasswordFromSocket.UpdateRequestChangeSysadminPasswordToJobProcess)
	case "RESP_REQ_DATE_ND6":
		err, payloadError = processToDB(nexsoftMessage, contextModel, RecieveBackDateFromSocket.UpdateRequestDateND6ToJobProcess)
	case "RESP_REQ_BACK_DATE_ND6":
		err, payloadError = processToDB(nexsoftMessage, contextModel, RecieveBackDateFromSocket.UpdateRequestBackDateND6ToJobProcess)
	case "RESP_REQ_USER_DATA":
		err, payloadError = processToDB(nexsoftMessage, contextModel, ReceiveND6UserFromSocket.UpdateRequestND6UserToJobProcess)
	case "RESP_REQ_CHANGE_USER":
		err, payloadError = processToDB(nexsoftMessage, contextModel, ReceiveND6UserFromSocket.UpdateRequestChangeND6UserToJobProcess)
	}

	if err.Error != nil {
		return
	}
	return
}

// wrap function untuk handle service ke tiap mapping data
func processToDB(data dtosocket.SocketMessage, contextModel *applicationModel.ContextModel, serve func(*sql.Tx, dtosocket.SocketMessage, applicationModel.ContextModel) (errorModel.ErrorModel, *dtosocket.SendMessage)) (err errorModel.ErrorModel, payloadError *dtosocket.SendMessage) {
	funcName := "insertToNexcoreTable"
	var errs error
	var tx *sql.Tx

	defer func() {
		if r := recover(); r != nil {
			err = errorModel.GenerateRecoverError()
			err.CausedBy = fmt.Errorf(string(debug.Stack()))
		} else if errs != nil || err.Error != nil {
			errs = tx.Rollback()
			if errs != nil {
				err = errorModel.GenerateInternalDBServerError("ReceiveHandler.go", funcName, errs)
				return
			}
		} else {
			errs = tx.Commit()
			if errs != nil {
				err = errorModel.GenerateInternalDBServerError("ReceiveHandler.go", funcName, errs)
				return
			}
		}
	}()

	tx, errs = serverconfig.ServerAttribute.DBConnection.Begin()
	if errs != nil {
		return
	}

	err, payloadError = serve(tx, data, *contextModel)
	if err.Error != nil {
		return
	}

	err = errorModel.GenerateNonErrorModel()
	return
}

// fungsi validasi untuk message dari socket server
func validateReceiveMessage(message dtosocket.ReceiveMessage) (err errorModel.ErrorModel) {
	// validasi message tidak boleh kosong
	if message.Msg == "" {
		err = errorModel.GenerateEmptyFieldError("ReceiveHandler", "validateReceiveMessage", "Message")
		return
	}
	// todo tambahkan validasi lain jika ada
	return
}
