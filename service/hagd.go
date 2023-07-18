package service

import (
	"ai-guandan/types"
	"sort"
)

func (p *Player) QueryTrip(a, b int) []types.Cards {
	return p.querySamePokerHandCards(a, b)
}

func (p *Player) querySamePokerHandCards(a, b int) []types.Cards {
	p.Parse()
	pairHands := make([]types.Cards, 0)
	numbers := getPokerNumbers()
	for _, viewNum := range numbers {
		card := p.getCards(viewNum, a, b)
		if len(card) >= a {
			phc := p.DFS(card, a)
			for _, pc := range phc {
				pairHands = append(pairHands, pc)
			}
		}
	}
	return pairHands
}

func getPokerNumbers() []string {
	return []string{
		types.PokerViewNum2, types.PokerViewNum3, types.PokerViewNum4, types.PokerViewNum5, types.PokerViewNum6, types.PokerViewNum7, types.PokerViewNum8, types.PokerViewNum9, types.PokerViewNumT, types.PokerViewNumJ, types.PokerViewNumQ, types.PokerViewNumK, types.PokerViewNumA, types.PokerViewNumB, types.PokerViewNumR,
	}
}

func (p *Player) Parse() {
	p.pokerHand = PokerHand{
		Level: p.Level,
		Cards: p.Cards,
		Name:  "player",
	}
	p.pokerHand.Parse()
}

func (ph *PokerHand) Parse() {
	ph.sort()
	ph.parse()
}

func (ph *PokerHand) sort() {
	sort.Sort(ph.Cards)
}

func (ph *PokerHand) parse() {

	if ph.parseDone {
		return
	}

	ph.Num = len(ph.Cards)
	ph.wildCard = newCard(types.PokerColorHearts, ph.Level)
	ph.minCard = types.Card{ViewNumber: types.PokerViewNumR}
	ph.maxCard = types.Card{ViewNumber: types.PokerViewNum2}
	ph.wildCardIndex = nil
	ph.wildCardNum = 0
	ph.wildCardIndex = make([]int, 2)
	ph.viewNumParsed = make(map[string]int, 10)
	ph.colorParsed = make(map[string]int, 10)

	for i := 0; i < ph.Num; i += 1 {

		// 统计通配牌的情况
		if Equal(ph.Cards[i], ph.wildCard) {
			ph.wildCardNum += 1
			ph.wildCardIndex = append(ph.wildCardIndex, i)
		} else if Compare(ph.Cards[i], ph.minCard) <= 0 {
			ph.minCard = ph.Cards[i]
		} else if Compare(ph.Cards[i], ph.maxCard) >= 0 {
			ph.maxCard = ph.Cards[i]
		}

		// 此处不统计 wildCard
		if !Equal(ph.Cards[i], ph.wildCard) {
			// 花色统计
			if _, ok := ph.colorParsed[ph.Cards[i].Color]; ok {
				ph.colorParsed[ph.Cards[i].Color] += 1
			} else {
				ph.colorParsed[ph.Cards[i].Color] = 1
			}

			// 统计数字情况
			if _, ok := ph.viewNumParsed[ph.Cards[i].ViewNumber]; ok {
				ph.viewNumParsed[ph.Cards[i].ViewNumber] += 1
			} else {
				ph.viewNumParsed[ph.Cards[i].ViewNumber] = 1
			}
		}

	}
	ph.parseDone = true
}

func newCard(c, n string) types.Card {
	return types.Card{Color: c, ViewNumber: n}
}

func Equal(a types.Card, b types.Card) bool {
	return b.Color == a.Color && b.ViewNumber == a.ViewNumber
}

func Compare(a types.Card, b types.Card) int {
	return types.MapViewToLevel[a.ViewNumber] - types.MapViewToLevel[b.ViewNumber]
}

func (p *Player) getCards(viewNum string, a, b int) types.Cards {
	cards := make([]types.Card, 0)
	//当牌为大小王或者出单张的时候做特殊处理，不加通配符
	if (a <= 1 && b <= 1) || types.PokerViewNumB == viewNum || types.PokerViewNumR == viewNum {
		for _, c := range p.Cards {
			if c.ViewNumber == viewNum {
				cards = append(cards, c)
			}
		}
	} else {
		for _, c := range p.Cards {
			condition := isLCard(c, p.Level)
			if c.ViewNumber == viewNum || condition {
				if condition {
					c.HViewNumber = viewNum
				}
				cards = append(cards, c)
			}
		}
	}
	return cards
}

func (p *Player) DFS(cards []types.Card, num int) []types.Cards {

	p.dfsCards = cards

	p.dfsCardDomain = len(p.dfsCards)
	p.dfsTargetNum = num
	p.dfsMasks = make([]int, p.dfsCardDomain)
	p.dfsTargetCard = make([]types.Card, p.dfsTargetNum)
	p.dfsResults = make([]types.Cards, 0)
	p.dfs(0, 0)
	return p.dfsResults
}

func (p *Player) dfs(u, m int) {
	if u == p.dfsTargetNum {
		cards := make([]types.Card, p.dfsTargetNum)
		for i := 0; i < p.dfsTargetNum; i++ {
			cards[i] = p.dfsTargetCard[i]
		}
		tempLen := len(p.dfsResults)
		if 0 == tempLen || !isSomeCards(p.dfsResults[tempLen-1], cards, p.dfsTargetNum) {
			p.dfsResults = append(p.dfsResults, cards)
		}
		return
	}
	for i := u; i < p.dfsCardDomain; i++ {

		if p.dfsMasks[i] == 0 && i >= m {
			p.dfsTargetCard[u] = p.dfsCards[i]
			p.dfsMasks[i] = 1
			p.dfs(u+1, i)
			p.dfsMasks[i] = 0
		}
	}
}

// isSomeCards @description: 判断前一个手牌是否和目前这个手牌相同
// @parameter pre
// @parameter cur
// @return bool
func isSomeCards(pre, cur []types.Card, len int) bool {
	for i := 0; i < len; i++ {
		if !Equal(pre[i], cur[i]) {
			return false
		}
	}
	return true
}
