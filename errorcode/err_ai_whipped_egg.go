package errorcode

const (
	ErrAWESeEnterRoom = AiWhippedEggErrorBase + 1 + iota
	ErrAWESeNotPlayRules
	ErrAWESeAllIsEnterRoom
	ErrAWESeOutCardsLTMaxCards
	ErrAWESeHaveSomeCardType
	ErrAWESeParameterId
	ErrAWESeUserInfoTAEmpty
	ErrAWESeUserInfoAEmpty
	ErrAWESeTableIdIllegal
	ErrAWESeHandCardsLengthInit
	ErrAWESeTributeGetCard
	ErrAWESeBackGetCard
	ErrAWESeGetCardInfoIllegal
	ErrAWESeUserInfoAEmptyC
	ErrAWESeTableIdIllegalC
	ErrAWESeNotExistUserInfo
	ErrAWESeWebSocktReuse
	ErrAWESeWSConnectionSuccessful
	ErrAWESeWSClosedSuccessful
	ErrAWESeHCAReqEmpty
	ErrAWESePokerDetectCall
	ErrAWESeAiHandCardLen
	ErrAWESePokerDecisionCall
	ErrAWESeGetTributeCardLength
	ErrAWESeGetBackCardLength
)

func init() {
	RegisterErrorCode(errAiWhippedEgg)
}

var errAiWhippedEgg = map[int]string{
	ErrAWESeEnterRoom:              "玩家进入Ai掼蛋牌桌失败，已存在此玩家",
	ErrAWESeNotPlayRules:           "识别到的牌不符合出牌规则,请重新识别",
	ErrAWESeAllIsEnterRoom:         "此牌桌的所有玩家还未都进入并准备好,请先进入房间牌桌",
	ErrAWESeOutCardsLTMaxCards:     "出的手牌小于这一轮的最大牌，请重新出牌",
	ErrAWESeHaveSomeCardType:       "此牌存在多种组合，请选择出哪种牌",
	ErrAWESeParameterId:            "非法入参id,请求拒绝！",
	ErrAWESeUserInfoTAEmpty:        "入参userInfo.TableId属性为空，请检查入参",
	ErrAWESeUserInfoAEmpty:         "入参userInfo有属性为空，请检查入参",
	ErrAWESeTableIdIllegal:         "入参tableId是非法的，不存在此桌",
	ErrAWESeUserInfoAEmptyC:        "客户端userInfo信息为空",
	ErrAWESeTableIdIllegalC:        "客户端tableId信息是非法的，不存在此桌",
	ErrAWESeHandCardsLengthInit:    "玩家手牌张数初始化错误",
	ErrAWESeTributeGetCard:         "玩家进贡的牌不符合要求",
	ErrAWESeBackGetCard:            "玩家还贡的牌不符合要求",
	ErrAWESeGetCardInfoIllegal:     "GetCardInfo入参不合法",
	ErrAWESeNotExistUserInfo:       "该桌不存在此用户信息,有可能是未登录或被初始化,请先登陆",
	ErrAWESeWebSocktReuse:          "WebSockt重复使用，一个WebSockt只能服务一个玩家",
	ErrAWESeWSConnectionSuccessful: "WebSockt建立连接成功啦",
	ErrAWESeWSClosedSuccessful:     "WebSockt关闭成功啦",
	ErrAWESeHCAReqEmpty:            "入参HandCardAnalysisReq有属性为空，请检查入参",
	ErrAWESePokerDetectCall:        "调用Ai手牌分析接口/poker_detect失败",
	ErrAWESeAiHandCardLen:          "手牌分析结果张数不对，请重新上传",
	ErrAWESePokerDecisionCall:      "调用Ai手牌决策接口/poker_decision失败",
	ErrAWESeGetTributeCardLength:   "玩家进贡的牌张数错误",
	ErrAWESeGetBackCardLength:      "玩家还贡的牌张数错误",
}
