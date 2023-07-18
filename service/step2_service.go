package service

import (
	"ai-guandan/errorcode"
	"ai-guandan/types"
	"ai-guandan/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

// FirstEntryMethod2 @description: 初次进入游戏逻辑处理及消息返回
// @parameter table
// @parameter c
// @parameter userInfo
func FirstEntryMethod2(table *types.Table, c *types.Client, userInfo types.UserInfo) {
	c.UserInfo = userInfo
	c.Ready = true
	table.TableClients[userInfo.UserName] = c
	//返回开始游戏成功信息
	c.WsSendMsg(types.StartGameRespId, errorcode.Successfully, userInfo)
	if 4 == len(table.TableClients) {
		tempState := true
		//判断四个人是否都准备好
		for _, v := range table.TableClients {
			tempState = tempState && v.Ready
		}
		//如果都准备好就开始下达拍照上传手牌指令
		if tempState {
			//初始化牌桌
			table.InitOneTableStart2()
			table.GameManage.AllIsEnterRoom = true
			table.TableClients[types.Ai1].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
			table.TableClients[types.Ai2].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
		}
	}
}

func FirstEntryMethod3(table *types.Table, c *types.Client, setHandCardReq types.SetHandCardReq) {
	c.UserInfo = setHandCardReq.UserInfo
	c.Ready = true
	//手牌转换
	c.PlayerCardsInfo.CardList, c.PlayerCardsInfo.LCard = ListStringToCards(setHandCardReq.Cards, table.GameManage.LevelCardPoint)
	table.TableClients[setHandCardReq.UserInfo.UserName] = c
	//返回开始游戏成功信息
	c.WsSendMsg(types.SetHandCardRespId, errorcode.Successfully, nil)
	if 4 == len(table.TableClients) {
		tempState := true
		//判断四个人是否都准备好
		for _, v := range table.TableClients {
			tempState = tempState && v.Ready
		}
		//如果都准备好就开始发牌
		if tempState {
			//初始化牌桌
			table.InitOneTableStart3()
			table.GameManage.AllIsEnterRoom = true
			tempUserInfo := table.TableClients[table.GameManage.FirstOutCard].UserInfo
			for k, v := range table.TableClients {
				if k == table.GameManage.FirstOutCard {
					//通知谁出牌
					v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, 1)
				} else {
					//通知其他玩家谁在出牌
					v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, tempUserInfo)
				}
			}
		}
	}
}

// TakePictureUploadCardsMethod @description: 拍照上传手牌逻辑处理
// @parameter req
// @parameter c
func TakePictureUploadCardsMethod(req types.AiWeWsReq, c *types.Client) {
	handCardAnalysisReq := new(types.HandCardAnalysisReq)
	code := utils.ObjectAToObjectB(req.Data, handCardAnalysisReq)
	if errorcode.Successfully != code {
		c.WsSendMsg(types.TPUploadCardsRespId, code, nil)
		return
	}
	//入参校验
	if 0 == len(handCardAnalysisReq.UserName) || 0 == len(handCardAnalysisReq.HandCardPhoto) {
		c.WsSendMsg(types.TPUploadCardsRespId, errorcode.ErrAWESeHCAReqEmpty, nil)
		return
	}
	//调用ai手牌分析接口，获取玩家手牌
	aiHandCards := make([]string, 0)
	log.Printf("userInfo:%+v,callUrl:%v,req:%v", c.UserInfo, types.PokerDetectUri, len(handCardAnalysisReq.HandCardPhoto))
	if code = utils.CommonPostCall("", types.PokerDetectUri, handCardAnalysisReq, &aiHandCards); errorcode.Successfully != code {
		c.WsSendMsg(types.TPUploadCardsRespId, errorcode.ErrAWESePokerDetectCall, nil)
		return
	}
	log.Printf("userInfo:%+v,callUrl:%v,resp:%+v", c.UserInfo, types.PokerDetectUri, aiHandCards)
	if 27 != len(aiHandCards) {
		c.WsSendMsg(types.TPUploadCardsRespId, errorcode.ErrAWESeAiHandCardLen, nil)
		return
	}
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	//手牌转换
	c.PlayerCardsInfo.CardList, c.PlayerCardsInfo.LCard = ListStringToCards(aiHandCards, c.PlayerCardsInfo.LCardNumber)
	sort.Sort(c.PlayerCardsInfo.CardList)
	//下发Ai识别的发牌
	ADealCard(*table, *c)
	//找队友
	teammateName := FindTeammate(c.UserInfo.UserName)
	//A1,A2都下发好牌之后
	if 0 != table.TableClients[teammateName].PlayerCardsInfo.CardList[0].Id {
		A1A2DealCard(table)
	}
}

// A1A2DealCard @description: A1,A2都下发好牌之后处理
// @parameter table
func A1A2DealCard(table *types.Table) {
	if 0 == len(table.GameManage.NextTributeUsers) { //判断是否是第一把
		NotifyWhoIsPlaying(*table, table.TableClients[table.GameManage.FirstOutCard].UserInfo, 1)
	} else {
		if IsOnePTribute(*table) { //单独P1或P2进贡，无法判断是否抗贡，需要玩家自己确认，发送给A1客户端需要P1确认是否抗贡的消息
			table.TableClients[types.Ai1].WsSendMsg(types.IsOnePTributeRespId, errorcode.Successfully, table.GameManage.NextTributeUsers[0])
			return
		}
		if IsResist2(*table) || types.PokerViewNum2 == table.GameManage.LevelCardPoint {
			if types.PokerViewNum2 != table.GameManage.LevelCardPoint {
				var tempStr string
				if 1 == len(table.GameManage.NextTributeUsers) {
					tempStr = fmt.Sprintf("%s玩家抗贡了", table.GameManage.NextTributeUsers[0].UserName)
				}
				if 2 == len(table.GameManage.NextTributeUsers) {
					tempStr = fmt.Sprintf("%s,%s玩家抗贡了", table.GameManage.NextTributeUsers[0].UserName, table.GameManage.NextTributeUsers[1].UserName)
				}
				for _, v := range table.TableClients {
					//通知抗贡情况
					v.WsSendMsg(types.TPUploadCardsRespId, errorcode.Successfully, tempStr)
				}
			}
			ResistanceGongDeal(table)
			return
		}
		for _, user := range table.GameManage.NextTributeUsers {
			table.TableClients[user.UserName].WsSendMsg(types.TributeGetCardRespId, errorcode.Successfully, "此局你要进贡啦")
			for _, tc := range table.TableClients {
				tc.WsSendMsg(types.TributeNoticeRespId, errorcode.Successfully, user)
			}
		}
		for _, user := range table.GameManage.NextBackUsers {
			table.TableClients[user.UserName].WsSendMsg(types.BackGetCardRespId, errorcode.Successfully, "此局你要还贡啦")
			for _, tc := range table.TableClients {
				tc.WsSendMsg(types.BackNoticeRespId, errorcode.Successfully, user)
			}
		}
	}
}

// IsOnePTributeMethod @description: P1或P2玩家反馈是否抗贡逻辑处理
// @parameter req
// @parameter c
func IsOnePTributeMethod(req types.AiWeWsReq, c *types.Client) {
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	if req.Data.(bool) { //是否抗贡
		var tempStr string
		if 1 == len(table.GameManage.NextTributeUsers) {
			tempStr = fmt.Sprintf("%s玩家抗贡了", table.GameManage.NextTributeUsers[0].UserName)
		}
		if 2 == len(table.GameManage.NextTributeUsers) {
			tempStr = fmt.Sprintf("%s,%s玩家抗贡了", table.GameManage.NextTributeUsers[0].UserName, table.GameManage.NextTributeUsers[1].UserName)
		}
		for _, v := range table.TableClients {
			//通知抗贡情况
			v.WsSendMsg(types.TPUploadCardsRespId, errorcode.Successfully, tempStr)
		}
		ResistanceGongDeal(table)
	} else {
		for _, user := range table.GameManage.NextTributeUsers {
			table.TableClients[user.UserName].WsSendMsg(types.TributeGetCardRespId, errorcode.Successfully, "此局你要进贡啦")
		}
		for _, user := range table.GameManage.NextBackUsers {
			table.TableClients[user.UserName].WsSendMsg(types.BackGetCardRespId, errorcode.Successfully, "此局你要还贡啦")
		}
	}
}

// InitTableMethod @description: 初始化table牌桌逻辑处理
// @parameter c
func InitTableMethod(c *types.Client) {
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	table.GameManage = *new(types.GameManage)
	types.InitTMapViewToLevel(&table.GameManage.MapViewToLevelH)
	table.GameManage.LevelCardPoint = "2"  //最开始默认打2
	table.GameManage.LevelCardPointP = "2" //P方最开始打2
	table.GameManage.LevelCardPointA = "2" //A方最开始打2
	for k, _ := range table.TableClients {
		if types.Ai1 == k || types.Ai2 == k {
			delete(table.TableClients, k)
			continue
		}
		table.TableClients[k].PlayerCardsInfo = *new(types.PlayerCardsInfo)
		table.TableClients[k].Next = nil
	}
	log.Warnf("userInfo:%+v,tableId:%v table has inited", c.UserInfo, c.UserInfo.TableId)
}

// SetHandCardMethod @description: 设置玩家手牌逻辑处理
// @parameter req
// @parameter c
func SetHandCardMethod(req types.AiWeWsReq, c *types.Client) {
	setHandCardReq := &types.SetHandCardReq{}
	code := utils.ObjectAToObjectB(req.Data, setHandCardReq)
	if errorcode.Successfully != code {
		c.WsSendMsg(types.SetHandCardRespId, code, nil)
		return
	}
	if 27 != len(setHandCardReq.Cards) {
		c.WsSendMsg(types.SetHandCardRespId, errorcode.ErrAWESeAiHandCardLen, nil)
		return
	}
	//入参用户信息校验
	if code := CheckUserInfoParameter(setHandCardReq.UserInfo); errorcode.Successfully != code {
		c.WsSendMsg(types.SetHandCardRespId, code, nil)
		return
	}
	table := types.RoomManagerObject.Rooms[setHandCardReq.UserInfo.RoomId].Tables[setHandCardReq.UserInfo.TableId]
	if _, ok := table.TableClients[setHandCardReq.UserInfo.UserName]; ok && table.TableClients[setHandCardReq.UserInfo.UserName].Ready {
		log.Warnf("%v已进入", setHandCardReq.UserInfo.UserName)
		//返回设置玩家手牌失败信息
		c.WsSendMsg(types.SetHandCardRespId, errorcode.ErrAWESeEnterRoom, nil)
		return
	}
	//初次进入
	FirstEntryMethod3(table, c, *setHandCardReq)
}

// ADealCard @description: 下发Ai识别的发牌
// @parameter table
// @parameter c
// @parameter firstOutCard
func ADealCard(table types.Table, c types.Client) {
	sendCardsResp := new(types.SendCardsResp)
	sendCardsResp.LCardNumber = c.PlayerCardsInfo.LCardNumber
	sendCardsResp.LCardNumberP = table.GameManage.LevelCardPointP
	sendCardsResp.LCardNumberA = table.GameManage.LevelCardPointA
	sendCardsResp.Cards = c.PlayerCardsInfo.CardList
	//sendCardsResp.IsResist = true
	sendCardsResp.AllIsEnterRoom = table.GameManage.AllIsEnterRoom
	c.WsSendMsg(types.SendCardsRespId, errorcode.Successfully, *sendCardsResp)
}

// NotifyWhoIsPlaying @description: 通知谁正在出牌，谁出牌
// @parameter table
// @parameter firstOutCard
// @parameter state(是否必须出牌)
func NotifyWhoIsPlaying(table types.Table, firstOutCard types.UserInfo, state int) {
	for k, v := range table.TableClients {
		if k == firstOutCard.UserName {
			//通知谁出牌
			v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, state)
		} else {
			//通知其他玩家谁在出牌
			v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, firstOutCard)
		}
	}
}

// OutCardMethod2 @description: 出牌步骤逻辑处理-Ai
// @parameter req
// @parameter c
func OutCardMethod2(req types.AiWeWsReq, c *types.Client) {
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	if table.GameManage.AllIsEnterRoom != true {
		log.Errorln("此牌桌的所有玩家还未都进入并准备好")
		c.WsSendMsg(types.OutCardRespId, errorcode.ErrAWESeAllIsEnterRoom, nil)
		return
	}
	cards := new(types.Cards)
	//获取玩家出的牌
	code := GetPlayerOutCards(*table, *c, cards, req)
	if errorcode.Successfully != code {
		for _, v := range table.TableClients {
			v.WsSendMsg(types.OutCardRespId, code, c.UserInfo)
		}
		return
	}
	dealCardStr := "不出牌"
	//出牌时
	if 0 != len(*cards) {
		//判断出的牌是否符合规则
		if code = CardsNotNUllMethod(table, c, cards); errorcode.Successfully != code {
			for k, v := range table.TableClients {
				if k == c.UserInfo.UserName {
					//识别有可能错误，通知它继续出牌
					v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, 1)
				} else {
					//通知其他玩家此玩家出的牌不符合规则
					v.WsSendMsg(types.OutCardRespId, code, c.UserInfo)
				}
			}
			return
		}
		dealCardStr = "出牌成功"
	}
	//出牌或不出牌处理消息通知
	OutCardNotice(table, c, *cards, dealCardStr)
}

// AllResultReqStageJudge @description: AllResultReq.Stage判断
// @parameter table
// @parameter c
// @return int
func AllResultReqStageJudge(table types.Table, c types.Client) int {
	stage := 3
	if 0 == len(table.GameManage.MaxHandCardPlayer) || 0 == table.GameManage.MaxHandCard.A || c.UserInfo.UserName == table.GameManage.MaxHandCardPlayer {
		stage = 2
		return stage
	}
	if c.UserInfo.UserName[0:1] == table.GameManage.MaxHandCardPlayer[0:1] {
		stage = 4
		return stage
	}
	return stage
}

// GetPlayerOutCards @description: 获取玩家出的牌
// @parameter table
// @parameter c
// @return *types.Cards
func GetPlayerOutCards(table types.Table, c types.Client, cards *types.Cards, req types.AiWeWsReq) int {
	code := errorcode.Successfully
	code, *cards = A1A2P1P2OutCards(table, c, req, AllResultReqStageJudge(table, c))
	return code
}

// DeleteHandCard2 @description: 删除已经出掉的牌-Ai
// @parameter c
// @parameter handCard
func DeleteHandCard2(c *types.Client, handCard types.HandCard) {
	cards := make([]types.Card, 0)
	if types.Ai1 == c.UserInfo.UserName || types.Ai2 == c.UserInfo.UserName {
		mapTemp := make(map[int]bool)
		lInt := 0
		for _, cs := range handCard.PokerHands {
			for _, card := range cs {
				mapTemp[card.Id] = true
				if isLCard(card, c.PlayerCardsInfo.LCardNumber) {
					lInt++
				}
			}
		}
		if 0 < lInt {
			count := len(c.PlayerCardsInfo.LCard)
			c.PlayerCardsInfo.LCard = c.PlayerCardsInfo.LCard[0:(count - lInt)]
		}
		for _, card := range c.PlayerCardsInfo.CardList {
			if _, ok := mapTemp[card.Id]; ok {
				continue
			}
			cards = append(cards, card)
		}
	}
	count := 0
	if types.Pi1 == c.UserInfo.UserName || types.Pi2 == c.UserInfo.UserName {
		for _, cs := range handCard.PokerHands {
			count = count + len(cs)
		}
		if count > len(c.PlayerCardsInfo.CardList) {
			log.Errorln("DeleteHandCard2 count大于len(c.PlayerCardsInfo.CardList)！")
			return
		}
		cards = c.PlayerCardsInfo.CardList[count:]
	}
	c.PlayerCardsInfo.CardList = cards
}

// AgainOneGameMethod2 @description: 再来一局逻辑处理-Ai
// @parameter req
// @parameter c
func AgainOneGameMethod2(c *types.Client) {
	types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].Lock.RLock()
	defer types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].Lock.RUnlock()
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	c = table.TableClients[c.UserInfo.UserName]
	c.Ready = true
	if 4 == len(table.TableClients) {
		tempState := true
		//判断四个人是否都准备好
		for _, v := range table.TableClients {
			tempState = tempState && v.Ready
		}
		//如果都准备好就开始下达拍照上传手牌指令
		if tempState {
			//初始化牌桌
			table.InitOneTableStart2()
			table.GameManage.AllIsEnterRoom = true
			table.TableClients[types.Ai1].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
			table.TableClients[types.Ai2].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
		}
	}
}

// IsResist2 @description:此局是否出现抗贡-Ai
// @parameter table
// @return bool
func IsResist2(table types.Table) bool {
	if nil == table.GameManage.NextTributeUsers || 0 == len(table.GameManage.NextTributeUsers) {
		log.Errorf("tableId:%v 需要进贡的玩家为空！")
		return false
	}
	if types.Ai1 == table.GameManage.NextTributeUsers[0].UserName || types.Ai1 == table.GameManage.NextTributeUsers[0].UserName {
		return 2 == IsTwoPokerViewNumR(table, table.GameManage.NextTributeUsers)
	}
	return 0 == IsTwoPokerViewNumR(table, table.GameManage.NextBackUsers)
}

// IsTwoPokerViewNumR @description: 进贡或还贡方的大王牌张数是否等于二
// @parameter table
// @parameter userInfos
// @return bool
func IsTwoPokerViewNumR(table types.Table, userInfos []types.UserInfo) int {
	count := 0
	for _, user := range userInfos {
		tempList := table.TableClients[user.UserName].PlayerCardsInfo.CardList
		if 27 == len(tempList) {
			if types.PokerViewNumR == tempList[25].ViewNumber {
				count++
			}
			if types.PokerViewNumR == tempList[26].ViewNumber {
				count++
			}
		}
	}
	return count
}

// IsOnePTribute @description: 此局是否存在只有P方(实体人)一个玩家进贡
// @parameter table
// @return bool
func IsOnePTribute(table types.Table) bool {
	return 1 == len(table.GameManage.NextTributeUsers) && (types.Pi1 == table.GameManage.NextTributeUsers[0].UserName || types.Pi2 == table.GameManage.NextTributeUsers[0].UserName)
}

// ResistanceGongDeal @description: 抗贡处理
// @parameter table
func ResistanceGongDeal(table *types.Table) {
	//初始化上一局进贡还贡玩家的记录
	table.GameManage.NextBackUsers = make([]types.UserInfo, 0)
	table.GameManage.NextTributeUsers = make([]types.UserInfo, 0)
	//初始化上一局进贡还贡牌信息
	table.GameManage.TributeGetCardInfos = make([]types.GetCardInfo, 0)
	table.GameManage.BackGetCardInfos = make([]types.GetCardInfo, 0)
	tempUserInfo := table.TableClients[table.GameManage.FirstOutCard].UserInfo
	for k, v := range table.TableClients {
		if k != table.GameManage.FirstOutCard {
			//通知其他玩家谁出第一手牌
			v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, tempUserInfo)

		} else {
			//通知玩家出第一手牌
			v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, tempUserInfo)
		}
	}
}

// GetTributeCard @description: 获取玩家进贡的牌
// @parameter table
// @parameter c
// @parameter getCardInfo
// @parameter req
// @return int
func GetTributeCard(table types.Table, c types.Client, getCardInfo *types.GetCardInfo, req types.AiWeWsReq) int {
	code, cards := A1A2P1P2OutCards(table, c, req, 0)
	if errorcode.Successfully != code {
		return code
	}
	if 1 != len(cards) {
		return errorcode.ErrAWESeGetTributeCardLength
	}
	getCardInfo.UserInfo = c.UserInfo
	getCardInfo.Card = cards[0]
	if types.Ai1 == c.UserInfo.UserName || types.Ai2 == c.UserInfo.UserName {
		a1A2TributeBackCardResp := new(types.A1A2TributeBackCardResp)
		a1A2TributeBackCardResp.Type = 0
		a1A2TributeBackCardResp.Card = cards[0]
		c.WsSendMsg(types.A1A2TributeBackCardRespId, errorcode.Successfully, *a1A2TributeBackCardResp)
	}
	return code
}

// GetBackCard @description: 获取玩家还贡的牌
// @parameter table
// @parameter c
// @parameter getCardInfo
// @parameter req
// @return int
func GetBackCard(table types.Table, c types.Client, getCardInfo *types.GetCardInfo, req types.AiWeWsReq) int {
	code, cards := A1A2P1P2OutCards(table, c, req, 1)
	if errorcode.Successfully != code {
		return code
	}
	if 1 != len(cards) {
		return errorcode.ErrAWESeGetBackCardLength
	}
	getCardInfo.UserInfo = c.UserInfo
	getCardInfo.Card = cards[0]
	if types.Ai1 == c.UserInfo.UserName || types.Ai2 == c.UserInfo.UserName {
		a1A2TributeBackCardResp := new(types.A1A2TributeBackCardResp)
		a1A2TributeBackCardResp.Type = 1
		a1A2TributeBackCardResp.Card = cards[0]
		c.WsSendMsg(types.A1A2TributeBackCardRespId, errorcode.Successfully, *a1A2TributeBackCardResp)
	}
	return code
}

// A1A2P1P2OutCards @description: A1,A2调用决策接口决定出的或进贡还贡的牌；P1,P2通过识别的结果获取牌
// @parameter table
// @parameter c
// @parameter req
// @return int
// @return []types.Card
func A1A2P1P2OutCards(table types.Table, c types.Client, req types.AiWeWsReq, stage int) (int, []types.Card) {
	cards := make([]types.Card, 0)
	index := -2 //-1不出牌
	if types.Ai1 == c.UserInfo.UserName || types.Ai2 == c.UserInfo.UserName {
		levelCardPoint := table.GameManage.LevelCardPoint
		allResultReq := new(types.AllResultReq)
		allResultReq.CurRank = levelCardPoint
		allResultReq.Stage = stage
		allResultReq.HandCards = CardsToListString(c.PlayerCardsInfo.CardList)
		handCards := make([]types.HandCard, 0)
		if 0 == allResultReq.Stage { //进贡
			handCards = GetTributePokerHandType(table, c)
		}
		if 1 == allResultReq.Stage { //还贡
			handCards = GetBackPokerHandType(table, c)
		}
		if 2 == allResultReq.Stage { //先出(必须出牌)
			handCards = GetAllPokerHandType(c.UserInfo.Location, c.UserInfo.UserName, levelCardPoint, c.PlayerCardsInfo.CardList)
		}
		if 3 == allResultReq.Stage || 4 == allResultReq.Stage { //接别人的牌，可以选择不出
			handCards = GetAllPokerHandTypeBigger(c.UserInfo.Location, c.UserInfo.UserName, levelCardPoint, c.PlayerCardsInfo.CardList, table.GameManage.MaxHandCard, table.GameManage.MapViewToLevelH)
		}
		allResultReq.ActionList = HandCardListToActionList(handCards)
		allResultReq.GreaterAction = HandCardToAction(table.GameManage.MaxHandCard)
		log.Printf("userInfo:%+v,callUrl:%v,req:%+v", c.UserInfo, types.PokerDecisionUri, *allResultReq)
		//调用Ai决策接口
		if code := utils.CommonPostCall("", types.PokerDecisionUri, allResultReq, &index); errorcode.Successfully != code || -2 == index || len(handCards)-1 < index {
			return errorcode.ErrAWESePokerDecisionCall, cards
		}
		log.Printf("userInfo:%+v,callUrl:%v,resp:%v", c.UserInfo, types.PokerDecisionUri, index)
		if -1 != index {
			for _, value := range handCards[index].PokerHands {
				cards = append(cards, value...)
			}
		}
	} else {
		handCardStr := make([]string, 0)
		//入参转换
		code := utils.ObjectAToObjectB(req.Data, &handCardStr)
		if errorcode.Successfully != code {
			return code, cards
		}
		cards, _ = ListStringToCards(handCardStr, table.GameManage.LevelCardPoint)
	}
	return errorcode.Successfully, cards
}

// GetCardI2 @description: 进贡还贡下发牌-Ai
// @parameter table
func GetCardI2(table *types.Table) {
	var ticker, sum = time.NewTicker(1 * time.Second), 0
	for range ticker.C {
		sum++
		//进贡方和还贡方都操作完
		if len(table.GameManage.TributeGetCardInfos) == len(table.GameManage.NextTributeUsers) && len(table.GameManage.BackGetCardInfos) == len(table.GameManage.NextBackUsers) {
			//出第一手牌玩家名称，默认只有一个玩家
			firstOutCard := table.GameManage.NextTributeUsers[0].UserName
			if 1 == len(table.GameManage.TributeGetCardInfos) { //只有一个玩家进贡
				//进贡方和还贡方牌处理
				TributeBackDeal(table, table.GameManage.NextTributeUsers[0].UserName, table.GameManage.NextBackUsers[0].UserName, table.GameManage.TributeGetCardInfos[0].Card, table.GameManage.BackGetCardInfos[0].Card)
				//for _, v := range table.TableClients {
				//	//通知四个玩家进贡还贡信息
				//	v.WsSendMsg(types.BackGetCardIRespId, errorcode.Successfully, fmt.Sprintf("%s玩家进贡了一张%s,%s玩家还贡了一张%s", table.GameManage.NextTributeUsers[0].UserName, table.GameManage.TributeGetCardInfos[0].Card.Name, table.GameManage.NextBackUsers[0].UserName, table.GameManage.BackGetCardInfos[0].Card.Name))
				//}
			}
			if 2 == len(table.GameManage.TributeGetCardInfos) { //有两个玩家进贡
				//从table.GameManage.TributeGetCardInfos中找出进贡最大的牌
				indexM := 0
				if table.GameManage.MapViewToLevelH[table.GameManage.TributeGetCardInfos[0].Card.ViewNumber] < table.GameManage.MapViewToLevelH[table.GameManage.TributeGetCardInfos[1].Card.ViewNumber] {
					indexM = 1
				}
				//从table.GameManage.BackGetCardInfos中找出头游
				indexF := 0
				if table.GameManage.BackGetCardInfos[0].UserInfo.UserName != table.GameManage.NextBackUsers[0].UserName {
					indexF = 1
				}
				//进贡最大的牌玩家和头游方牌处理
				TributeBackDeal(table, table.GameManage.TributeGetCardInfos[indexM].UserInfo.UserName, table.GameManage.NextBackUsers[0].UserName, table.GameManage.TributeGetCardInfos[indexM].Card, table.GameManage.BackGetCardInfos[indexF].Card)
				//进贡最小的牌玩家和二游方牌处理
				TributeBackDeal(table, table.GameManage.TributeGetCardInfos[1-indexM].UserInfo.UserName, table.GameManage.NextBackUsers[1].UserName, table.GameManage.TributeGetCardInfos[1-indexM].Card, table.GameManage.BackGetCardInfos[1-indexF].Card)
				firstOutCard = table.GameManage.TributeGetCardInfos[indexM].UserInfo.UserName
				//for _, v := range table.TableClients {
				//	//通知四个玩家进贡还贡信息
				//	v.WsSendMsg(types.BackGetCardIRespId, errorcode.Successfully, fmt.Sprintf("%s,%s玩家分别进贡了一张%s,%s;%s,%s玩家分别还贡了一张%s,%s", table.GameManage.NextTributeUsers[0].UserName, table.GameManage.NextTributeUsers[1].UserName, table.GameManage.TributeGetCardInfos[0].Card.Name, table.GameManage.TributeGetCardInfos[1].Card.Name, table.GameManage.NextBackUsers[0].UserName, table.GameManage.NextBackUsers[1].UserName, table.GameManage.BackGetCardInfos[0].Card.Name, table.GameManage.BackGetCardInfos[1].Card.Name))
				//}
			}
			table.GameManage.NextOutCard = firstOutCard
			tempUserInfo := table.TableClients[firstOutCard].UserInfo
			for k, v := range table.TableClients {
				if k != firstOutCard {
					//通知其他玩家谁出第一手牌
					v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, tempUserInfo)
				}
			}
			//初始化上一局进贡还贡玩家的记录
			table.GameManage.NextBackUsers = make([]types.UserInfo, 0)
			table.GameManage.NextTributeUsers = make([]types.UserInfo, 0)
			//初始化上一局进贡还贡牌信息
			table.GameManage.TributeGetCardInfos = make([]types.GetCardInfo, 0)
			table.GameManage.BackGetCardInfos = make([]types.GetCardInfo, 0)
			//延时5秒通知谁出牌
			time.Sleep(5 * time.Second)
			//通知进贡给头游的玩家出第一手牌
			table.TableClients[firstOutCard].WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, 1)
			return
		}
		if 600 == sum { //等待超过10分钟后，超时退出
			return
		}
	}
}

// GetTributePokerHandType @description: 获取可以进贡的牌种类
// @parameter table
// @parameter c
// @return []types.HandCard
func GetTributePokerHandType(table types.Table, c types.Client) []types.HandCard {
	result := make([]types.HandCard, 0)
	var maxViewNumber string
	//获取通配牌名称
	strLName := types.PokerColorHearts + table.GameManage.LevelCardPoint
	//找最大的牌
	for _, card := range c.PlayerCardsInfo.CardList {
		if table.GameManage.MapViewToLevelH[card.ViewNumber] > table.GameManage.MapViewToLevelH[maxViewNumber] {
			maxViewNumber = card.ViewNumber
		}
	}
	for _, card := range c.PlayerCardsInfo.CardList {
		if card.Name != strLName && card.ViewNumber == maxViewNumber {
			handCard := new(types.HandCard)
			handCard.A = 1
			handCard.B = 1
			handCard.Type = types.PokerHandSingle
			handCard.LevelStr = card.ViewNumber
			handCard.Name = card.Name
			handCard.PokerHands = []types.Cards{[]types.Card{card}}
			result = append(result, *handCard)
		}
	}
	if 0 == len(result) {
		max := len(c.PlayerCardsInfo.CardList) - 1
		if c.PlayerCardsInfo.CardList[len(c.PlayerCardsInfo.CardList)-1].Name == strLName {
			max = len(c.PlayerCardsInfo.CardList) - 2
		}
		for i := max; i > 15; i-- {
			if c.PlayerCardsInfo.CardList[i].ViewNumber == c.PlayerCardsInfo.CardList[max].ViewNumber {
				handCard := new(types.HandCard)
				handCard.A = 1
				handCard.B = 1
				handCard.Type = types.PokerHandSingle
				handCard.LevelStr = c.PlayerCardsInfo.CardList[i].ViewNumber
				handCard.Name = c.PlayerCardsInfo.CardList[i].Name
				handCard.PokerHands = []types.Cards{[]types.Card{c.PlayerCardsInfo.CardList[i]}}
				result = append(result, *handCard)
			}
		}
	}
	return result
}

// GetBackPokerHandType @description: 获取可以还贡的牌种类
// @parameter table
// @parameter c
// @return []types.HandCard
func GetBackPokerHandType(table types.Table, c types.Client) []types.HandCard {
	result := make([]types.HandCard, 0)
	//获取通配牌名称
	strLName := types.PokerColorHearts + table.GameManage.LevelCardPoint
	for _, card := range c.PlayerCardsInfo.CardList {
		if card.Name != strLName && types.PokerLevel10 > table.GameManage.MapViewToLevelH[card.ViewNumber] {
			handCard := new(types.HandCard)
			handCard.A = 1
			handCard.B = 1
			handCard.Type = types.PokerHandSingle
			handCard.LevelStr = card.ViewNumber
			handCard.Name = card.Name
			handCard.PokerHands = []types.Cards{[]types.Card{card}}
			result = append(result, *handCard)
		}
	}
	return result
}

// TributeBackState @description: 重新进入后去获取进贡还贡的状态
// @parameter table
// @parameter c
// @return int
func TributeBackState(table types.Table, c types.Client) int {
	var result int
	if 0 == len(table.GameManage.NextTributeUsers) || 0 == len(table.GameManage.NextBackUsers) {
		return result
	}
	var tempTB int
	for _, tributeUser := range table.GameManage.NextTributeUsers {
		if tributeUser.UserName == c.UserInfo.UserName {
			tempTB = 1
			break
		}
	}
	for _, backUser := range table.GameManage.NextBackUsers {
		if backUser.UserName == c.UserInfo.UserName {
			tempTB = 2
			break
		}
	}
	if 1 == tempTB {
		for _, tributeInfo := range table.GameManage.TributeGetCardInfos {
			if tributeInfo.UserInfo.UserName == c.UserInfo.UserName { //已进贡完
				return result
			}
		}
		result = 1
	}
	if 2 == tempTB {
		for _, backInfo := range table.GameManage.BackGetCardInfos {
			if backInfo.UserInfo.UserName == c.UserInfo.UserName { //已还贡完
				return result
			}
		}
		result = 2
	}
	return result
}

// IsPictureUpload @description: 重进是否要拍照上传手牌
// @parameter table
// @parameter c
// @return bool
func IsPictureUpload(table types.Table, c types.Client) {
	if 4 != len(table.TableClients) {
		return
	}
	for _, v := range table.TableClients {
		if !v.Ready {
			return
		}
	}
	if (types.Ai1 == c.UserInfo.UserName || types.Ai2 == c.UserInfo.UserName) && 27 == len(c.PlayerCardsInfo.CardList) && 0 == c.PlayerCardsInfo.CardList[0].Id {
		c.WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
	}
}
