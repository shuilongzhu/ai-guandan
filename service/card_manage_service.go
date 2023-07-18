package service

import (
	"ai-guandan/types"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
	"time"
)

type Player struct {
	Location int
	Name     string
	Role     int   // 0 自己局 ， 1 对方局
	Score    []int // 0 1
	Level    string
	Cards    types.Cards
	LCard    []types.Card //通配牌

	// 选牌组合
	dfsCards      types.Cards
	dfsMasks      []int
	dfsResults    []types.Cards
	dfsTargetNum  int
	dfsCardDomain int
	dfsTargetCard types.Cards
	//

	pokerHand PokerHand
}

func DefaultPlayer(location int, playerName, lCardNumber string, cards, lCards types.Cards) *Player {
	sort.Sort(cards)
	return &Player{
		Location: location,
		Name:     playerName,
		Level:    lCardNumber,
		Cards:    cards,
		LCard:    lCards,
	}
}

type PokerHand struct {
	Name  string // 调试标识用
	Type  string // 牌型,自动计算出来或者应用赋值
	Level string // 当前牌的级别
	Num   int

	Cards types.Cards // 当前手牌集合

	minCard types.Card // 排序后最小的牌,  不计算配子
	maxCard types.Card // 排序后最大的牌, 不计算配子

	Trip types.Card // 统计3点数用
	Pair types.Card //  统计对子点数用过

	wildCard      types.Card //  配子
	wildCardNum   int        // 0-2
	wildCardIndex []int

	parseDone bool

	colorParsed   map[string]int
	viewNumParsed map[string]int

	parseResults []types.Cards // 分析完成后结果保存在这里, 输出所有可能的牌型
}

// PokerHandAnalysis @description:手牌分析检查
// @parameter cards
// @parameter lCardNumber(通配牌牌点)
// @return []types.HandCard
func PokerHandAnalysis(playerName string, location int, cards types.Cards, lCardNumber string, mapViewToLevelH map[string]int) []types.HandCard {
	lCards := make([]types.Card, 0)
	count := len(cards)
	if 1 > count {
		return nil
	}
	//获取手牌中的通配牌
	for _, lCard := range cards {
		if isLCard(lCard, lCardNumber) {
			lCards = append(lCards, lCard)
		}
	}
	//把一手牌当成一个玩家手中的所有牌来统一处理
	p := DefaultPlayer(location, playerName, lCardNumber, cards, lCards)
	//牌为大小王的判断
	condition1 := types.PokerViewNumB == cards[0].ViewNumber || types.PokerViewNumR == cards[0].ViewNumber
	condition2 := 1 == count || 2 == count || 3 == count || ((4 == count) && !condition1) || 7 == count || 8 == count || 9 == count || 10 == count
	//手牌张数为四时，四个王判断
	if 4 == count && condition1 {
		return p.analysisFourWang()
	}
	//手牌张数为五时，牌型判断
	if 5 == count {
		return p.analysisCountFive(mapViewToLevelH)
	}
	//手牌张数为六时，牌型判断
	if 6 == count {
		return p.analysisCountSix(mapViewToLevelH)
	}
	//a=n,b=1手牌的统一处理，a(每种牌点相同的张数)，b(连续不同牌点的张数)
	if condition2 {
		return p.analysisSingle(count)
	}
	return nil
}

// analysisSingle @description: a=n,b=1手牌的统一处理，a(每种牌点相同的张数)，b(连续不同牌点的张数)
// @receiver p
// @parameter count
// @return []types.HandCard
func (p *Player) analysisSingle(count int) []types.HandCard {
	channel1 := make(chan []types.HandCard)
	go p.ContinuousPokerHandGo(channel1, count, 1, false)
	channelObject1 := <-channel1
	return channelObject1
}

// analysisFourWang @description: 手牌张数为四时，四个王判断
// @receiver p
// @return []types.HandCard
func (p *Player) analysisFourWang() []types.HandCard {
	result := make([]types.HandCard, 0)
	var sum int
	for _, card := range p.Cards {
		sum += types.MapViewToLevel[card.ViewNumber]
	}
	if 68 == sum {
		handCard := new(types.HandCard)
		handCard.A = 22
		handCard.B = 11
		handCard.Type = types.PokerHandJokerBoom
		handCard.LevelStr = types.SmallWang
		handCard.Name = types.SmallWang + types.BigWang + types.SmallWang + types.BigWang
		temp := make([]types.Cards, 0)
		temp = append(temp, p.Cards)
		handCard.PokerHands = temp
		result = append(result, *handCard)
	}
	return result
}

// analysisFourWangGo @description: 手牌张数为四时，四个王判断（并发）
// @receiver p
// @parameter channel
func (p *Player) analysisFourWangGo(channel chan []types.HandCard) {
	result := make([]types.HandCard, 0)
	var bSum int
	var rSum int
	for _, card := range p.Cards {
		if types.PokerViewNumB == card.ViewNumber {
			bSum++
			continue
		}
		if types.PokerViewNumR == card.ViewNumber {
			rSum++
		}
	}
	if 2 == bSum && 2 == rSum {
		handCard := new(types.HandCard)
		handCard.A = 22
		handCard.B = 11
		handCard.Type = types.PokerHandJokerBoom
		handCard.LevelStr = types.SmallWang
		handCard.Name = types.SmallWang + types.BigWang + types.SmallWang + types.BigWang
		temp := make([]types.Cards, 0)
		temp = append(temp, p.Cards)
		handCard.PokerHands = temp
		result = append(result, *handCard)
	}
	channel <- result
}

// analysisCountFive @description: 手牌张数为五时，牌型判断
// @receiver p
// @return []types.HandCard
func (p *Player) analysisCountFive(mapViewToLevelH map[string]int) []types.HandCard {
	result := make([]types.HandCard, 0)
	//1.定义三个通道， 并发处理接收数据
	channel1 := make(chan []types.HandCard)
	channel2 := make(chan []types.HandCard)
	channel3 := make(chan []types.HandCard)
	channel4 := make(chan []types.HandCard)
	//同花顺的判断
	go p.ContinuousPokerHandGo(channel1, 1, 5, true)
	//五张相同牌炸的判断
	go p.ContinuousPokerHandGo(channel2, 5, 1, false)
	//顺子的判断
	go p.ContinuousPokerHandGo(channel3, 1, 5, false)
	//三带二的判断
	go p.GetThreeWithTwoHandCardsGo(channel4)
	handCard1 := <-channel1
	handCard2 := <-channel2
	handCard3 := <-channel3
	handCard4 := <-channel4
	if maxHandCard1 := SameCardTypeFindMax(handCard1, 5, mapViewToLevelH); nil != maxHandCard1 {
		result = append(result, *maxHandCard1)
		return result
	}
	if maxHandCard2 := SameCardTypeFindMax(handCard2, 1, mapViewToLevelH); nil != maxHandCard2 {
		result = append(result, *maxHandCard2)
		return result
	}
	if maxHandCard3 := SameCardTypeFindMax(handCard3, 5, mapViewToLevelH); nil != maxHandCard3 {
		result = append(result, *maxHandCard3)
	}
	if maxHandCard4 := SameCardTypeFindMax(handCard4, 11, mapViewToLevelH); nil != maxHandCard4 {
		result = append(result, *maxHandCard4)
	}
	return result
}

// analysisCountSix @description:手牌张数为六时，牌型判断
// @receiver p
// @return []types.HandCard
func (p *Player) analysisCountSix(mapViewToLevelH map[string]int) []types.HandCard {
	result := make([]types.HandCard, 0)
	//1.定义三个通道， 并发处理接收数据
	channel1 := make(chan []types.HandCard)
	channel2 := make(chan []types.HandCard)
	channel3 := make(chan []types.HandCard)
	//六张相同牌炸的判断
	go p.ContinuousPokerHandGo(channel1, 6, 1, false)
	//钢板(333444)的判断
	go p.ContinuousPokerHandGo(channel2, 3, 2, false)
	//三连对的判断
	go p.ContinuousPokerHandGo(channel3, 2, 3, false)
	handCard1 := <-channel1
	handCard2 := <-channel2
	handCard3 := <-channel3
	if maxHandCard1 := SameCardTypeFindMax(handCard1, 1, mapViewToLevelH); nil != maxHandCard1 {
		result = append(result, *maxHandCard1)
		return result
	}
	if maxHandCard2 := SameCardTypeFindMax(handCard2, 2, mapViewToLevelH); nil != maxHandCard2 {
		result = append(result, *maxHandCard2)
	}
	if maxHandCard3 := SameCardTypeFindMax(handCard3, 3, mapViewToLevelH); nil != maxHandCard3 {
		result = append(result, *maxHandCard3)
	}
	return result
}

// SameCardTypeFindMax @description: 相同的牌型中找最大的手牌
// @parameter handCards
// @parameter b(连续不同牌点的张数)
// @return *types.HandCard
func SameCardTypeFindMax(handCards []types.HandCard, b int, mapViewToLevelH map[string]int) *types.HandCard {
	if 1 > len(handCards) {
		return nil
	}
	result := handCards[0]
	for i := 1; i < len(handCards); i++ {
		//牌型不为连子level大小
		cc := mapViewToLevelH[result.LevelStr]
		dd := mapViewToLevelH[handCards[i].LevelStr]
		//牌型为连子的判断
		if 1 < b && b < 6 {
			cc = types.MapViewToLevel[result.LevelStr]
			dd = types.MapViewToLevel[handCards[i].LevelStr]
			//连子开头为A的特殊考虑
			if types.PokerViewNumA == result.LevelStr {
				cc = 1
			}
			if types.PokerViewNumA == handCards[i].LevelStr {
				dd = 1
			}
		}
		if cc < dd {
			result = handCards[i]
		}
	}
	return &result
}

// HandCardCompare @description: 比较两手牌大小(handCard1大于handCard2)
// @parameter handCard1
// @parameter handCard2
// @return bool
func HandCardCompare(handCard1 types.HandCard, handCard2 types.HandCard, mapViewToLevelH map[string]int) bool {
	aa := types.PokerHandLevelMap[handCard1.Type]
	bb := types.PokerHandLevelMap[handCard2.Type]
	if aa > bb {
		return true
	}
	if handCard1.Type != handCard2.Type {
		return false
	}
	//牌型不为连子level大小
	cc := mapViewToLevelH[handCard1.LevelStr]
	dd := mapViewToLevelH[handCard2.LevelStr]
	//牌型为连子的判断
	if 1 < handCard1.B && handCard1.B < 6 {
		cc = types.MapViewToLevel[handCard1.LevelStr]
		dd = types.MapViewToLevel[handCard2.LevelStr]
		//连子开头为A的特殊考虑
		if types.PokerViewNumA == handCard1.LevelStr {
			cc = 1
		}
		if types.PokerViewNumA == handCard2.LevelStr {
			dd = 1
		}
	}
	if cc > dd {
		return true
	}
	return false
}

// FourKingBomb @description: 获取四王炸
// @receiver p
// @return [][]types.Cards
func (p *Player) FourKingBomb() []types.HandCard {
	result := make([]types.HandCard, 0)
	p.Parse()
	count := len(p.Cards)
	if 4 > count {
		return result
	}
	var sum int
	tempList := p.Cards[count-4:]
	for _, card := range tempList {
		sum += types.MapViewToLevel[card.ViewNumber]
	}
	if 60 == sum {
		handCard := new(types.HandCard)
		handCard.A = 22
		handCard.B = 11
		handCard.Type = types.PokerHandJokerBoom
		handCard.LevelStr = types.SmallWang
		handCard.Name = types.SmallWang + types.BigWang + types.SmallWang + types.BigWang
		temp := make([]types.Cards, 0)
		temp = append(temp, tempList)
		handCard.PokerHands = temp
		result = append(result, *handCard)
	}
	return result
}

// GetThreeWithTwoHandCards @description: 获取所有三带二手牌
// @return AllPokerHandResp
func (p *Player) GetThreeWithTwoHandCards() []types.HandCard {
	result := make([]types.HandCard, 0)
	allResultMap := make(map[string]bool, 0)
	//获取3张相同的牌所有种类
	tt1 := p.ASameCardAll(3, 11)
	//获取2张相同的牌所有种类
	tt2 := p.ASameCardAll(2, 11)
	if 0 == len(tt1) || 0 == len(tt2) {
		return result
	}
	cardsList1 := make([]types.Cards, 0)
	cardsList2 := make([]types.Cards, 0)
	for _, value := range tt1 {
		cardsList1 = append(cardsList1, value...)
	}
	for _, value := range tt2 {
		cardsList2 = append(cardsList2, value...)
	}
	sets := make([][]types.Cards, 0)
	sets = append(sets, cardsList1)
	sets = append(sets, cardsList2)
	//排列三带二牌型所有组合
	result = allCombination(sets, p.Level, 32, 11, len(p.LCard), false, &allResultMap)
	//方便看，所有牌的种类,调试用
	mapTest := make(map[string]bool, 0)
	listTest := make([]string, 0)
	for _, a := range result {
		if _, ok := mapTest[a.Name]; !ok {
			listTest = append(listTest, a.Name)
			mapTest[a.Name] = true
		}
	}
	return result
}

// GetThreeWithTwoHandCardsGo @description: 获取所有三带二手牌(并发)
// @receiver p
// @parameter channel
// @parameter allResultMap
func (p *Player) GetThreeWithTwoHandCardsGo(channel chan []types.HandCard) {
	result := make([]types.HandCard, 0)
	allResultMap := make(map[string]bool, 0)
	//获取3张相同的牌所有种类
	tt1 := p.ASameCardAll(3, 11)
	//获取2张相同的牌所有种类
	tt2 := p.ASameCardAll(2, 11)
	if 0 == len(tt1) || 0 == len(tt2) {
		channel <- result
		return
	}
	cardsList1 := make([]types.Cards, 0)
	cardsList2 := make([]types.Cards, 0)
	for _, value := range tt1 {
		cardsList1 = append(cardsList1, value...)
	}
	for _, value := range tt2 {
		cardsList2 = append(cardsList2, value...)
	}
	sets := make([][]types.Cards, 0)
	sets = append(sets, cardsList1)
	sets = append(sets, cardsList2)
	//排列三带二牌型所有组合
	result = allCombination(sets, p.Level, 32, 11, len(p.LCard), false, &allResultMap)
	//方便看，所有牌的种类,调试用
	mapTest := make(map[string]bool, 0)
	listTest := make([]string, 0)
	for _, a := range result {
		if _, ok := mapTest[a.Name]; !ok {
			listTest = append(listTest, a.Name)
			mapTest[a.Name] = true
		}
	}
	channel <- result
}

// ContinuousPokerHand @description: 每种连续手牌所有不同结果集获取
// @receiver p
// @parameter a(每种牌点相同的张数)
// @parameter b(连续不同牌点的张数)
// @parameter state(是否是同花顺)
// @parameter allResultMap
// @return []types.HandCard
func (p *Player) ContinuousPokerHand(a, b int, state bool) []types.HandCard {
	result := make([]types.HandCard, 0)
	allResultMap := make(map[string]bool, 0)
	//获取a张相同的牌所有种类,按牌的大小进行分组
	samePokerHands := p.ASameCardAll(a, b)
	if 0 == len(samePokerHands) {
		return result
	}
	for i := 0; i < len(samePokerHands); i++ {
		if j := i + b - 1; j < len(samePokerHands) {
			//判断是否连续
			if continuousCondition(samePokerHands, i, j, a, b) {
				//排列此牌型所有组合
				temp := allCombination(samePokerHands[i:j+1], p.Level, a, b, len(p.LCard), state, &allResultMap)
				result = append(result, temp...)
			}
		}
	}
	//方便看，所有牌的种类,调试用
	//mapTest := make(map[string]bool, 0)
	//listTest := make([]string, 0)
	//for _, a := range result {
	//	if _, ok := mapTest[a.Name]; !ok {
	//		listTest = append(listTest, a.Name)
	//		mapTest[a.Name] = true
	//	}
	//}
	return result
}

// ContinuousPokerHandGo @description: 每种连续手牌所有不同结果集获取(并发)
// @receiver p
// @parameter channel
// @parameter a(每种牌点相同的张数)
// @parameter b(连续不同牌点的张数)
// @parameter state
// @parameter allResultMap
func (p *Player) ContinuousPokerHandGo(channel chan []types.HandCard, a, b int, state bool) {
	result := make([]types.HandCard, 0)
	allResultMap := make(map[string]bool, 0)
	//获取a张相同的牌所有种类,按牌的大小进行分组
	samePokerHands := p.ASameCardAll(a, b)
	if 0 == len(samePokerHands) {
		channel <- result
		return
	}
	for i := 0; i < len(samePokerHands); i++ {
		if j := i + b - 1; j < len(samePokerHands) {
			//判断是否连续
			if continuousCondition(samePokerHands, i, j, a, b) {
				//排列此牌型所有组合
				temp := allCombination(samePokerHands[i:j+1], p.Level, a, b, len(p.LCard), state, &allResultMap)
				result = append(result, temp...)
			}
		}
	}
	channel <- result
}

func (p Player) ASameCardAll(a, b int) [][]types.Cards {
	pairHands := make([][]types.Cards, 0)
	numbers := getPokerNumbers()
	aIndex := 0
	for _, viewNum := range numbers {
		card := p.getCards(viewNum, a, b)
		if len(card) >= a {
			sort.Sort(card)
			phc := p.DFS(card, a)
			pairHands = append(pairHands, phc)
			//记录A牌分组的index
			if (2 == b || 3 == b || 5 == b) && types.PokerViewNumA == viewNum {
				aIndex = len(pairHands) - 1
			}
		}
	}
	//把A手牌分组在result最前面也放置一份，因为有可能A,2连续牌型的可能
	if aIndex > 0 {
		tempList := make([][]types.Cards, 0)
		tempList = append(tempList, pairHands[aIndex])
		tempList = append(tempList, pairHands...)
		pairHands = tempList
	}
	return pairHands
}

// SameCardCombination @description:相同牌点数的a张所有组合
// @receiver p
// @parameter a
// @parameter b
// @parameter index
// @parameter pairHands
func (p Player) SameCardCombination(a, b int, cards []types.Card, aIndex *int, pairHands *[][]types.Cards) {
	tempCards := make([]types.Card, 0)
	viewNumber := cards[0].ViewNumber
	//不加通配牌的时候
	condition := 0 == len(p.LCard) || (a == 1 && b == 1) || types.PokerViewNumB == viewNumber || types.PokerViewNumR == viewNumber || p.Level == viewNumber
	if !condition {
		//通配牌放前面
		tempLCards := new(types.Cards)
		*tempLCards = p.LCard
		for i := 0; i < len(*tempLCards); i++ {
			(*tempLCards)[i].HViewNumber = viewNumber
		}
		tempCards = append(tempCards, *tempLCards...)
	}
	tempCards = append(tempCards, cards...)
	if len(tempCards) >= a {
		phc := p.DFS(tempCards, a)
		*pairHands = append(*pairHands, phc)
		//记录A牌分组的index
		if (2 == b || 3 == b || 5 == b) && types.PokerViewNumA == viewNumber {
			*aIndex = len(*pairHands) - 1
		}
	}
}

// continuousCondition @description: 判断是否连续的条件
// @parameter samePokerHands
// @parameter i
// @parameter j
// @parameter a(每种牌点相同的张数)
// @parameter b(连续不同牌点的张数)
// @return bool
func continuousCondition(samePokerHands [][]types.Cards, i, j, a, b int) bool {
	c1 := types.MapViewToLevel[samePokerHands[j][len(samePokerHands[j])-1][a-1].HViewNumber]-types.MapViewToLevel[samePokerHands[i][len(samePokerHands[i])-1][a-1].HViewNumber]+1 == b
	c2 := false
	//特殊处理含有A的连续的牌型条件
	if b > 1 && 0 == i {
		c2 = 14-(types.MapViewToLevel[samePokerHands[i][len(samePokerHands[i])-1][a-1].HViewNumber]-types.MapViewToLevel[samePokerHands[j][len(samePokerHands[j])-1][a-1].HViewNumber]) == b
	}
	return c1 || c2
}

// groupSamePokerHands @description:按牌点进行手牌分组
// @parameter a(每种牌点相同的张数)
// @parameter b(连续不同牌点的张数)
// @parameter samePokerHands
// @parameter lCard
// @return [][]types.Cards
func groupSamePokerHands(a, b int, samePokerHands []types.Cards) [][]types.Cards {
	result := make([][]types.Cards, 0)
	index := 0
	aIndex := 0
	//l := len(lCard)
	c := len(samePokerHands)
	for i := 0; i < c; i++ {
		//判断是否是同一个元素
		if samePokerHands[index][a-1].HViewNumber != samePokerHands[i][a-1].HViewNumber {
			result = append(result, samePokerHands[index:i])
			index = i
			lr := len(result)
			//记录A牌分组的index
			if b > 1 && types.PokerViewNumA == result[lr-1][0][0].HViewNumber {
				aIndex = lr - 1
			}
		}
		//判断是否是最后一个元素
		if i == c-1 {
			result = append(result, samePokerHands[index:])
			lr := len(result)
			//记录A牌分组的index
			if b > 1 && types.PokerViewNumA == result[lr-1][0][0].HViewNumber {
				aIndex = lr - 1
			}
		}
	}
	//把A手牌分组在result最前面也放置一份，因为有可能A,2连续牌型的可能
	if aIndex > 0 {
		tempList := make([][]types.Cards, 0)
		tempList = append(tempList, result[aIndex])
		tempList = append(tempList, result...)
		result = tempList
	}
	return result
}

// allCombination @description: 多个数组的排列组合（笛卡尔积算法）
// @parameter sets
// @parameter pokerViewNum
// @parameter aa
// @parameter bb
// @parameter count
// @parameter state
// @parameter allResultMap
// @return []types.HandCard
func allCombination(sets [][]types.Cards, pokerViewNum string, aa, bb, count int, state bool, allResultMap *map[string]bool) []types.HandCard {
	lens := func(i int) int { return len(sets[i]) }
	product := make([]types.HandCard, 0)
	for ix := make([]int, len(sets)); ix[0] < lens(0); nextIndex(ix, lens) {
		a := 0
		var r []types.Cards
		var name string
		var levelStr string
		sfMap := make(map[string]bool)
		for j, k := range ix {
			//一些牌型的特殊判断，是否符合
			if !handlingSpecialCases(sets, r, sfMap, pokerViewNum, aa, bb, j, k, state) {
				goto breakHere
			}
			r = append(r, sets[j][k])
			//获取所需的信息
			if !getNeededCardInfo(sets, pokerViewNum, j, k, count, &a, &name, &levelStr) {
				goto breakHere
			}
		}
		//去除重复的手牌
		if _, ok := (*allResultMap)[name]; !ok {
			handCard := new(types.HandCard)
			handCard.A = aa
			handCard.B = bb
			handCard.Type = getPokerHandType(aa, bb, state)
			handCard.LevelStr = levelStr
			handCard.Name = name
			handCard.PokerHands = r
			product = append(product, *handCard)
			(*allResultMap)[name] = true
		}
	breakHere:
		//log.Warnln("此种牌型组合不符合")
	}
	return product
}

func nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}

// handlingSpecialCases @description: 一些牌型的特殊判断，是否符合
// @parameter sets
// @parameter r
// @parameter sfMap
// @parameter pokerViewNum
// @parameter aa
// @parameter bb
// @parameter j
// @parameter k
// @parameter state
// @return bool
func handlingSpecialCases(sets [][]types.Cards, r []types.Cards, sfMap map[string]bool, pokerViewNum string, aa, bb, j, k int, state bool) bool {
	rl := len(r)
	//三带二牌型特殊处理(五张牌不能一样)
	if 32 == aa && 11 == bb && rl == 1 {
		tempMap := make(map[string]bool, 0)
		if !isLCard(r[0][0], pokerViewNum) {
			tempMap[r[0][0].ViewNumber] = true
		}
		if !isLCard(r[0][1], pokerViewNum) {
			tempMap[r[0][1].ViewNumber] = true
		}
		if !isLCard(r[0][2], pokerViewNum) {
			tempMap[r[0][2].ViewNumber] = true
		}
		if !isLCard(sets[j][k][0], pokerViewNum) {
			tempMap[sets[j][k][0].ViewNumber] = true
		}
		if !isLCard(sets[j][k][1], pokerViewNum) {
			tempMap[sets[j][k][1].ViewNumber] = true
		}
		if 1 == len(tempMap) {
			return false
		}
	}
	if state {
		//特殊情况，同花顺的判断
		if !isStraightFlush(sets, sfMap, j, k, pokerViewNum) {
			return false
		}
	}
	//顺子牌型特殊处理
	if 5 == bb && !state && 1 == aa {
		if !isLCard(sets[j][k][0], pokerViewNum) {
			sfMap[sets[j][k][0].Color] = true
		}
		if 4 == len(r) {
			//顺子牌型中去除同花顺牌型
			if len(sfMap) == 1 {
				return false
			}
		}
	}
	return true
}

// getNeededCardInfo @description: 获取所需的信息
// @parameter sets
// @parameter pokerViewNum
// @parameter j
// @parameter k
// @parameter count
// @parameter a
// @parameter name
// @parameter levelStr
// @return bool
func getNeededCardInfo(sets [][]types.Cards, pokerViewNum string, j, k, count int, a *int, name, levelStr *string) bool {
	for _, card := range sets[j][k] {
		//是否是通配牌
		if isLCard(card, pokerViewNum) {
			*a++
			//如果统计的通配牌的数量大于牌手手中的通配牌的数量就不符合
			if *a > count {
				return false
			}
		}
		//name的获取
		*name = *name + card.Color + card.ViewNumber
		//levelStr的获取
		if 0 == j {
			if 0 != len(card.HViewNumber) {
				*levelStr = card.HViewNumber
			} else {
				*levelStr = card.ViewNumber
			}
		}
	}
	return true
}

// isStraightFlush @description: 判断是否是同花顺
// @parameter sets
// @parameter sfMap
// @parameter j
// @parameter k
// @parameter pokerViewNum
// @return bool
func isStraightFlush(sets [][]types.Cards, sfMap map[string]bool, j, k int, pokerViewNum string) bool {
	if !isLCard(sets[j][k][0], pokerViewNum) {
		sfMap[sets[j][k][0].Color] = true
	}
	return len(sfMap) <= 1
}

// getPokerHandType @description: 根据a,b获取手牌类型
// @parameter a
// @parameter b
// @parameter is(是否是同花顺)
// @return string
func getPokerHandType(a, b int, is bool) string {
	//同花顺
	if is {
		return types.PokerHandStraightFlush
	}
	//3
	if 1 == a && 1 == b {
		return types.PokerHandSingle
	}
	//33
	if 2 == a && 1 == b {
		return types.PokerHandPair
	}
	//333
	if 3 == a && 1 == b {
		return types.PokerHandTrips
	}
	//333444
	if 3 == a && 2 == b {
		return types.PokerHandTwoTrips
	}
	//223344
	if 2 == a && 3 == b {
		return types.PokerHandThreePair
	}
	//33322
	if 32 == a && 11 == b {
		return types.PokerHandThreeWithTwo
	}
	//23456
	if 1 == a && 5 == b {
		return types.PokerHandStraight
	}
	//四张相同牌炸
	if 4 == a && 1 == b {
		return types.PokerHandBoomFour
	}
	//五张相同牌炸
	if 5 == a && 1 == b {
		return types.PokerHandBoomFive
	}
	//六张相同牌炸
	if 6 == a && 1 == b {
		return types.PokerHandBoomSix
	}
	//七张相同牌炸
	if 7 == a && 1 == b {
		return types.PokerHandBoomSeven
	}
	//八张相同牌炸
	if 8 == a && 1 == b {
		return types.PokerHandBoomEight
	}
	//九张相同牌炸
	if 9 == a && 1 == b {
		return types.PokerHandBoomNine
	}
	//十张相同牌炸
	if 10 == a && 1 == b {
		return types.PokerHandBoomTen
	}
	//四个王
	if 22 == a && 11 == b {
		return types.PokerHandJokerBoom
	}
	return ""
}

// Shuffle @description: 打乱两副牌
// @parameter data
func Shuffle(data types.Cards) {
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
func SendCards() []types.Cards {
	cardsList := make([]types.Cards, 4)
	Shuffle(types.TwoPlayCards)
	for index, card := range types.TwoPlayCards {
		temp := index % 4
		cardsList[temp] = append(cardsList[temp], card)
	}
	return cardsList
}

// IsLCard @description: 判断是否是通配牌
// @parameter card
// @parameter lCardNumber
// @return bool
func isLCard(card types.Card, lCardNumber string) bool {
	return lCardNumber == card.ViewNumber && types.PokerColorHearts == card.Color
}

// CardsToListString @description: []types.Card转换为[]string
// @parameter cards
// @return []string
func CardsToListString(cards []types.Card) []string {
	result := make([]string, 0)
	for _, card := range cards {
		result = append(result, card.Name)
	}
	return result
}

// ListStringToCards @description: []string转换为[]types.Card
// @parameter strList
// @parameter lStr
// @return []types.Card
// @return []types.Card
func ListStringToCards(strList []string, lStr string) ([]types.Card, []types.Card) {
	cards := make([]types.Card, 0)
	//通配牌
	lCards := make([]types.Card, 0)
	if 0 == len(strList) {
		return cards, lCards
	}
	for _, str := range strList {
		if 2 != len(str) {
			log.Errorf("cardName:%v err", str)
			return cards, lCards
		}
		types.CardIdAi++
		color := str[0:1]
		number := str[1:]
		card := new(types.Card)
		card.Id = types.CardIdAi
		card.Name = str
		card.Color = color
		card.ViewNumber = number
		card.HViewNumber = number
		cards = append(cards, *card)
		if isLCard(*card, lStr) {
			lCards = append(lCards, *card)
		}
	}
	return cards, lCards
}

// HandCardListToActionList @description: []types.HandCard转换为[]types.Action
// @parameter handCards
// @return []types.Action
func HandCardListToActionList(handCards []types.HandCard) []types.Action {
	result := make([]types.Action, 0)
	for _, handCard := range handCards {
		result = append(result, HandCardToAction(handCard))
	}
	return result
}

func HandCardToAction(handCard types.HandCard) types.Action {
	action := new(types.Action)
	if 0 == handCard.A {
		return *action
	}
	action.Type = handCard.Type
	tempStrings := make([]string, 0)
	for _, cards := range handCard.PokerHands {
		tempStrings = append(tempStrings, CardsToListString(cards)...)
	}
	action.Cards = tempStrings
	return *action
}

// GetAllPokerHandType @description: 获取玩家所有出牌的可能（必须出牌时）
// @parameter location
// @parameter playerName
// @parameter lCardNumber(通配牌牌点)
// @parameter cards
// @return []types.HandCard
func GetAllPokerHandType(location int, playerName, lCardNumber string, cards types.Cards) []types.HandCard {
	result := make([]types.HandCard, 0)
	lCards := make([]types.Card, 0)
	count := len(cards)
	if 1 > count {
		return nil
	}
	//获取手牌中的通配牌
	for _, lCard := range cards {
		if isLCard(lCard, lCardNumber) {
			lCards = append(lCards, lCard)
		}
	}
	//把一手牌当成一个玩家手中的所有牌来统一处理
	p := DefaultPlayer(location, playerName, lCardNumber, cards, lCards)
	channel1 := make(chan []types.HandCard)
	channel2 := make(chan []types.HandCard)
	channel3 := make(chan []types.HandCard)
	channel4 := make(chan []types.HandCard)
	channel5 := make(chan []types.HandCard)
	channel6 := make(chan []types.HandCard)
	channel7 := make(chan []types.HandCard)
	channel8 := make(chan []types.HandCard)
	//单张所有结果（3）
	go p.ContinuousPokerHandGo(channel1, 1, 1, false)
	//一对所有结果（33）
	go p.ContinuousPokerHandGo(channel2, 2, 1, false)
	////三联对所有结果（223344）
	go p.ContinuousPokerHandGo(channel3, 2, 3, false)
	////获取333，333444牌型所有可能
	go p.getPokerHandTrips(channel4)
	////三带二所有结果
	go p.GetThreeWithTwoHandCardsGo(channel5)
	////获取顺子，同花顺的所有可能
	go p.getPokerHandStraight(channel6)
	////获取相同牌炸弹所有可能
	go p.getPokerHandBoom(channel7)
	////四个王所有结果
	go p.analysisFourWangGo(channel8)
	handCard1 := <-channel1
	handCard2 := <-channel2
	handCard3 := <-channel3
	handCard4 := <-channel4
	handCard5 := <-channel5
	handCard6 := <-channel6
	handCard7 := <-channel7
	handCard8 := <-channel8
	if 0 != len(handCard1) {
		result = append(result, handCard1...)
	}
	if 0 != len(handCard2) {
		result = append(result, handCard2...)
	}
	if 0 != len(handCard3) {
		result = append(result, handCard3...)
	}
	if 0 != len(handCard4) {
		result = append(result, handCard4...)
	}
	if 0 != len(handCard5) {
		result = append(result, handCard5...)
	}
	if 0 != len(handCard6) {
		result = append(result, handCard6...)
	}
	if 0 != len(handCard7) {
		result = append(result, handCard7...)
	}
	if 0 != len(handCard8) {
		result = append(result, handCard8...)
	}
	return result
}

// getPokerHandStraight @description: 获取顺子，同花顺的所有可能
// @receiver p
// @parameter channel
func (p *Player) getPokerHandStraight(channel chan []types.HandCard) {
	result := make([]types.HandCard, 0)
	result = append(result, p.ContinuousPokerHand(1, 5, false)...) //顺子
	if 0 != len(result) {
		result = append(result, p.ContinuousPokerHand(1, 5, true)...) //同花顺
	}
	channel <- result
}

// getPokerHandTrips @description: 获取333，333444牌型所有可能
// @receiver p
// @return channel
func (p *Player) getPokerHandTrips(channel chan []types.HandCard) {
	result := make([]types.HandCard, 0)
	result = append(result, p.ContinuousPokerHand(3, 1, false)...) //333
	if 0 != len(result) {
		result = append(result, p.ContinuousPokerHand(3, 2, false)...) //333444
	}
	channel <- result
}

// getPokerHandBoom @description: 获取相同牌炸弹所有可能
// @receiver p
// @parameter channel
func (p *Player) getPokerHandBoom(channel chan []types.HandCard) {
	result := make([]types.HandCard, 0)
	tempPokerHand := p.ContinuousPokerHand(4, 1, false)
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(5, 1, false)
	}
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(6, 1, false)
	}
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(7, 1, false)
	}
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(8, 1, false)
	}
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(9, 1, false)
	}
	if 0 != len(tempPokerHand) {
		result = append(result, tempPokerHand...)
		tempPokerHand = p.ContinuousPokerHand(10, 1, false)
	}
	result = append(result, tempPokerHand...)
	channel <- result
}

// GetAllPokerHandTypeBigger @description: 获取玩家大于这一轮最大牌所有出牌的可能（可以不出）-待完善
// @parameter b
// @parameter location
// @parameter playerName
// @parameter lCardNumber(通配牌牌点)
// @parameter cards
// @parameter maxHandCard(这一轮最大的牌)
// @parameter mapViewToLevelH
// @return []types.HandCard
func GetAllPokerHandTypeBigger(location int, playerName, lCardNumber string, cards types.Cards, maxHandCard types.HandCard, mapViewToLevelH map[string]int) []types.HandCard {
	handCardsList := make([][]types.HandCard, 0)
	lCards := make([]types.Card, 0)
	count := len(cards)
	if 1 > count {
		return nil
	}
	//获取手牌中的通配牌
	for _, lCard := range cards {
		if isLCard(lCard, lCardNumber) {
			lCards = append(lCards, lCard)
		}
	}
	//把一手牌当成一个玩家手中的所有牌来统一处理
	p := DefaultPlayer(location, playerName, lCardNumber, cards, lCards)
	channel1 := make(chan []types.HandCard)
	channel2 := make(chan []types.HandCard)
	channel3 := make(chan []types.HandCard)
	channel4 := make(chan []types.HandCard)
	channel5 := make(chan []types.HandCard)
	channel6 := make(chan []types.HandCard)
	channel7 := make(chan []types.HandCard)
	channel8 := make(chan []types.HandCard)
	//单张所有结果（3）
	go p.ContinuousPokerHandGo(channel1, 1, 1, false)
	//一对所有结果（33）
	go p.ContinuousPokerHandGo(channel2, 2, 1, false)
	////三联对所有结果（223344）
	go p.ContinuousPokerHandGo(channel3, 2, 3, false)
	////获取333，333444牌型所有可能
	go p.getPokerHandTrips(channel4)
	////三带二所有结果
	go p.GetThreeWithTwoHandCardsGo(channel5)
	////获取顺子，同花顺的所有可能
	go p.getPokerHandStraight(channel6)
	////获取相同牌炸弹所有可能
	go p.getPokerHandBoom(channel7)
	////四个王所有结果
	go p.analysisFourWangGo(channel8)
	handCard1 := <-channel1
	handCard2 := <-channel2
	handCard3 := <-channel3
	handCard4 := <-channel4
	handCard5 := <-channel5
	handCard6 := <-channel6
	handCard7 := <-channel7
	handCard8 := <-channel8
	if 0 != len(handCard1) {
		handCardsList = append(handCardsList, handCard1)
	}
	if 0 != len(handCard2) {
		handCardsList = append(handCardsList, handCard2)
	}
	if 0 != len(handCard3) {
		handCardsList = append(handCardsList, handCard3)
	}
	if 0 != len(handCard4) {
		handCardsList = append(handCardsList, handCard4)
	}
	if 0 != len(handCard5) {
		handCardsList = append(handCardsList, handCard5)
	}
	if 0 != len(handCard6) {
		handCardsList = append(handCardsList, handCard6)
	}
	if 0 != len(handCard7) {
		handCardsList = append(handCardsList, handCard7)
	}
	if 0 != len(handCard8) {
		handCardsList = append(handCardsList, handCard8)
	}
	return GetBiggerPokerHandType(handCardsList, maxHandCard, mapViewToLevelH)
}

// GetBiggerPokerHandType @description: 获取更大的牌
// @parameter handCardsList
// @parameter maxHandCard
// @parameter mapViewToLevelH
// @return []types.HandCard
func GetBiggerPokerHandType(handCardsList [][]types.HandCard, maxHandCard types.HandCard, mapViewToLevelH map[string]int) []types.HandCard {
	result := make([]types.HandCard, 0)
	for _, handCards := range handCardsList {
		tempLen := len(handCards)
		if 0 != tempLen && types.PokerHandLevelMap[maxHandCard.Type] < types.PokerHandLevelMap[handCards[0].Type] {
			result = append(result, handCards...)
			continue
		}
		if 0 != tempLen && maxHandCard.Type == handCards[0].Type {
			condition := 2 == handCards[0].B || 3 == handCards[0].B || 5 == handCards[0].B
			cc := mapViewToLevelH[maxHandCard.LevelStr]
			for _, handCard := range handCards {
				//牌型不为连子level大小
				dd := mapViewToLevelH[handCard.LevelStr]
				//牌型为连子的判断
				if condition {
					cc = types.MapViewToLevel[maxHandCard.LevelStr]
					dd = types.MapViewToLevel[handCard.LevelStr]
					//连子开头为A的特殊考虑
					if types.PokerViewNumA == maxHandCard.LevelStr {
						cc = 1
					}
					if types.PokerViewNumA == handCard.LevelStr {
						dd = 1
					}
				}
				if cc < dd {
					result = append(result, handCard)
				}
			}
		}
	}
	return result
}
