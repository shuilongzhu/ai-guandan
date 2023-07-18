package types

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
	"time"
)

// InitRoomManager @description: 初始化房间管理器（初始化一个房间10桌）
// @return *RoomManager
func InitRoomManager() *RoomManager {
	roomManager := new(RoomManager)
	roomManager.Rooms = make(map[RoomId]*Room, 0)
	for _, roomId := range RoomIds {
		room := new(Room)
		room.RoomId = int(roomId)
		room.Tables = make(map[TableId]*Table, 0)
		for _, tableId := range TableIds {
			table := new(Table)
			table.TableId = int(tableId)
			table.GameManage.LevelCardPoint = "2"  //最开始默认打2
			table.GameManage.LevelCardPointP = "2" //P方最开始打2
			table.GameManage.LevelCardPointA = "2" //A方最开始打2
			table.TableClients = make(map[string]*Client, 0)
			InitTMapViewToLevel(&table.GameManage.MapViewToLevelH)
			room.Tables[tableId] = table
		}
		roomManager.Rooms[roomId] = room
	}
	return roomManager
}

// InitOneTableStart @description: 初始化牌桌(一局开始之前)
// @receiver table
func (table *Table) InitOneTableStart() {
	//初始化牌桌规则
	table.GameManage.InitGameManage()
	cardsList := SendCards()
	lCardNumber := table.GameManage.LevelCardPoint
	index := 0
	for _, v := range table.TableClients {
		//初始化一桌每个玩家客户端
		v.InitClient(cardsList[index], lCardNumber)
		index++
	}
}

// InitOneTableStart2 @description: 初始化牌桌(一局开始之前)-Ai
// @receiver table
func (table *Table) InitOneTableStart2() {
	//初始化牌桌规则
	table.GameManage.InitGameManage()
	lCardNumber := table.GameManage.LevelCardPoint
	for _, v := range table.TableClients {
		//初始化一桌每个玩家客户端
		v.InitClient2(lCardNumber)
	}
}

// InitOneTableStart3 @description: 初始化牌桌(一局开始之前)-训练模型
// @receiver table
func (table *Table) InitOneTableStart3() {
	//初始化牌桌规则
	table.GameManage.InitGameManage()
	lCardNumber := table.GameManage.LevelCardPoint
	for _, v := range table.TableClients {
		//初始化一桌每个玩家客户端
		v.InitClient3(lCardNumber)
	}
}

// InitOneTableEnd @description: 初始化牌桌(一局结束之后)
// @receiver table
func (table *Table) InitOneTableEnd() {
	CardIdAi = 0
	InitTMapViewToLevel(&table.GameManage.MapViewToLevelH)
	table.GameManage.MapPlayOrder = make([]string, 0)
	table.GameManage.FirstOutCard = ""
	table.GameManage.MaxHandCard = *new(HandCard)
	table.GameManage.MaxHandCardPlayer = ""
	table.GameManage.AllIsEnterRoom = false
	table.GameManage.OutAllCardUsers = make([]UserInfo, 0)
	//table.GameManage.NextOutCard = ""
	for k, client := range table.TableClients {
		table.TableClients[k].PlayerCardsInfo = *new(PlayerCardsInfo)
		if 1 == client.UserInfo.RoomId { //四个玩家.Ready置为false
			table.TableClients[k].Ready = false
		} else {
			if Ai1 == k || Ai2 == k { // 只把A1,A2.Ready置为false
				table.TableClients[k].Ready = false
			}
		}
	}
}

// InitClient @description: 初始化一桌每个玩家客户端
// @receiver c
// @parameter cards
// @parameter lCardNumber
func (c *Client) InitClient(cards Cards, lCardNumber string) {
	lCards := make([]Card, 0)
	sort.Sort(cards)
	for i := 0; i < len(cards); i++ {
		//获取通配牌
		if isLCard(cards[i], lCardNumber) {
			lCards = append(lCards, cards[i])
		}
	}
	c.PlayerCardsInfo.LCardNumber = lCardNumber
	c.PlayerCardsInfo.CardList = cards
	c.PlayerCardsInfo.LCard = lCards
	gameManage := RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].GameManage
	index := FindStrListIndexN(gameManage.MapPlayOrder, c.UserInfo.UserName)
	if -1 == index {
		log.Errorf("未找到%v的索引", c.UserInfo.UserName)
		return
	}
	c.Next = RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].TableClients[gameManage.MapPlayOrder[index+1]]
}

// InitClient2 @description: 初始化一桌每个玩家客户端-Ai
// @receiver c
// @parameter lCardNumber
func (c *Client) InitClient2(lCardNumber string) {
	c.PlayerCardsInfo.LCardNumber = lCardNumber
	c.PlayerCardsInfo.CardList = make([]Card, 27)
	gameManage := RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].GameManage
	index := FindStrListIndexN(gameManage.MapPlayOrder, c.UserInfo.UserName)
	if -1 == index {
		log.Warnf("未找到%v的索引", c.UserInfo.UserName)
		return
	}
	c.Next = RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].TableClients[gameManage.MapPlayOrder[index+1]]
}

// InitClient3 @description: 初始化一桌每个玩家客户端-训练模型
// @receiver c
// @parameter lCardNumber
func (c *Client) InitClient3(lCardNumber string) {
	c.PlayerCardsInfo.LCardNumber = lCardNumber
	gameManage := RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].GameManage
	index := FindStrListIndexN(gameManage.MapPlayOrder, c.UserInfo.UserName)
	if -1 == index {
		log.Warnf("未找到%v的索引", c.UserInfo.UserName)
		return
	}
	c.Next = RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId].TableClients[gameManage.MapPlayOrder[index+1]]
}

// InitGameManage @description: 初始化牌桌规则
func (gameManage *GameManage) InitGameManage() {
	//随机数种子
	rand.Seed(time.Now().UnixNano())
	//出牌顺序先写死
	gameManage.MapPlayOrder = []string{"P1", "A1", "P2", "A2", "P1"}
	if nil == gameManage.NextBackUsers || 0 == len(gameManage.NextBackUsers) {
		//第一局随机确定出第一手牌的人名字
		gameManage.FirstOutCard = gameManage.MapPlayOrder[rand.Intn(4)]
	} else {
		//不是第一局出第一手牌的人为上一局的头游
		gameManage.FirstOutCard = gameManage.NextBackUsers[0].UserName
	}
	gameManage.MapViewToLevelH[gameManage.LevelCardPoint] = 15
}

// Shuffle @description: 打乱两副牌
// @receiver data
func (data Cards) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(data) > 0 {
		n := len(data)
		randIndex := r.Intn(n)
		data[n-1], data[randIndex] = data[randIndex], data[n-1]
		data = data[:n-1]
	}
}

// SendCards @description: 把两副牌分为四组
// @return [][]types.Card
func SendCards() []Cards {
	cardsList := make([]Cards, 4)
	TwoPlayCards.Shuffle()
	for index, card := range TwoPlayCards {
		temp := index % 4
		cardsList[temp] = append(cardsList[temp], card)
	}
	return cardsList
}

// IsLCard @description: 判断是否是通配牌
// @parameter card
// @parameter lCardNumber
// @return bool
func isLCard(card Card, lCardNumber string) bool {
	return lCardNumber == card.ViewNumber && PokerColorHearts == card.Color
}

// FindStrListIndexN @description: 找到target的index(方便找之后的元素)
// @parameter data
// @parameter target
// @return int
func FindStrListIndexN(data []string, target string) int {
	for index, value := range data {
		if value == target {
			return index
		}
	}
	return -1
}

// FindStrListIndexF @description: 找到target的index(方便找之前的元素)
// @parameter data
// @parameter target
// @return int
func FindStrListIndexF(data []string, target string) int {
	result := -1
	for index, value := range data {
		if value == target {
			result = index
		}
	}
	return result
}

func InitTwoPlayCards() {
	CardId = 0
	//生成一副牌
	oneCards := []Card{
		NewCardWithId(PokerColorHearts, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorSpades, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorDiamonds, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorClubs, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorHearts, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorSpades, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorDiamonds, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorClubs, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorHearts, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorSpades, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorDiamonds, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorClubs, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorHearts, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorSpades, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorDiamonds, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorClubs, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorHearts, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorSpades, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorDiamonds, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorClubs, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorHearts, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorSpades, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorDiamonds, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorClubs, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorHearts, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorSpades, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorDiamonds, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorClubs, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorHearts, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorSpades, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorDiamonds, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorClubs, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorHearts, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorSpades, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorDiamonds, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorClubs, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorHearts, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorSpades, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorDiamonds, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorClubs, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorHearts, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorSpades, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorDiamonds, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorClubs, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorHearts, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorSpades, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorDiamonds, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorClubs, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorHearts, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorSpades, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorDiamonds, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorClubs, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorSpades, PokerViewNumB, PokerViewNumB),
		NewCardWithId(PokerColorHearts, PokerViewNumR, PokerViewNumR),
	}
	//生成另一副牌
	otherCards := []Card{
		NewCardWithId(PokerColorHearts, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorSpades, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorDiamonds, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorClubs, PokerViewNum2, PokerViewNum2),
		NewCardWithId(PokerColorHearts, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorSpades, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorDiamonds, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorClubs, PokerViewNum3, PokerViewNum3),
		NewCardWithId(PokerColorHearts, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorSpades, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorDiamonds, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorClubs, PokerViewNum4, PokerViewNum4),
		NewCardWithId(PokerColorHearts, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorSpades, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorDiamonds, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorClubs, PokerViewNum5, PokerViewNum5),
		NewCardWithId(PokerColorHearts, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorSpades, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorDiamonds, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorClubs, PokerViewNum6, PokerViewNum6),
		NewCardWithId(PokerColorHearts, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorSpades, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorDiamonds, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorClubs, PokerViewNum7, PokerViewNum7),
		NewCardWithId(PokerColorHearts, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorSpades, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorDiamonds, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorClubs, PokerViewNum8, PokerViewNum8),
		NewCardWithId(PokerColorHearts, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorSpades, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorDiamonds, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorClubs, PokerViewNum9, PokerViewNum9),
		NewCardWithId(PokerColorHearts, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorSpades, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorDiamonds, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorClubs, PokerViewNumT, PokerViewNumT),
		NewCardWithId(PokerColorHearts, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorSpades, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorDiamonds, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorClubs, PokerViewNumJ, PokerViewNumJ),
		NewCardWithId(PokerColorHearts, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorSpades, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorDiamonds, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorClubs, PokerViewNumQ, PokerViewNumQ),
		NewCardWithId(PokerColorHearts, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorSpades, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorDiamonds, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorClubs, PokerViewNumK, PokerViewNumK),
		NewCardWithId(PokerColorHearts, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorSpades, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorDiamonds, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorClubs, PokerViewNumA, PokerViewNumA),
		NewCardWithId(PokerColorSpades, PokerViewNumB, PokerViewNumB),
		NewCardWithId(PokerColorHearts, PokerViewNumR, PokerViewNumR),
	}
	//获取两副牌
	TwoPlayCards = make(Cards, 0)
	TwoPlayCards = append(TwoPlayCards, oneCards...)
	TwoPlayCards = append(TwoPlayCards, otherCards...)
}

func InitTMapViewToLevel(MapViewToLevel *map[string]int) {
	*MapViewToLevel = make(map[string]int, 0)
	(*MapViewToLevel)[PokerViewNum2] = PokerLevel2
	(*MapViewToLevel)[PokerViewNum3] = PokerLevel3
	(*MapViewToLevel)[PokerViewNum4] = PokerLevel4
	(*MapViewToLevel)[PokerViewNum5] = PokerLevel5
	(*MapViewToLevel)[PokerViewNum6] = PokerLevel6
	(*MapViewToLevel)[PokerViewNum7] = PokerLevel7
	(*MapViewToLevel)[PokerViewNum8] = PokerLevel8
	(*MapViewToLevel)[PokerViewNum9] = PokerLevel9
	(*MapViewToLevel)[PokerViewNumT] = PokerLevel10
	(*MapViewToLevel)[PokerViewNumJ] = PokerLevelJ
	(*MapViewToLevel)[PokerViewNumQ] = PokerLevelQ
	(*MapViewToLevel)[PokerViewNumK] = PokerLevelK
	(*MapViewToLevel)[PokerViewNumA] = PokerLevelA
	(*MapViewToLevel)[PokerViewNumB] = PokerLevelB
	(*MapViewToLevel)[PokerViewNumR] = PokerLevelR
}

func InitMapViewToLevel() {
	MapViewToLevel = make(map[string]int, 0)
	MapViewToLevel[PokerViewNum2] = PokerLevel2
	MapViewToLevel[PokerViewNum3] = PokerLevel3
	MapViewToLevel[PokerViewNum4] = PokerLevel4
	MapViewToLevel[PokerViewNum5] = PokerLevel5
	MapViewToLevel[PokerViewNum6] = PokerLevel6
	MapViewToLevel[PokerViewNum7] = PokerLevel7
	MapViewToLevel[PokerViewNum8] = PokerLevel8
	MapViewToLevel[PokerViewNum9] = PokerLevel9
	MapViewToLevel[PokerViewNumT] = PokerLevel10
	MapViewToLevel[PokerViewNumJ] = PokerLevelJ
	MapViewToLevel[PokerViewNumQ] = PokerLevelQ
	MapViewToLevel[PokerViewNumK] = PokerLevelK
	MapViewToLevel[PokerViewNumA] = PokerLevelA
	MapViewToLevel[PokerViewNumB] = PokerLevelB
	MapViewToLevel[PokerViewNumR] = PokerLevelR
}

func InitMapLevelToView() {
	MapLevelToView = make(map[int]string, 0)
	MapLevelToView[PokerLevel2] = PokerViewNum2
	MapLevelToView[PokerLevel3] = PokerViewNum3
	MapLevelToView[PokerLevel4] = PokerViewNum4
	MapLevelToView[PokerLevel5] = PokerViewNum5
	MapLevelToView[PokerLevel6] = PokerViewNum6
	MapLevelToView[PokerLevel7] = PokerViewNum7
	MapLevelToView[PokerLevel8] = PokerViewNum8
	MapLevelToView[PokerLevel9] = PokerViewNum9
	MapLevelToView[PokerLevel10] = PokerViewNumT
	MapLevelToView[PokerLevelJ] = PokerViewNumJ
	MapLevelToView[PokerLevelQ] = PokerViewNumQ
	MapLevelToView[PokerLevelK] = PokerViewNumK
	MapLevelToView[PokerLevelA] = PokerViewNumA
	MapLevelToView[PokerLevelB] = PokerViewNumB
	MapLevelToView[PokerLevelR] = PokerViewNumR
}

func InitPokerHandLevelMap() {
	PokerHandLevelMap = make(map[string]int, 0)
	PokerHandLevelMap[PokerHandSingle] = 1
	PokerHandLevelMap[PokerHandPair] = 1
	PokerHandLevelMap[PokerHandTrips] = 1
	PokerHandLevelMap[PokerHandTwoTrips] = 1
	PokerHandLevelMap[PokerHandThreePair] = 1
	PokerHandLevelMap[PokerHandThreeWithTwo] = 1
	PokerHandLevelMap[PokerHandStraight] = 1
	PokerHandLevelMap[PokerHandBoomFour] = 2
	PokerHandLevelMap[PokerHandBoomFive] = 3
	PokerHandLevelMap[PokerHandStraightFlush] = 4
	PokerHandLevelMap[PokerHandBoomSix] = 5
	PokerHandLevelMap[PokerHandBoomSeven] = 6
	PokerHandLevelMap[PokerHandBoomEight] = 7
	PokerHandLevelMap[PokerHandBoomNine] = 8
	PokerHandLevelMap[PokerHandBoomTen] = 9
	PokerHandLevelMap[PokerHandJokerBoom] = 10
}

func init() {
	InitMapViewToLevel()
	InitMapLevelToView()
	InitPokerHandLevelMap()
	InitTwoPlayCards()
}
