package types

import (
	"ai-guandan/errorcode"
	"ai-guandan/utils"
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"time"
)

type WsResponseTemplate struct {
	Id      int         `form:"id" json:"id"`
	Code    int         `form:"code" json:"code"`
	Data    interface{} `form:"data" json:"data"`
	Message string      `form:"message" json:"message" example:"响应信息"`
}

func (c *Client) WsSendMsg(respId int, code int, data interface{}) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	sendMsgObject := new(SendMsgObject)
	sendMsgObject.Client = c
	sendMsgObject.Id = respId
	msg := new(WsResponseTemplate)
	msg = WsResponseJson(respId, code, data)
	sendMsgObject.SendMsg(*msg)
}

// SendMsg @description: WebSocket发送数据通用方法
// @parameter c
// @parameter msg
func (sendMsgObject *SendMsgObject) SendMsg(msg WsResponseTemplate) {
	//sendMsgObject.Client.Lock.RLock()
	//defer sendMsgObject.Client.Lock.RUnlock()
	err := sendMsgObject.Client.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
	if err != nil {
		log.Errorf("send msg SetWriteDeadline err:%v", err)
		return
	}
	//判断WebSocket是否关闭
	if err := sendMsgObject.Client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		log.Errorf("%v->%v WebSocket has closed:", sendMsgObject.Client.Conn.LocalAddr(), sendMsgObject.Client.Conn.RemoteAddr())
		return
	}
	//打印发送的消息日志
	log.Printf("userInfo:%+v,sendMsg:%+v", sendMsgObject.Client.UserInfo, msg)
	msgByte, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("send msg [%v] marsha1 err:%v", string(msgByte), err)
		return
	}
	err = sendMsgObject.Client.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
	if err != nil {
		log.Errorf("send msg SetWriteDeadline [%v] err:%v", string(msgByte), err)
		return
	}
	w, err := sendMsgObject.Client.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		err = sendMsgObject.Client.Conn.Close()
		if err != nil {
			log.Errorf("close client err:%v", err)
		}
	}
	_, err = w.Write(msgByte)
	if err != nil {
		log.Errorf("Write msg [%v] err: %v", string(msgByte), err)
	}
	if err := w.Close(); err != nil {
		err = sendMsgObject.Client.Conn.Close()
		if err != nil {
			log.Errorf("close err: %v", err)
		}
	}
}

func WsResponseJson(id int, er int, data interface{}) *WsResponseTemplate {
	msg := errorcode.QueryErrorMessage(er)
	result := new(WsResponseTemplate)
	result.Id = id
	result.Code = er
	result.Data = data
	result.Message = msg
	if 200 != er {
		//请求统一错误日志输出
		errorInfo := utils.CallerInfo(4)
		log.Errorln(msg, errorInfo)
	}
	return result
}
