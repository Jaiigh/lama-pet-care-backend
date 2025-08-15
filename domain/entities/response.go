package entities

type ResponseMessage struct {
	Message string `json:"message"`
}

type ResponseModel struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Status  int         `json:"status"`
}

type ResponseBool struct {
	Message string `json:"message"`
	IsTrue  bool   `json:"istrue"`
}
