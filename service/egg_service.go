package service

import (
	"ai-guandan/types"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func ReadPump(c *types.Client) {
	defer func() {
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(types.MaxMessageSize)
	//c.Conn.SetReadDeadline(time.Now().Add(types.PongWait))
	//c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(types.PongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("userInfo:%+v,ReadPump() ReadMessage() err:%v", c.UserInfo, err)
			WsClosedDeal(c)
			return
		}
		aiWeWsReq := new(types.AiWeWsReq)
		err = json.Unmarshal(message, &aiWeWsReq)
		if 100 == aiWeWsReq.Id { //手牌上传包含图片，不打印日志
			log.Printf("userInfo:%+v,receiveMsg:{Id:100 Data:handCard picture}", c.UserInfo)
		} else {
			log.Println("userInfo:%+v,receiveMsg:%+v", c.UserInfo, *aiWeWsReq)
		}
		if err != nil {
			log.Errorf("message unmarsha1 err, user_id[%v] err:%v", c.UserInfo.UserId, err)
			return
		}
		WsRequest(*aiWeWsReq, c)
	}
}

// WsClosedDeal @description: Websocket关闭逻辑处理
// @parameter c
func WsClosedDeal(c *types.Client) {
	log.Errorf("WsClosedDeal() userInfo:%+v websocket has closed!", c.UserInfo)
	//客户端用户信息校验
	if state := CheckUserInfoPC(*c); !state {
		return
	}
	table := types.RoomManagerObject.Rooms[c.UserInfo.RoomId].Tables[c.UserInfo.TableId]
	//判断客户端是否是这一桌玩家最新的客户端
	if table.TableClients[c.UserInfo.UserName] == c {
		c.Ready = false
		table.GameManage.AllIsEnterRoom = false
	}
}
