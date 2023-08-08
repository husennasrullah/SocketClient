package dtosocket

// Struct untuk body request di fungsi login
type SocketLogin struct {
	UserID   string `json:"uID"`
	Password string `json:"pw"`
}

// Struct untuk body message saat melakukan pengiriman ke socketserver
type SendMessage struct {
	Touid     string `json:"toUID"`
	SrcSender string `json:"srcSender"`
	Type      string `json:"type"`
	MsgID     string `json:"msgID"`
	MsgSeq    string `json:"msgSeq"`
	Msg       string `json:"msg"`
}

// Struct body message saat menerima message dari socket
type ReceiveMessage struct {
	SndName      string `json:"sndName"`
	SndUID       string `json:"sndUID"`
	Msg          string `json:"msg"`
	MsgSeq       int32  `json:"msgSeq"`
	Type         string `json:"type"`
	SmsgID       string `json:"smsgID"`
	Source       string `json:"source"`
	MsgID        string `json:"msgID"`
	DeliveryDate string `json:"deliveryDate"`
}

// Struct body message saat pengiriman status bahwa message sudah diterima di nexchief
type DeliveredMessage struct {
	SmsgID string `json:"smsgID"`
	//Token  string `json:"token"`
}

// Struct Body message dari destinasi -> status
type ReceiveDelivered struct {
	RcvName string `json:"rcvName"`
	RcvId   string `json:"rcvID"`
	SmsgId  string `json:"smsgID"`
	MsgId   string `json:"msgID"`
	MsgSeq  string `json:"msgSeq"`
}

// Struct Body message dari destinasi -> pong
type Pong struct {
	Result string `json:"result"`
}

// Struct Body Message dari desinasi -> NotifSend
type NotifReceive struct {
	Code    int    `json:"code"`
	MsgID   string `json:"msgID"`
	MsgSeq  string `json:"msgSeq"`
	Message string `json:"message"`
}

type StatusConnection struct {
	Status string `json:"status"`
}

type ResetConnection struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
