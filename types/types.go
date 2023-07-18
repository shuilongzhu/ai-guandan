package types

const (
	PokerColorHearts   = "H"
	PokerColorSpades   = "S"
	PokerColorDiamonds = "D"
	PokerColorClubs    = "C"
	SmallWang          = "SB"
	BigWang            = "HR"
)

const (
	PokerHandSingle        = "Single"        //3
	PokerHandPair          = "Pair"          //33
	PokerHandTrips         = "Trips"         //333
	PokerHandTwoTrips      = "TwoTrips"      // 333444
	PokerHandThreePair     = "TreePair"      // 223344
	PokerHandThreeWithTwo  = "ThreeWithTwo"  //33322
	PokerHandStraight      = "Straight"      //23456
	PokerHandBoom          = "Boom"          //炸弹
	PokerHandBoomFour      = "BoomFour"      //四张相同牌炸
	PokerHandBoomFive      = "BoomFive"      //五张相同牌炸
	PokerHandBoomSix       = "BoomSix"       //六张相同牌炸
	PokerHandBoomSeven     = "BoomSeven"     //七张相同牌炸
	PokerHandBoomEight     = "BoomEight"     //八张相同牌炸
	PokerHandBoomNine      = "BoomNine"      //九张相同牌炸
	PokerHandBoomTen       = "BoomTen"       //十张相同牌炸
	PokerHandStraightFlush = "StraightFlush" //同花顺(5张)
	PokerHandJokerBoom     = "JokerBoom"     //四个王
	PokerHandPASS          = "PASS"          //过牌
	PokerHandTribute       = "tribute"       //进贡
	PokerHandBack          = "back"          //回贡
	PokerHandError         = "Error"
	PokerHandMan           = "Manual"
)

const (
	PokerViewNum2 = "2"
	PokerViewNum3 = "3"
	PokerViewNum4 = "4"
	PokerViewNum5 = "5"
	PokerViewNum6 = "6"
	PokerViewNum7 = "7"
	PokerViewNum8 = "8"
	PokerViewNum9 = "9"
	PokerViewNumT = "T"
	PokerViewNumJ = "J"
	PokerViewNumQ = "Q"
	PokerViewNumK = "K"
	PokerViewNumA = "A"
	PokerViewNumB = "B"
	PokerViewNumR = "R"
)

const (
	PokerLevel2 = iota + 2
	PokerLevel3
	PokerLevel4
	PokerLevel5
	PokerLevel6
	PokerLevel7
	PokerLevel8
	PokerLevel9
	PokerLevel10
	PokerLevelJ
	PokerLevelQ
	PokerLevelK
	PokerLevelA
)
const (
	PokerLevelB = 16
	PokerLevelR = 18
)

var (
	MapViewToLevel    map[string]int //牌的顺序
	MapLevelToView    map[int]string
	PokerHandLevelMap map[string]int
	TwoPlayCards      Cards
	CardId            int
	CardIdAi          int
)

type Cards []Card

func (cards Cards) Len() int {
	return len(cards)
}
func (cards Cards) Less(i, j int) bool {
	if cards[i].ViewNumber == cards[j].ViewNumber {
		return cards[i].Color < cards[j].Color
	}
	return MapViewToLevel[cards[i].ViewNumber] < MapViewToLevel[cards[j].ViewNumber]
}

func (cards Cards) Swap(i, j int) {
	// @TODO
	// 这个与结构体赋值的区别是什么
	cards[i], cards[j] = cards[j], cards[i]
}

type Card struct {
	Id          int    `from:"id" json:"id"`
	Name        string `from:"name" json:"name"`
	Color       string `from:"color" json:"color"`
	ViewNumber  string `from:"viewNumber" json:"viewNumber"`
	HViewNumber string `from:"hViewNumber" json:"hViewNumber"` //如果是通配牌，它所代表的牌点
	//Level       int    `from:"level" json:"level"`
}

func NewCardWithId(c, n, hn string) Card {
	CardId++
	return Card{Id: CardId, Color: c, ViewNumber: n, HViewNumber: hn, Name: c + n}
}

type HandCard struct {
	A          int     `json:"a"`         //每种牌点相同的张数
	B          int     `json:"b"`         //连续不同牌点的张数
	Type       string  `json:"type"`      //手牌的类型
	LevelStr   string  `json:"level_str"` //按从小到大顺序手牌最开始的牌点，用于比较手牌大小
	Name       string  `json:"name"`
	PokerHands []Cards `json:"poker_hands"`
}
