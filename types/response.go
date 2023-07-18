package types

const (
	TestClosedRespId              = -1
	TestConnectRespId             = 0
	HeartbeatRespId               = 1
	IllegalParameterIdRespId      = 2
	WebsocketClientClosed         = 3
	JoinRoomRespId                = 10
	StartGameRespId               = 11
	StartGameAgainRespId          = 12
	ClientReconnectionRespId      = 14
	ConfirmCancelRespId           = 15
	SendCardsRespId               = 22
	TributeGetCardRespId          = 24 //进贡
	BackGetCardRespId             = 25 //还贡
	TributeNoticeRespId           = 26 //进贡信息通知
	BackNoticeRespId              = 27 //还贡信息通知
	OutCardRespId                 = 33
	SendOutCardRespId             = 34
	NoticeNextPlayerOutCardRespId = 35
	ClearDeskTopRespId            = 36
	OutAllCardSortRespId          = 37
	EndOneGameRespId              = 38
	CardTypeConfirmRespId         = 39
	NoticeWhoOutCardRespId        = 40 //通知谁正在出牌
	GetRemainTableIdRespId        = 55
	ExitRoomRespId                = 66
	TributeGetCardNRespId         = 77 //进贡牌是否符合通知
	TributeGetCardIRespId         = 78 //进贡返回牌
	BackGetCardNRespId            = 88 //还贡是否符合通知
	BackGetCardIRespId            = 89 //还贡返回牌
	TributeMaxOutCardRespId       = 90 //进贡最大的牌出第一手牌

	TakePictureUploadCardsRespId = 100 //拍照上传手牌
	TPUploadCardsRespId          = 101
	IsOnePTributeRespId          = 102
	A1A2TributeBackCardRespId    = 103 //通知A1或A2该进贡或还贡什么牌
	SetHandCardRespId            = 110
)

type JoinRoomResp struct {
	UserName  string `form:"user_name" json:"user_name"`   //用户名称
	IsEntered bool   `form:"is_entered" json:"is_entered"` //此角色是否已经进入
}

type SendCardsResp struct {
	IsOutCard      int    `form:"is_out_card" json:"is_out_card"`             //是否第一个出牌;0:不出牌，1:出牌
	AllIsEnterRoom bool   `form:"all_is_enter_room" json:"all_is_enter_room"` //此牌桌的所有玩家是否都进入并准备好；false：未准备好；true：已准备好
	LCardNumber    string `form:"l_card_number" json:"l_card_number"`         //此局通配牌牌点
	LCardNumberP   string `form:"l_card_number_p" json:"l_card_number_p"`     //P方打的通配牌牌点
	LCardNumberA   string `form:"l_card_number_a" json:"l_card_number_a"`     //A方打的通配牌牌点
	Cards          Cards  `form:"cards" json:"cards"`
	IsResist       bool   `form:"is_resist" json:"is_resist"` //是否存在抗贡;false:不抗贡，true:抗贡
}

// SendCardsAgainResp
// @Description: 再一次进入发牌
type SendCardsAgainResp struct {
	UserInfo         UserInfo `form:"user_info" json:"user_info"`                 //此玩家的用户信息
	IsOutCard        bool     `form:"is_out_card" json:"is_out_card"`             //是否出牌;false:不出牌，true:出牌
	IsMustOutCard    bool     `form:"is_must_out_card" json:"is_must_out_card"`   //是否必须出牌;false:不必须，true:必须
	IsOneMoreGame    bool     `form:"is_one_more_game" json:"is_one_more_game"`   //是否再来一局;false:不再来一局，true:再来一局
	AllIsEnterRoom   bool     `form:"all_is_enter_room" json:"all_is_enter_room"` //此牌桌的所有玩家是否都进入并准备好；false：未准备好；true：已准备好
	LCardNumber      string   `form:"l_card_number" json:"l_card_number"`         //通配牌牌点
	LCardNumberP     string   `form:"l_card_number_p" json:"l_card_number_p"`     //P方打的通配牌牌点
	LCardNumberA     string   `form:"l_card_number_a" json:"l_card_number_a"`     //A方打的通配牌牌点
	Cards            Cards    `form:"cards" json:"cards"`
	TributeBackState int      `form:"tribute_back_state" json:"tribute_back_state"` //进贡还贡状态；0:不进贡还贡；1：进贡；2还贡
}

type SendOutCardResp struct {
	UserInfo       UserInfo `form:"user_info" json:"user_info"`
	Cards          Cards    `form:"cards" json:"cards"`
	RemainCardsNum int      `form:"remain_cards_num" json:"remain_cards_num"`
}

type EndOneGameResp struct {
	RankingInfos []RankingInfo `form:"ranking_infos" json:"ranking_infos"` //一局结束后玩家排名信息
	Score        int           `form:"score" json:"score"`                 //上游的一方此局得分
}

type RankingInfo struct {
	Ranking  int     `form:"ranking" json:"ranking"` //排名
	UserId   UserId  `form:"user_id" json:"user_id"`
	UserName string  `form:"user_name" json:"user_name"`
	UserRole int     `form:"user_role" json:"user_role"` // 1 ， 2 ；UserRole一样为一家
	Location int     `form:"location" json:"location"`
	TableId  TableId `form:"table_id" json:"table_id"`
}

type A1A2TributeBackCardResp struct {
	Type int  `form:"type" json:"type"` //0:进贡；1:还贡
	Card Card `form:"card" json:"card"` //进贡或还贡的牌
}
