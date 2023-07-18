package service

import (
	"ai-guandan/errorcode"
	"ai-guandan/types"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// WsRequest @description: 处理websocket具体请求
// @parameter req
// @parameter c
func WsRequest(req types.AiWeWsReq, c *types.Client) {
	defer catchMainError(c)
	switch req.Id {
	case types.TestClosedReqId: //测试关闭客户端
		TestClosedMethod(req, c)
	case types.TestConnectReqId: //测试连接客户端
		TestConnectMethod(req, c)
	case types.JoinRoomReqId: //加入房间
		JoinRoomMethod(req, c)
	case types.StartGameReqId: //开始游戏
		StartGameMethod(req, c)
	case types.ConfirmCancelReqId: //取消确认，在选角色时
		ConfirmCancelMethod(req, c)
	case types.OutCardReqId: //出牌
		OutCardMethod(req, c)
	case types.DecidedCardTypeReqId: //出的牌存在多种牌型，确认哪种牌型
		DecidedCardTypeMethod(req, c)
	case types.AgainOneGameReqId: //再来一局
		AgainOneGameMethod(c)
	case types.GetRemainTableIdReqId: //获取剩余未满的桌号
		GetRemainTableIdMethod(c)
	case types.ExitRoomReqId: //退出房间
		ExitRoomMethod(c)
	case types.TributeGetCardReqId: //进贡
		TributeGetCardMethod(req, c)
	case types.BackGetCardReqId: //还贡
		BackGetCardMethod(req, c)
	case types.ClientReconnectionReqId: //客户端断开重连
		ClientReconnectionMethod(req, c)
	case types.HeartbeatReqId: //客户端心跳
		HeartbeatMethod(c)
	case types.TakePictureUploadCardsReqId: //拍照上传手牌
		TakePictureUploadCardsMethod(req, c)
	case types.IsOnePTributeReqId: //P1或P2玩家反馈是否抗贡
		IsOnePTributeMethod(req, c)
	case types.InitTableReqId: //初始化table牌桌
		InitTableMethod(c)
	case types.SetHandCardReqId: //设置玩家手牌
		SetHandCardMethod(req, c)
	default:
		log.Errorln("非法入参id,请求拒绝！")
		c.WsSendMsg(types.IllegalParameterIdRespId, errorcode.ErrAWESeParameterId, nil)
	}
}

// catchMainError
// @description: 捕获主函数错误
func catchMainError(c *types.Client) {
	// 捕获全局错误
	if err := recover(); nil != err {
		log.Errorln("wsRequest panic:", err)
		c.WsSendMsg(-1, errorcode.ErrService, fmt.Sprintf("wsRequest panic:%v ", err))
	}
}
