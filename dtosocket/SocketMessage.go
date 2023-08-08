package dtosocket

// Struct Body Message yang dikirimkan oleh client lain ke Nexchief
type SocketMessage struct {
	NexsoftMessage NexsoftMessage `json:"nexsoft_message"`
}

type NexsoftMessage struct {
	Header  Header  `json:"header"`
	Payload Payload `json:"payload"`
}

type Header struct {
	MessageID   string `json:"message_id"`
	UserID      string `json:"user_id"`
	Password    string `json:"password"`
	Version     string `json:"version"`
	PrincipalID string `json:"principal_id"`
	Timestamp   string `json:"timestamp"`
	Action      Action `json:"action"`
}

type Action struct {
	ClassName string `json:"class_name"`
	TypeName  string `json:"type_name"`
}

type Payload struct {
	Header PayloadHeader `json:"header"`
	Data   interface{}   `json:"data"`
}

type PayloadHeader struct {
	Status    int       `json:"status"`
	Size      int       `json:"size"`
	Range     RangeData `json:"range"`
	ReqStatus string    `json:"reqStatus"`
	Message   string    `json:"message"`
}

type RangeData struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type NexposFailedInsertResponseSocket struct {
	MasterName string     `json:"master_name"`
	MasterData MasterData `json:"master_data"`
}

type MasterData struct {
	CompanyID    string `json:"companyID"`
	BranchID     string `json:"branchID"`
	VendorID     string `json:"vendorID"`
	CustomerID   string `json:"customerID"`
	SalesmanID   string `json:"salesmanID"`
	ProductCode  string `json:"productCode"`
	MessageError string `json:"messageError"`
}
