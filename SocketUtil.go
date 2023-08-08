package socket

import (
	"nexsoft.co.id/nexchief2/config"
	"nexsoft.co.id/nexchief2/model/applicationModel"
	"nexsoft.co.id/nexchief2/util/logUtil"
)

// Untuk mencetak Log File
// --
func log(message string, fileName string, funcName string, logLevel int, statusCode int) {
	logModel := applicationModel.GenerateLogModel("-", config.ApplicationConfiguration.GetServerResourceID())
	logModel.Message = message
	logModel.Status = statusCode
	logModel.Class = "[" + fileName + "," + funcName + "]"
	logUtil.PrintLogWithLevel(logLevel, logModel)
}

