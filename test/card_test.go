package test

import (
	"ai-guandan/service"
	"ai-guandan/types"
	log "github.com/sirupsen/logrus"
	"testing"
)

// TestGetAllPokerHandType @description: 获取玩家所有出牌的可能（必须出牌时）测试方法
// @parameter t
func TestGetAllPokerHandType(t *testing.T) {
	//生成27张牌
	cards := types.Cards{
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNum7, types.PokerViewNum7),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum8, types.PokerViewNum8),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumT, types.PokerViewNumT),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumT, types.PokerViewNumT),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumQ, types.PokerViewNumQ),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumQ, types.PokerViewNumQ),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNumQ, types.PokerViewNumQ),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNum9, types.PokerViewNum9),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNum9, types.PokerViewNum9),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum9, types.PokerViewNum9),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum9, types.PokerViewNum9),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNumA, types.PokerViewNumA),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumB, types.PokerViewNumB),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumR, types.PokerViewNumR),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum4, types.PokerViewNum4),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumA, types.PokerViewNumA),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumA, types.PokerViewNumA),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNum2, types.PokerViewNum2),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum4, types.PokerViewNum4),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNumA, types.PokerViewNumA),
	}
	//获取玩家所有出牌的可能（必须出牌时）
	handCards := service.GetAllPokerHandType(1, "P1", "2", cards)
	log.Println(len(handCards))
}

// TestGetAllPokerHandTypeBigger @description: //获取玩家大于这一轮最大牌所有出牌的可能（可以不出）测试方法
// @parameter t
func TestGetAllPokerHandTypeBigger(t *testing.T) {
	//生成14张牌
	cards := types.Cards{
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNum7, types.PokerViewNum7),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum8, types.PokerViewNum8),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNum6, types.PokerViewNum6),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumR, types.PokerViewNumR),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNum4, types.PokerViewNum4),
		types.NewCardWithId(types.PokerColorSpades, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorClubs, types.PokerViewNumK, types.PokerViewNumK),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNumA, types.PokerViewNumA),
		types.NewCardWithId(types.PokerColorHearts, types.PokerViewNum2, types.PokerViewNum2),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum4, types.PokerViewNum4),
		types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNumA, types.PokerViewNumA),
	}
	oldBigHandCard := types.HandCard{
		A:          1,
		B:          1,
		Type:       types.PokerHandSingle,
		LevelStr:   "6",
		Name:       "D6",
		PokerHands: []types.Cards{{types.NewCardWithId(types.PokerColorDiamonds, types.PokerViewNum6, types.PokerViewNum6)}},
	}
	mapViewToLevelH := map[string]int{
		types.PokerViewNum2: types.PokerLevel2,
		types.PokerViewNum3: types.PokerLevel3,
		types.PokerViewNum4: types.PokerLevel4,
		types.PokerViewNum5: types.PokerLevel5,
		types.PokerViewNum6: types.PokerLevel6,
		types.PokerViewNum7: types.PokerLevel7,
		types.PokerViewNum8: types.PokerLevel8,
		types.PokerViewNum9: types.PokerLevel9,
		types.PokerViewNumT: types.PokerLevel10,
		types.PokerViewNumJ: types.PokerLevelJ,
		types.PokerViewNumQ: types.PokerLevelQ,
		types.PokerViewNumK: types.PokerLevelK,
		types.PokerViewNumA: types.PokerLevelA,
		types.PokerViewNumB: types.PokerLevelB,
		types.PokerViewNumR: types.PokerLevelR,
	}
	//获取玩家大于这一轮最大牌所有出牌的可能（可以不出）
	handCards := service.GetAllPokerHandTypeBigger(1, "P2", "2", cards, oldBigHandCard, mapViewToLevelH)
	log.Println(len(handCards))
}
