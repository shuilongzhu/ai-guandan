package main

import (
	"ai-guandan/controller"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() { //函数入口，以main为方法名
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	// pprof
	pprof.Register(r, "/airiacloud/pprof")
	r.GET("/airiacloud/api/ai_whipped_egg/ws", controller.ServeWs)
	r.Run(":5022")
}
