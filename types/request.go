package types

const (
	TestClosedReqId             = -1
	TestConnectReqId            = 0
	HeartbeatReqId              = 1
	InitTableReqId              = 5
	JoinRoomReqId               = 10
	StartGameReqId              = 11
	ConfirmCancelReqId          = 12
	OutCardReqId                = 33
	DecidedCardTypeReqId        = 40
	AgainOneGameReqId           = 44
	ClientReconnectionReqId     = 45
	GetRemainTableIdReqId       = 55
	ExitRoomReqId               = 66
	TributeGetCardReqId         = 77
	BackGetCardReqId            = 88
	TakePictureUploadCardsReqId = 100
	IsOnePTributeReqId          = 102
	SetHandCardReqId            = 110
)

const (
	PokerDetectUri   = "http://172.10.50.71:8000/poker_detect"
	PokerDecisionUri = "http://172.10.50.71:8400/poker_decision"
)

type AiWeWsReq struct {
	Id   int         `form:"id" json:"id"`
	Data interface{} `form:"data" json:"data"`
}

type HandCardAnalysisReq struct {
	UserName      string `form:"user_name" json:"user_name"`
	HandCardPhoto string `form:"hand_card_photo" json:"hand_card_photo"`
}

type AllResultReq struct {
	CurRank       string   `form:"curRank" json:"curRank"`             //牌级1，2，3，4
	Stage         int      `form:"stage" json:"stage"`                 //当前状态，还贡，最先出牌，接队友牌，接对手牌(0(进贡)， 1（还贡）， 2（先出）， 3（接对手牌）， 4（接队友牌）)
	HandCards     []string `form:"handCards" json:"handCards"`         //手上的牌
	ActionList    []Action `form:"actionList" json:"actionList"`       //所有出牌的可选项
	GreaterAction Action   `form:"greaterAction" json:"greaterAction"` //这一轮的最大牌
}

type Action struct {
	Type  string   `form:"type" json:"type"`   //手牌的类型
	Cards []string `form:"cards" json:"cards"` //这一手牌的所有牌
}

type SetHandCardReq struct {
	UserInfo UserInfo `form:"user_info" json:"user_info"`
	Cards    []string `form:"cards" json:"cards"`
}
