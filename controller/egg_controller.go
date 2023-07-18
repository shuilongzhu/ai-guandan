package controller

import (
	"ai-guandan/service"
	"ai-guandan/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ServeWs @description: Ai掼蛋websocket请求操作
// @parameter w
// @parameter r
func ServeWs(c *gin.Context) {
	conn, err := types.UpGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("UpGrader err:%v", err)
		return
	}
	client := &types.Client{Conn: *conn, UserInfo: types.UserInfo{}}
	service.ReadPump(client)
}
