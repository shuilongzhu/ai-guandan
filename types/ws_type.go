package types

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	WriteWait      = 1 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = 2 * time.Minute
	MaxMessageSize = 10000000

	RoleFarmer   = 0
	RoleLandlord = 1

	Ai1 = "A1"
	Ai2 = "A2"
	Pi1 = "P1"
	Pi2 = "P2"
)

var (
	Newline  = []byte{'\n'}
	Space    = []byte{' '}
	UpGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, //不验证origin，解决跨域问题
	}
	// RoomIds 房间号
	RoomIds = []RoomId{1, 2, 3}
	//TableIds 桌号
	TableIds = []TableId{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	//UserNames 固定用户名称
	UserNames = []string{"P1", "A1", "P2", "A2"}
	//RoomManagerObject 房间管理器
	RoomManagerObject = InitRoomManager()
	//PathLog 日志路径
	PathLog = "./log/GuanDanLog.log"
	//PathLog = "D:/temp/AiriaCloudLog/GuanDanLog.log"
)

type RoomId int
type TableId int
type UserId int

type RoomManager struct {
	Lock       sync.RWMutex
	Rooms      map[RoomId]*Room
	TableIdInc TableId
}

type Room struct {
	RoomId      int                `form:"room_id" json:"room_id"`
	Lock        sync.RWMutex       `form:"lock" json:"lock"`
	Tables      map[TableId]*Table `form:"tables" json:"tables"`
	EntranceFee int                `form:"entrance_fee" json:"entrance_fee"` //入场费
}

type Table struct {
	Lock         sync.RWMutex       `form:"lock" json:"lock"`
	TableId      int                `form:"table_id" json:"table_id"`
	Creator      *Client            `form:"creator" json:"creator"`
	TableClients map[string]*Client `form:"table_clients" json:"table_clients"`
	GameManage   GameManage         `form:"game_manage" json:"game_manage"`
}

type GameManage struct {
	LevelCardPoint      string         `form:"level_card_point" json:"level_card_point"`             //通配牌牌点
	LevelCardPointP     string         `form:"level_card_point_p" json:"level_card_point_p"`         //P方打的通配牌牌点
	LevelCardPointA     string         `form:"level_card_point_a" json:"level_card_point_a"`         //A方打的通配牌牌点
	MapPlayOrder        []string       `form:"map_play_order" json:"map_play_order"`                 // 出牌顺序
	FirstOutCard        string         `form:"first_out_card" json:"first_out_card"`                 //第一个出牌人的名字
	MaxHandCard         HandCard       `form:"max_hand_card" json:"max_hand_card"`                   //一轮最大手牌
	MaxHandCardPlayer   string         `form:"max_hand_card_player" json:"max_hand_card_player"`     //一轮最大手牌玩家名字
	AllIsEnterRoom      bool           `form:"all_is_enter_room" json:"all_is_enter_room"`           //此牌桌的所有玩家是否都进入并准备好；0：未准备好；1：已准备好
	OutAllCardUsers     []UserInfo     `form:"out_all_card_users" json:"out_all_card_users"`         //出完所有牌的玩家集合
	NextOutCard         string         `form:"next_out_card" json:"next_out_card"`                   //下一个出牌人的名字
	NextTributeUsers    []UserInfo     `form:"next_tribute_users" json:"next_tribute_users"`         //下一局进贡玩家集合
	NextBackUsers       []UserInfo     `form:"next_back_users" json:"next_back_users"`               //下一局还贡玩家集合
	TributeGetCardInfos []GetCardInfo  `form:"tribute_get_card_infos" json:"tribute_get_card_infos"` //进贡牌信息
	BackGetCardInfos    []GetCardInfo  `form:"back_get_card_infos" json:"back_get_card_infos"`       //还贡牌信息
	MapViewToLevelH     map[string]int `form:"map_view_to_level_h" json:"map_view_to_level_h"`       //牌的大小
}

type Client struct {
	Lock            sync.RWMutex    `form:"lock" json:"lock"`
	Conn            websocket.Conn  `form:"conn" json:"conn"`
	PlayerCardsInfo PlayerCardsInfo `form:"player_cards_info" json:"player_cards_info"`
	UserInfo        UserInfo        `form:"user_info" json:"user_info"`
	Room            Room            `form:"room" json:"room"`
	Ready           bool            `form:"ready" json:"ready"`
	IsCalled        bool            `form:"is_called" json:"is_called"`               //是否叫完分
	Next            *Client         `form:"next" json:"next"`                         //链表
	LatestHeartbeat int64           `form:"latest_heartbeat" json:"latest_heartbeat"` //最新心跳（时间戳）
}

type PlayerCardsInfo struct {
	Score       int    `form:"score" json:"score"`                 // 玩家得分
	LCardNumber string `form:"l_card_number" json:"l_card_number"` //通配牌牌点
	CardList    Cards  `form:"card_list" json:"card_list"`         //玩家手中剩余的所有牌
	LCard       []Card `form:"l_card" json:"l_card"`               //玩家手中剩余的通配牌
	//pokerHand guandan.PokerHand
}

type UserInfo struct {
	UserId   UserId  `form:"user_id" json:"user_id"`
	UserName string  `form:"user_name" json:"user_name"`
	UserRole int     `form:"user_role" json:"user_role"` // 1 ， 2 ；UserRole一样为一家
	Location int     `form:"location" json:"location"`
	TableId  TableId `form:"table_id" json:"table_id"`
	RoomId   RoomId  `form:"room_id" json:"room_id"` //1:四人线上模式；2:人机对战模式
}

type SendMsgObject struct {
	Lock   sync.RWMutex `form:"lock" json:"lock"`
	Id     int          `form:"id" json:"id"`
	Client *Client      `form:"client" json:"client"`
}

type GetCardInfo struct {
	UserInfo UserInfo `form:"user_info" json:"user_info"`
	Card     Card     `form:"card" json:"card"`
}
