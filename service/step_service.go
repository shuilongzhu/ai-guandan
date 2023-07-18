package service

import (
	"ai-guandan/errorcode"
	"ai-guandan/types"
	"ai-guandan/utils"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func TestClosedMethod(req types.AiWeWsReq, c *types.Client) {
	c.WsSendMsg(types.TestClosedRespId, errorcode.ErrAWESeWSClosedSuccessful, req.Data)
	c.Conn.Close()
	log.Println("websocket客户端关闭成功")
}

func TestConnectMethod(req types.AiWeWsReq, c *types.Client) {
	c.WsSendMsg(types.TestConnectRespId, errorcode.ErrAWESeWSConnectionSuccessful, req.Data)
	log.Println("websocket客户端连接成功")
}

// JoinRoomMethod @description: 加入房间逻辑处理
// @parameter req
// @parameter c
func JoinRoomMethod(req types.AiWeWsReq, c *types.Client) {
	userInfo := new(types.UserInfo)
	utils.MapToStruct(req.Data.(map[string]interface{}), userInfo)
	//入参校验
	if state := CheckUserInfoParameterJ(*userInfo, c); !state {
		return
	}
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId]
	joinRoomRespS := make([]types.JoinRoomResp, 0)
	for _, userName := range types.UserNames {
		joinRoomResp := new(types.JoinRoomResp)
		joinRoomResp.UserName = userName
		_, joinRoomResp.IsEntered = table.TableClients[userName]
		if joinRoomResp.IsEntered {
			joinRoomResp.IsEntered = table.TableClients[userName].Ready
		}
		joinRoomRespS = append(joinRoomRespS, *joinRoomResp)
	}
	c.WsSendMsg(types.JoinRoomRespId, errorcode.Successfully, joinRoomRespS)
}

// ConfirmCancelMethod @description: 取消确认，在选角色时逻辑处理
// @parameter req
// @parameter c
func ConfirmCancelMethod(req types.AiWeWsReq, c *types.Client) {
	userInfo := new(types.UserInfo)
	utils.MapToStruct(req.Data.(map[string]interface{}), userInfo)
	//入参校验
	if code := CheckUserInfoParameter(*userInfo); errorcode.Successfully != code {
		c.WsSendMsg(types.ConfirmCancelRespId, code, nil)
		return
	}
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId]
	for _, v := range table.TableClients {
		if v.Conn.RemoteAddr().String() == c.Conn.RemoteAddr().String() {
			delete(table.TableClients, userInfo.UserName)
		}
	}
	c.WsSendMsg(types.ConfirmCancelRespId, errorcode.Successfully, req.Data)
}

// StartGameMethod @description: 开始游戏步骤逻辑处理(Ai)
// @parameter req
// @parameter c
func StartGameMethod(req types.AiWeWsReq, c *types.Client) {
	userInfo := new(types.UserInfo)
	utils.MapToStruct(req.Data.(map[string]interface{}), userInfo)
	//入参校验
	if code := CheckUserInfoParameter(*userInfo); errorcode.Successfully != code {
		c.WsSendMsg(types.StartGameRespId, code, nil)
		return
	}
	//加锁
	types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId].Lock.RLock()
	defer types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId].Lock.RUnlock()
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId]
	if _, ok := table.TableClients[userInfo.UserName]; ok {
		if table.TableClients[userInfo.UserName].Ready {
			log.Warnf("%v已进入", userInfo.UserName)
			//返回开始游戏失败信息
			c.WsSendMsg(types.StartGameRespId, errorcode.ErrAWESeEnterRoom, *userInfo)
		} else {
			//再次进入
			StartGameAgainMethod(table, c, *userInfo, types.StartGameAgainRespId)
		}
		return
	}
	//限制一个websocket只能服务一个玩家
	for _, v := range table.TableClients {
		if v.Conn.RemoteAddr().String() == c.Conn.RemoteAddr().String() {
			c.WsSendMsg(types.StartGameRespId, errorcode.ErrAWESeWebSocktReuse, *userInfo)
			return
		}
	}
	if 1 == userInfo.RoomId { //四人线上
		//初次进入
		FirstEntryMethod(table, c, *userInfo)
	}
	if 2 == userInfo.RoomId { //人机对战
		//初次进入
		FirstEntryMethod2(table, c, *userInfo)
	}
}

// OutCardMethod @description: 出牌步骤逻辑处理
// @parameter req
// @parameter c
func OutCardMethod(req types.AiWeWsReq, c *types.Client) {
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	if table.GameManage.AllIsEnterRoom != true {
		log.Errorln("此牌桌的所有玩家还未都进入并准备好")
		c.WsSendMsg(types.OutCardRespId, errorcode.ErrAWESeAllIsEnterRoom, nil)
		return
	}
	cards := new(types.Cards)
	var code = errorcode.Successfully
	//获取玩家出的牌
	if 1 == c.UserInfo.RoomId { //四人线上
		code = utils.ObjectAToObjectB(req.Data, cards)
	} else { //人机对战，模型训练
		//获取玩家出的牌
		code = GetPlayerOutCards(*table, *c, cards, req)
	}
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
			if 1 == c.UserInfo.RoomId {
				c.WsSendMsg(types.OutCardRespId, code, c.UserInfo)
			} else {
				for k, v := range table.TableClients {
					if k == c.UserInfo.UserName {
						//识别有可能错误，通知它继续出牌
						v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, 1)
					} else {
						//通知其他玩家此玩家出的牌不符合规则
						v.WsSendMsg(types.OutCardRespId, code, c.UserInfo)
					}
				}
			}
			return
		}
		dealCardStr = "出牌成功"
	}
	//出牌或不出牌处理消息通知
	OutCardNotice(table, c, *cards, dealCardStr)
}

// DecidedCardTypeMethod @description: 出的牌存在多种牌型，确认哪种牌型
// @parameter req
// @parameter c
func DecidedCardTypeMethod(req types.AiWeWsReq, c *types.Client) {
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	handCard := new(types.HandCard)
	//入参转换
	code := utils.ObjectAToObjectB(req.Data, handCard)
	if errorcode.Successfully != code {
		c.WsSendMsg(types.OutCardRespId, code, nil)
		return
	}
	if 1 < len(handCard.PokerHands) {
		log.Errorln("入参手牌为空")
		return
	}
	cards := handCard.PokerHands[0]
	for i := 1; i < len(handCard.PokerHands); i++ {
		cards = append(cards, handCard.PokerHands[i]...)
	}
	//出牌处理消息通知
	OutCardNotice(table, c, cards, "出牌成功")
}

// OutCardNotice @description: 出牌处理消息通知
// @parameter table
// @parameter c
// @parameter cards
// @parameter dealCardStr
func OutCardNotice(table *types.Table, c *types.Client, cards types.Cards, dealCardStr string) {
	//通知出牌成功
	c.WsSendMsg(types.OutCardRespId, errorcode.Successfully, c.UserInfo.UserName+dealCardStr)
	//日志输出玩家出的手牌
	OutputHandRecord(c.UserInfo.TableId, c.UserInfo.UserName, cards)
	//把玩家出的牌，发给每个玩家，在app端显示出的牌
	for _, v := range table.TableClients {
		sendOutCardResp := new(types.SendOutCardResp)
		sendOutCardResp.UserInfo = c.UserInfo
		sendOutCardResp.Cards = cards
		sendOutCardResp.RemainCardsNum = len(c.PlayerCardsInfo.CardList)
		v.WsSendMsg(types.SendOutCardRespId, errorcode.Successfully, *sendOutCardResp)
	}
	//是否结束此局
	isEndState := false
	//此玩家牌出完
	if 0 == len(c.PlayerCardsInfo.CardList) {
		isEndState = OutAllCardDeal(table, c)
	}
	if isEndState {
		OneGameIsEnd(table)
		return
	}
	//一局还未结束处理
	GameUnFinished(table, c)
}

// AgainOneGameMethod @description: 再来一局逻辑处理
// @parameter req
// @parameter c
func AgainOneGameMethod(c *types.Client) {
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
		if !tempState {
			return
		}
		//如果都准备好就开始发牌
		if 1 == c.UserInfo.RoomId {
			//初始化牌桌
			table.InitOneTableStart()
			//判断此局是否出现抗贡
			isResist := IsResist(*table)
			table.GameManage.AllIsEnterRoom = true
			for k, v := range table.TableClients {
				sendCardsResp := new(types.SendCardsResp)
				if k == table.GameManage.FirstOutCard {
					sendCardsResp.IsOutCard = 1
				}
				sendCardsResp.LCardNumber = v.PlayerCardsInfo.LCardNumber
				sendCardsResp.LCardNumberP = table.GameManage.LevelCardPointP
				sendCardsResp.LCardNumberA = table.GameManage.LevelCardPointA
				sendCardsResp.Cards = v.PlayerCardsInfo.CardList
				sendCardsResp.IsResist = isResist
				sendCardsResp.AllIsEnterRoom = table.GameManage.AllIsEnterRoom
				v.WsSendMsg(types.SendCardsRespId, errorcode.Successfully, *sendCardsResp)
			}
			//是否需要进贡还贡
			IsTributeBack(table, isResist)
		} else {
			//初始化牌桌
			table.InitOneTableStart2()
			table.GameManage.AllIsEnterRoom = true
			table.TableClients[types.Ai1].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
			table.TableClients[types.Ai2].WsSendMsg(types.TakePictureUploadCardsRespId, errorcode.Successfully, nil)
		}
	}
}

// GetRemainTableIdMethod @description:获取剩余未满的桌号逻辑处理
// @parameter c
func GetRemainTableIdMethod(c *types.Client) {
	tableIds := make([]int, 0)
	for _, v := range types.RoomManagerObject.Rooms[1].Tables {
		lenT := len(v.TableClients)
		if 4 > lenT {
			tableIds = append(tableIds, v.TableId)
		}
		if 4 == lenT {
			tempState := true
			//判断四个人是否都准备好
			for _, vv := range v.TableClients {
				tempState = tempState && vv.Ready
			}
			if !tempState {
				tableIds = append(tableIds, v.TableId)
			}
		}
	}
	c.WsSendMsg(types.GetRemainTableIdRespId, errorcode.Successfully, tableIds)
}

// ExitRoomMethod @description: 退出房间逻辑处理
// @parameter c
func ExitRoomMethod(c *types.Client) {
	//客户端用户信息校验
	if state := CheckUserInfoPC(*c); !state {
		return
	}
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	c.Ready = false
	table.GameManage.AllIsEnterRoom = false
	c.WsSendMsg(types.ExitRoomRespId, errorcode.Successfully, "退出成功")
	for _, v := range table.TableClients {
		if c.UserInfo.UserName != v.UserInfo.UserName {
			v.WsSendMsg(types.ExitRoomRespId, errorcode.Successfully, c.UserInfo.UserName+"玩家退出此局比赛")
		}
	}
}

// TributeGetCardMethod @description: 进贡逻辑处理
// @parameter req
// @parameter c
func TributeGetCardMethod(req types.AiWeWsReq, c *types.Client) {
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	getCardInfo := new(types.GetCardInfo)
	var code int
	//入参获取
	if 1 == c.UserInfo.RoomId {
		code = utils.ObjectAToObjectB(req.Data, getCardInfo)
	} else {
		code = GetTributeCard(*table, *c, getCardInfo, req)
	}
	if errorcode.Successfully != code {
		c.WsSendMsg(types.TributeGetCardNRespId, code, nil)
		c.WsSendMsg(types.TributeGetCardRespId, errorcode.Successfully, "此局你要进贡啦") //进贡错误，重新下发进贡消息
		return
	}
	//入参校验
	if state := IsLegalGetCardInfo(*getCardInfo); !state {
		log.Errorln("进贡请求GetCardInfo入参不合法")
		c.WsSendMsg(types.TributeGetCardNRespId, errorcode.ErrAWESeGetCardInfoIllegal, nil)
		return
	}
	//获取通配牌名称
	strLName := types.PokerColorHearts + table.GameManage.LevelCardPoint
	if 27 != len(c.PlayerCardsInfo.CardList) {
		log.Errorf("%v 玩家手牌初始化错误", c.UserInfo.UserName)
		c.WsSendMsg(types.TributeGetCardNRespId, errorcode.ErrAWESeHandCardsLengthInit, nil)
		return
	}
	//进贡的牌是否符合要求
	if state := IsTributeGetCard(*c, strLName, getCardInfo.Card, table.GameManage.MapViewToLevelH); !state {
		c.WsSendMsg(types.TributeGetCardNRespId, errorcode.ErrAWESeTributeGetCard, nil)
		return
	}
	//只允许一次进入
	if 0 == len(table.GameManage.TributeGetCardInfos) && 0 != len(table.GameManage.NextTributeUsers) {
		//进贡还贡下发牌
		go GetCardI2(table)
	}
	table.GameManage.TributeGetCardInfos = append(table.GameManage.TributeGetCardInfos, *getCardInfo)
}

// BackGetCardMethod @description: 还贡逻辑处理
// @parameter req
// @parameter c
func BackGetCardMethod(req types.AiWeWsReq, c *types.Client) {
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	getCardInfo := new(types.GetCardInfo)
	var code int
	//入参获取
	if 1 == c.UserInfo.RoomId {
		code = utils.ObjectAToObjectB(req.Data, getCardInfo)
	} else {
		code = GetBackCard(*table, *c, getCardInfo, req)
	}
	if errorcode.Successfully != code {
		c.WsSendMsg(types.BackGetCardNRespId, code, nil)
		c.WsSendMsg(types.BackGetCardRespId, errorcode.Successfully, "此局你要还贡啦") //还贡错误，重新下发进贡消息
		return
	}
	//入参校验
	if state := IsLegalGetCardInfo(*getCardInfo); !state {
		log.Errorln("还贡请求GetCardInfo入参不合法")
		c.WsSendMsg(types.BackGetCardNRespId, errorcode.ErrAWESeGetCardInfoIllegal, nil)
		return
	}
	//获取通配牌名称
	strLName := types.PokerColorHearts + table.GameManage.LevelCardPoint
	if 27 != len(c.PlayerCardsInfo.CardList) {
		log.Errorf("%v 玩家手牌初始化错误", c.UserInfo.UserName)
		c.WsSendMsg(types.BackGetCardNRespId, errorcode.ErrAWESeHandCardsLengthInit, nil)
		return
	}
	//还贡的牌是否符合要求
	if strLName == getCardInfo.Card.Name || types.PokerLevel10 <= table.GameManage.MapViewToLevelH[getCardInfo.Card.ViewNumber] {
		c.WsSendMsg(types.BackGetCardNRespId, errorcode.ErrAWESeBackGetCard, nil)
		return
	}
	table.GameManage.BackGetCardInfos = append(table.GameManage.BackGetCardInfos, *getCardInfo)
}

// ClientReconnectionMethod @description: 客户端断开重新连接逻辑处理
// @parameter req
// @parameter c
func ClientReconnectionMethod(req types.AiWeWsReq, c *types.Client) {
	userInfo := new(types.UserInfo)
	utils.MapToStruct(req.Data.(map[string]interface{}), userInfo)
	//入参校验
	if code := CheckUserInfoParameter(*userInfo); errorcode.Successfully != code {
		c.WsSendMsg(types.ClientReconnectionRespId, code, nil)
		return
	}
	//加锁
	types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId].Lock.RLock()
	defer types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId].Lock.RUnlock()
	//获得具体哪一桌
	table := types.RoomManagerObject.Rooms[userInfo.RoomId].Tables[userInfo.TableId]
	if _, ok := table.TableClients[userInfo.UserName]; ok {
		if table.TableClients[userInfo.UserName].Ready {
			log.Warnf("%v已进入", userInfo.UserName)
			//返回开重连失败信息
			c.WsSendMsg(types.ClientReconnectionRespId, errorcode.ErrAWESeEnterRoom, *userInfo)
		} else {
			//再次重连进入
			StartGameAgainMethod(table, c, *userInfo, types.ClientReconnectionRespId)
		}
		return
	}
	//不存在此玩家，消息通知
	c.WsSendMsg(types.ClientReconnectionRespId, errorcode.ErrAWESeNotExistUserInfo, nil)
}

// HeartbeatMethod @description: 心跳逻辑处理
// @parameter c
func HeartbeatMethod(c *types.Client) {
	state := 0 == c.LatestHeartbeat
	c.LatestHeartbeat = time.Now().Unix()
	//如果客户端有用户信息，说明他已经登陆进去
	if code := CheckUserInfoParameter(c.UserInfo); errorcode.Successfully != code {
		table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
		//如果这一桌的第一个人出牌不为空，说明这一局正在玩，说明他已经准备好
		if 0 != len(table.GameManage.FirstOutCard) {
			c.Ready = true
			tempState := true
			//判断四个人是否都准备好
			for _, v := range table.TableClients {
				if 0 != len(v.PlayerCardsInfo.CardList) {
					tempState = tempState && v.Ready
				}
			}
			//这一桌的四个玩家是否都进入准备好
			if tempState && 4 == len(table.TableClients) {
				table.GameManage.AllIsEnterRoom = true
			}
		}
	}
	if state {
		go HeartbeatTimeJudge(c)
	}
	//心跳成功消息通知
	c.WsSendMsg(types.HeartbeatRespId, errorcode.Successfully, "心跳发送成功")
}

// DeleteHandCard @description: 删除已经出掉的牌
// @parameter c
// @parameter handCard
func DeleteHandCard(c *types.Client, handCard types.HandCard) {
	cards := make([]types.Card, 0)
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
	c.PlayerCardsInfo.CardList = cards
}

// IsTwoOneSidePlayer @description: 出完牌的所有玩家中是一组的人有2个就结束这一局
// @parameter userInfos
// @return bool
func IsTwoOneSidePlayer(userInfos []types.UserInfo) bool {
	count1 := 0
	count2 := 0
	for _, userInfo := range userInfos {
		if 1 == userInfo.UserRole {
			count1++
			continue
		}
		if 2 == userInfo.UserRole {
			count2++
		}
	}
	if 2 == count1 || 2 == count2 {
		return true
	}
	return false
}

// OutAllCardDeal @description: 玩家牌出完逻辑处理
// @parameter table
// @parameter c
// @return bool(是否结束此局)
func OutAllCardDeal(table *types.Table, c *types.Client) bool {
	table.GameManage.OutAllCardUsers = append(table.GameManage.OutAllCardUsers, c.UserInfo)
	ocuCount := len(table.GameManage.OutAllCardUsers)
	if 1 < ocuCount {
		//结束一局的判断
		if state := IsTwoOneSidePlayer(table.GameManage.OutAllCardUsers); state {
			//结束一局时下一个出牌玩家的记录
			table.GameManage.OutAllCardUsers = append(table.GameManage.OutAllCardUsers, c.Next.UserInfo)
			return true
		}
	}
	//发给每个玩家，此玩家出完牌的排名
	for _, v := range table.TableClients {
		v.WsSendMsg(types.OutAllCardSortRespId, errorcode.Successfully, ocuCount)
	}
	return false
}

// ChangeOutCardOrder @description: 改变出牌顺序
// @parameter strList
// @return []string
func ChangeOutCardOrder(strList []string, str string) []string {
	result := make([]string, 0)
	i := -1
	for index, value := range strList {
		if value == str {
			i = index
			continue
		}
		result = append(result, value)
	}
	//当str为strList中的最后一个做特殊处理
	if i == len(strList)-1 {
		result = append(result, strList[1])
	}
	return result
}

// FindTeammate @description: 找队友
// @parameter user
// @return types.UserInfo
func FindTeammate(str string) string {
	if 2 > len(str) {
		log.Errorln("玩家名称长度错误")
		return ""
	}
	result := str[0:1]
	if "1" == str[1:] {
		log.Printf("找到%v玩家的队友%v", str, result+"2")
		return result + "2"
	}
	log.Printf("找到%v玩家的队友%v", str, result+"1")
	return result + "1"
}

// CardsNotNUllMethod @description: 当玩家出的牌不为空时的处理
// @parameter table
// @parameter c
// @parameter cards
// @return bool(出的牌是否符合规则)
func CardsNotNUllMethod(table *types.Table, c *types.Client, cards *types.Cards) int {
	//手牌分析检测
	handCards := PokerHandAnalysis(c.UserInfo.UserName, c.UserInfo.Location, *cards, table.GameManage.LevelCardPoint, table.GameManage.MapViewToLevelH)
	if 0 == len(handCards) {
		log.Warnln("此牌不符合出牌规则")
		return errorcode.ErrAWESeNotPlayRules
	}
	handCard := handCards[0]
	//判断最大牌记录是否为空
	condition := 0 == table.GameManage.MaxHandCard.A
	if 1 < len(handCards) {
		if condition {
			log.Warnln("此牌存在多种组合")
			c.WsSendMsg(types.CardTypeConfirmRespId, errorcode.ErrAWESeHaveSomeCardType, handCards)
			return errorcode.ErrAWESeHaveSomeCardType
		} else {
			//是否符合出牌规则
			state := false
			for _, value := range handCards {
				if value.Type == table.GameManage.MaxHandCard.Type {
					handCard = value
					state = true
					break
				}
			}
			if !state {
				log.Warnln("此牌不符合出牌规则")
				return errorcode.ErrAWESeNotPlayRules
			}
		}
	}
	HandCardsToCards(handCard, cards)
	//最大牌记录不为空时
	if !condition {
		//比较出的牌和记录最大的牌
		if state := HandCardCompare(handCard, table.GameManage.MaxHandCard, table.GameManage.MapViewToLevelH); !state {
			log.Warnf("%v出的手牌没有大过这一轮的最大牌，请重新出牌", c.UserInfo.UserName)
			return errorcode.ErrAWESeOutCardsLTMaxCards
		}
	}
	//记录最大牌
	table.GameManage.MaxHandCard = handCard
	//记录出最大牌人的用户名
	table.GameManage.MaxHandCardPlayer = c.UserInfo.UserName
	//删除已经出掉的牌
	if 2 == c.UserInfo.RoomId { //人机对战
		DeleteHandCard2(c, handCard)
	} else {
		DeleteHandCard(c, handCard)
	}
	log.Printf("%v出牌成功", c.UserInfo.UserName)
	return errorcode.Successfully
}

// GameUnFinished @description: 一局还未结束处理
// @parameter table
// @parameter c
func GameUnFinished(table *types.Table, c *types.Client) {
	//如果下一家出完牌，就改变出牌的顺序
	for 0 == len(c.Next.PlayerCardsInfo.CardList) {
		if c.Next.UserInfo.UserName == table.GameManage.MaxHandCardPlayer {
			log.Printf("%v玩家出完牌，并且他这一轮出牌最大", c.Next.UserInfo.UserName)
			//如果下一家是出一轮最大牌的人，他必须出牌
			state := 1
			//把最大牌记录给清除
			table.GameManage.MaxHandCard = *new(types.HandCard)
			//如果下一家是出一轮最大牌的人，就下发清除桌面的消息给这一桌的每个客户端
			for _, v := range table.TableClients {
				v.WsSendMsg(types.ClearDeskTopRespId, errorcode.Successfully, "一轮牌出完，请清理桌面")
			}
			//找下一家的队友
			userName := FindTeammate(c.Next.UserInfo.UserName)
			//改变出牌顺序
			table.GameManage.MapPlayOrder = ChangeOutCardOrder(table.GameManage.MapPlayOrder, c.Next.UserInfo.UserName)
			c.Next = c.Next.Next
			//记录下一家谁出牌
			table.GameManage.NextOutCard = userName
			for _, v := range table.TableClients {
				if v.UserInfo.UserName != userName {
					//通知其他玩家谁在出牌
					v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, table.TableClients[userName].UserInfo)
				} else {
					//通知下一家的队友出牌
					v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, state)
				}

			}
			return
		}
		//改变出牌顺序
		table.GameManage.MapPlayOrder = ChangeOutCardOrder(table.GameManage.MapPlayOrder, c.Next.UserInfo.UserName)
		c.Next = c.Next.Next
	}
	state := 0
	if c.Next.UserInfo.UserName == table.GameManage.MaxHandCardPlayer {
		//如果下一家是出一轮最大牌的人，他必须出牌
		state = 1
		//把最大牌记录给清除
		table.GameManage.MaxHandCard = *new(types.HandCard)
		//如果下一家是出一轮最大牌的人，就下发清除桌面的消息给这一桌的每个客户端
		for _, v := range table.TableClients {
			v.WsSendMsg(types.ClearDeskTopRespId, errorcode.Successfully, "一轮牌出完，请清理桌面")
		}
	}
	//记录下一家谁出牌
	table.GameManage.NextOutCard = c.Next.UserInfo.UserName
	for _, v := range table.TableClients {
		if v.UserInfo.UserName != c.Next.UserInfo.UserName {
			//通知其他玩家谁在出牌
			v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, c.Next.UserInfo)
		} else {
			//通知下一家出牌
			v.WsSendMsg(types.NoticeNextPlayerOutCardRespId, errorcode.Successfully, state)
		}
	}
}

// HandCardsToCards @description: types.HandCard类型转换为types.Cards
// @parameter handCards
// @return types.Cards
func HandCardsToCards(handCard types.HandCard, cards *types.Cards) {
	if 1 > len(handCard.PokerHands) {
		return
	}
	*cards = make([]types.Card, 0)
	for i := 0; i < len(handCard.PokerHands); i++ {
		*cards = append(*cards, handCard.PokerHands[i]...)
	}
}

// CheckUserInfoParameterJ @description: 校验入参UserInfo.TableId是否合规
// @parameter userInfo
// @parameter c
// @return bool
func CheckUserInfoParameterJ(userInfo types.UserInfo, c *types.Client) bool {
	if 0 == userInfo.TableId {
		log.Warnln("userInfo.TableId parameter is empty")
		c.WsSendMsg(types.JoinRoomRespId, errorcode.ErrAWESeUserInfoTAEmpty, nil)
		return false
	}
	if 0 > userInfo.TableId || int(userInfo.TableId) > len(types.TableIds) {
		log.Warnln("tableId parameter is illegal")
		c.WsSendMsg(types.JoinRoomRespId, errorcode.ErrAWESeTableIdIllegal, nil)
		return false
	}
	return true
}

// CheckUserInfoParameter @description: 校验入参UserInfo是否合规
// @parameter userInfo
// @parameter c
// @return bool
func CheckUserInfoParameter(userInfo types.UserInfo) int {
	if 0 == userInfo.UserId || "" == userInfo.UserName || 0 == userInfo.UserRole || 0 == userInfo.Location || 0 == userInfo.TableId || 0 == userInfo.RoomId {
		log.Warnln("userInfo parameter is empty")
		return errorcode.ErrAWESeUserInfoAEmpty
	}
	if 0 > userInfo.TableId || int(userInfo.TableId) > len(types.TableIds) {
		log.Warnln("tableId parameter is illegal")
		return errorcode.ErrAWESeTableIdIllegal
	}
	return errorcode.Successfully
}

// CheckUserInfoPC @description: 客户端用户信息校验
// @parameter c
// @return bool
func CheckUserInfoPC(c types.Client) bool {
	if 0 == c.UserInfo.UserId || "" == c.UserInfo.UserName || 0 == c.UserInfo.UserRole || 0 == c.UserInfo.Location || 0 == c.UserInfo.TableId || 0 == c.UserInfo.RoomId {
		log.Warnln("客户端userInfo信息为空")
		c.WsSendMsg(types.ExitRoomRespId, errorcode.ErrAWESeUserInfoAEmptyC, nil)
		return false
	}
	if 0 > c.UserInfo.TableId || int(c.UserInfo.TableId) > len(types.TableIds) {
		log.Warnln("客户端tableId为非法id")
		c.WsSendMsg(types.ExitRoomRespId, errorcode.ErrAWESeTableIdIllegalC, nil)
		return false
	}
	return true
}

// OutputHandRecord @description: 日志输出玩家出的手牌
// @parameter cards
func OutputHandRecord(tableId types.TableId, userName string, cards types.Cards) {
	if 0 == len(cards) {
		log.Errorf("tableId:%v桌玩家%v不出", tableId, userName)
	} else {
		hcStr := ""
		for _, card := range cards {
			hcStr = hcStr + card.Name
		}
		log.Errorf("tableId:%v桌玩家%v出了:%v", tableId, userName, hcStr)
	}
}

// StartGameAgainMethod @description: 再一次进入游戏逻辑处理及消息返回
// @parameter table
// @parameter c
// @parameter userInfo
func StartGameAgainMethod(table *types.Table, c *types.Client, userInfo types.UserInfo, idOwn int) {
	c.PlayerCardsInfo = table.TableClients[userInfo.UserName].PlayerCardsInfo
	c.UserInfo = userInfo
	c.Room = table.TableClients[userInfo.UserName].Room
	c.Ready = true
	c.IsCalled = table.TableClients[userInfo.UserName].IsCalled
	c.Next = table.TableClients[userInfo.UserName].Next
	table.TableClients[userInfo.UserName] = c
	tempState := true
	if 4 != len(table.TableClients) {
		tempState = false
	} else {
		//判断四个人是否都准备好
		for _, v := range table.TableClients {
			tempState = tempState && v.Ready
		}
	}
	//这一桌的所有玩家是否都进入准备好
	table.GameManage.AllIsEnterRoom = tempState
	sendCardsAgainResp := new(types.SendCardsAgainResp)
	sendCardsAgainResp.UserInfo = userInfo
	sendCardsAgainResp.Cards = c.PlayerCardsInfo.CardList
	sendCardsAgainResp.LCardNumber = c.PlayerCardsInfo.LCardNumber
	sendCardsAgainResp.LCardNumberP = table.GameManage.LevelCardPointP
	sendCardsAgainResp.LCardNumberA = table.GameManage.LevelCardPointA
	//是否是他出牌
	sendCardsAgainResp.IsOutCard = table.GameManage.NextOutCard == userInfo.UserName
	//是否必须出牌
	sendCardsAgainResp.IsMustOutCard = table.GameManage.MaxHandCardPlayer == userInfo.UserName
	sendCardsAgainResp.AllIsEnterRoom = table.GameManage.AllIsEnterRoom
	sendCardsAgainResp.IsOneMoreGame = IsOneMoreGame(*table, *c)
	//判断是否有玩家出过牌，如果有玩家出过牌table.GameManage.NextOutCard就不为空
	if 0 == len(table.GameManage.NextOutCard) {
		tempState := table.GameManage.FirstOutCard == userInfo.UserName
		//是否是他出牌
		sendCardsAgainResp.IsOutCard = tempState
		//是否必须出牌
		sendCardsAgainResp.IsMustOutCard = tempState
	}
	sendCardsAgainResp.TributeBackState = TributeBackState(*table, *c)
	//重进是否要拍照上传手牌
	IsPictureUpload(*table, *c)
	log.Printf("%v玩家重新进入游戏成功", userInfo.UserName)
	//返回给此玩家再次进入开始游戏所需的信息
	c.WsSendMsg(idOwn, errorcode.Successfully, *sendCardsAgainResp)
}

// FirstEntryMethod @description: 初次进入游戏逻辑处理及消息返回
// @parameter table
// @parameter c
// @parameter userInfo
func FirstEntryMethod(table *types.Table, c *types.Client, userInfo types.UserInfo) {
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
		//如果都准备好就开始发牌
		if tempState {
			//初始化牌桌
			//InitOneTableStart(table)
			table.InitOneTableStart()
			table.GameManage.AllIsEnterRoom = true
			tempUserInfo := table.TableClients[table.GameManage.FirstOutCard].UserInfo
			for k, v := range table.TableClients {
				sendCardsResp := new(types.SendCardsResp)
				if k == table.GameManage.FirstOutCard {
					sendCardsResp.IsOutCard = 1
				} else {
					//通知其他玩家谁在出牌
					v.WsSendMsg(types.NoticeWhoOutCardRespId, errorcode.Successfully, tempUserInfo)
				}
				sendCardsResp.LCardNumber = v.PlayerCardsInfo.LCardNumber
				sendCardsResp.LCardNumberP = table.GameManage.LevelCardPointP
				sendCardsResp.LCardNumberA = table.GameManage.LevelCardPointA
				sendCardsResp.Cards = v.PlayerCardsInfo.CardList
				sendCardsResp.IsResist = true
				sendCardsResp.AllIsEnterRoom = table.GameManage.AllIsEnterRoom
				v.WsSendMsg(types.SendCardsRespId, errorcode.Successfully, *sendCardsResp)
			}
		}
	}
}

// OneGameIsEnd @description: 一局结束逻辑处理
// @parameter table
func OneGameIsEnd(table *types.Table) {
	endOneGameResp := new(types.EndOneGameResp)
	code := utils.ObjectAToObjectB(table.GameManage.OutAllCardUsers, &endOneGameResp.RankingInfos)
	if errorcode.Successfully != code {
		log.Errorln("table.GameManage.OutAllCardUsers convert endOneGameResp err")
		return
	}
	//此局结算
	ScoreCalculate(table, endOneGameResp, table.GameManage.OutAllCardUsers)
	//通配牌升级
	LevelCardPointUpgrade(table, endOneGameResp.Score)
	//InitOneTableEnd(table)
	table.InitOneTableEnd()
	log.Printf("table_id:%d号桌此局结束", table.TableId)
	//发给每个玩家，此局结束
	for _, v := range table.TableClients {
		v.WsSendMsg(types.EndOneGameRespId, errorcode.Successfully, *endOneGameResp)
	}
}

// ScoreCalculate @description: 此局结算
// @parameter table
// @parameter endOneGameResp
func ScoreCalculate(table *types.Table, endOneGameResp *types.EndOneGameResp, outAllCardUsers []types.UserInfo) {
	tempLength := len(table.GameManage.OutAllCardUsers)
	if 3 == tempLength {
		//此时有一方为双下，最后两名的排名随机给
		endOneGameResp.Score = 3
		tempStr := endOneGameResp.RankingInfos[2].UserName[0:1]
		if endOneGameResp.RankingInfos[2].UserName[1:2] == "1" {
			tempStr = tempStr + "2"
		} else {
			tempStr = tempStr + "1"
		}
		tempRankingInfo := new(types.RankingInfo)
		utils.ObjectAToObjectB(table.TableClients[tempStr].UserInfo, tempRankingInfo)
		endOneGameResp.RankingInfos = append(endOneGameResp.RankingInfos, *tempRankingInfo)
		table.GameManage.NextBackUsers = append(table.GameManage.NextBackUsers, outAllCardUsers[0])
		table.GameManage.NextBackUsers = append(table.GameManage.NextBackUsers, outAllCardUsers[1])
		table.GameManage.NextTributeUsers = append(table.GameManage.NextTributeUsers, outAllCardUsers[2])
		table.GameManage.NextTributeUsers = append(table.GameManage.NextTributeUsers, table.TableClients[tempStr].UserInfo)
	}
	for i := 0; i < len(endOneGameResp.RankingInfos); i++ {
		endOneGameResp.RankingInfos[i].Ranking = i + 1
	}
	if 4 == tempLength {
		if table.GameManage.OutAllCardUsers[0].UserName[0:1] == table.GameManage.OutAllCardUsers[2].UserName[0:1] {
			endOneGameResp.Score = 2
			//把获胜的一方放在一边，输的放在一边
			endOneGameResp.RankingInfos[1], endOneGameResp.RankingInfos[2] = endOneGameResp.RankingInfos[2], endOneGameResp.RankingInfos[1]
		} else {
			endOneGameResp.Score = 1
			//把获胜的一方放在一边，输的放在一边
			endOneGameResp.RankingInfos[1], endOneGameResp.RankingInfos[3] = endOneGameResp.RankingInfos[3], endOneGameResp.RankingInfos[1]
		}
		table.GameManage.NextBackUsers = append(table.GameManage.NextBackUsers, outAllCardUsers[0])
		table.GameManage.NextTributeUsers = append(table.GameManage.NextTributeUsers, outAllCardUsers[3])
	}
}

// LevelCardPointUpgrade @description: 通配牌升级
// @parameter table
// @parameter score
func LevelCardPointUpgrade(table *types.Table, score int) {
	tempPoint := ""
	cPoint := strings.Contains(table.GameManage.OutAllCardUsers[0].UserName, "P")
	if cPoint { //P方升级
		tempPoint = table.GameManage.LevelCardPointP
	} else { //A方升级
		tempPoint = table.GameManage.LevelCardPointA
	}
	level := types.MapViewToLevel[tempPoint]
	if types.PokerViewNumA == tempPoint {
		table.GameManage.LevelCardPoint = types.PokerViewNum2
		table.GameManage.LevelCardPointP = types.PokerViewNum2
		table.GameManage.LevelCardPointA = types.PokerViewNum2
		return
	} else {
		level = level + score
		if types.PokerLevelA < level {
			level = types.PokerLevelA
		}
	}
	if cPoint { //P方升级
		table.GameManage.LevelCardPointP = types.MapLevelToView[level]
	} else { //A方升级
		table.GameManage.LevelCardPointA = types.MapLevelToView[level]
	}
	table.GameManage.LevelCardPoint = types.MapLevelToView[level]
}

// IsResist @description:此局是否出现抗贡
// @parameter table
// @return bool
func IsResist(table types.Table) bool {
	count := 0
	for _, user := range table.GameManage.NextTributeUsers {
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
	return 2 == count
}

// IsTributeGetCard @description: 进贡的牌是否符合要求
// @parameter c
// @parameter strLName
// @parameter card
// @return bool
func IsTributeGetCard(c types.Client, strLName string, card types.Card, mapViewToLevelH map[string]int) bool {
	//如果是通配牌
	if strLName == card.Name {
		return false
	}
	state := true
	for _, value := range c.PlayerCardsInfo.CardList {
		//判断进贡的牌是否是此玩家所有牌中最大的牌
		if mapViewToLevelH[value.ViewNumber] > mapViewToLevelH[card.ViewNumber] {
			state = false
		}
	}
	return state
}

// GetCardI @description: 进贡还贡下发牌
// @parameter table
func GetCardI(table *types.Table) {
	for {
		//进贡方和还贡方都操作完
		if len(table.GameManage.TributeGetCardInfos) == len(table.GameManage.NextTributeUsers) && len(table.GameManage.BackGetCardInfos) == len(table.GameManage.NextBackUsers) {
			//出第一手牌玩家名称，默认只有一个玩家
			firstOutCard := table.GameManage.NextTributeUsers[0].UserName
			if 1 == len(table.GameManage.TributeGetCardInfos) { //只有一个玩家进贡
				//进贡方和还贡方牌处理
				TributeBackDeal(table, table.GameManage.NextTributeUsers[0].UserName, table.GameManage.NextBackUsers[0].UserName, table.GameManage.TributeGetCardInfos[0].Card, table.GameManage.BackGetCardInfos[0].Card)
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
			}
			//通知进贡给头游的玩家出第一手牌
			table.TableClients[firstOutCard].WsSendMsg(types.TributeMaxOutCardRespId, errorcode.Successfully, "进贡完你出第一手牌")
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
			return
		}
	}
}

// ReplaceCard @description: 替换玩家手里的牌
// @parameter c
// @parameter old
// @parameter new
func ReplaceCard(c *types.Client, old, new types.Card) {
	for i := 0; i < len(c.PlayerCardsInfo.CardList); i++ {
		if old.Name == c.PlayerCardsInfo.CardList[i].Name {
			new.Id = old.Id
			c.PlayerCardsInfo.CardList[i] = new
			return
		}
	}
}

// IsLegalGetCardInfo @description: 入参types.GetCardInfo校验
// @parameter getCardInfo
// @return bool
func IsLegalGetCardInfo(getCardInfo types.GetCardInfo) bool {
	if 0 == len(getCardInfo.UserInfo.UserName) || 0 == getCardInfo.Card.Id || 0 == len(getCardInfo.Card.Name) {
		return false
	}
	return true
}

// TributeBackDeal @description: 进贡还贡牌处理
// @parameter table
// @parameter tributeName(进贡方的名称)
// @parameter backName(还贡方的名称)
// @parameter tributeCard(进贡方进贡的牌)
// @parameter backCard(还贡方还贡的牌)
func TributeBackDeal(table *types.Table, tributeName, backName string, tributeCard, backCard types.Card) {
	//返还给进贡方还贡方的牌
	table.TableClients[tributeName].WsSendMsg(types.TributeGetCardIRespId, errorcode.Successfully, backCard)
	//返还给还贡方进贡方的牌
	table.TableClients[backName].WsSendMsg(types.BackGetCardIRespId, errorcode.Successfully, tributeCard)
	//替换进贡玩家手里的牌
	ReplaceCard(table.TableClients[tributeName], tributeCard, backCard)
	//替换还贡玩家手里的牌
	ReplaceCard(table.TableClients[backName], backCard, tributeCard)
}

// HeartbeatTimeJudge @description: 心跳超时判断
// @parameter c
func HeartbeatTimeJudge(c *types.Client) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		if 0 == c.LatestHeartbeat {
			return
		}
		//心跳超过6秒
		if time.Now().Unix()-c.LatestHeartbeat > 6 {
			log.Errorf("userInfo:%+v,HeartbeatTimeJudge() websocket No heartbeat!", c.UserInfo)
			//客户端用户信息校验
			if state := CheckUserInfoPC(*c); !state {
				return
			}
			table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
			//判断客户端是否是这一桌玩家最新的客户端
			if table.TableClients[c.UserInfo.UserName] == c {
				log.Printf("userInfo:%+v is latest client,location:HeartbeatTimeJudge() ", c.UserInfo)
				go MoreFiveMinutesInit(c, c.LatestHeartbeat)
				c.Ready = false
				table.GameManage.AllIsEnterRoom = false
				c.LatestHeartbeat = 0
			}
			return
		}
	}
}

func MoreFiveMinutesInit(c *types.Client, heartbeat int64) {
	ticker := time.NewTicker(4 * time.Minute)
	for range ticker.C {
		if time.Now().Unix()-heartbeat > 300 {
			//客户端用户信息校验
			if state := CheckUserInfoPC(*c); !state {
				return
			}
			table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
			//如果这一桌的有一个玩家最近上一次的心跳距离现在超过了5分钟，就初始化这一桌
			if table.TableClients[c.UserInfo.UserName] == c && heartbeat >= c.LatestHeartbeat && 0 != len(table.TableClients) {
				log.Errorf("userInfo:%+v,tableId:%v table has inited", c.UserInfo, c.UserInfo.TableId)
				table.TableClients = make(map[string]*types.Client, 0)
				table.GameManage = *new(types.GameManage)
				types.InitTMapViewToLevel(&table.GameManage.MapViewToLevelH)
				table.GameManage.LevelCardPoint = "2"  //最开始默认打2
				table.GameManage.LevelCardPointP = "2" //P方最开始打2
				table.GameManage.LevelCardPointA = "2" //A方最开始打2
			}
			return
		}
	}
}

// IsOneMoreGame @description: 重进是否再来一局判断
// @parameter table
// @parameter c
// @return bool
func IsOneMoreGame(table types.Table, c types.Client) bool {
	if 4 != len(table.TableClients) {
		return false
	}
	for _, v := range table.TableClients {
		if 0 != len(v.PlayerCardsInfo.CardList) {
			return false
		}
	}
	return !c.Ready
}

// IsTributeBack @description: 是否需要进贡还贡
// @parameter table
// @parameter isResist
func IsTributeBack(table *types.Table, isResist bool) {
	if isResist || "2" == table.GameManage.LevelCardPoint {
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

			}
		}
		return
	}
	for _, user := range table.GameManage.NextTributeUsers {
		table.TableClients[user.UserName].WsSendMsg(types.TributeGetCardRespId, errorcode.Successfully, "此局你要进贡啦")
	}
	for _, user := range table.GameManage.NextBackUsers {
		table.TableClients[user.UserName].WsSendMsg(types.BackGetCardRespId, errorcode.Successfully, "此局你要还贡啦")
	}
}
