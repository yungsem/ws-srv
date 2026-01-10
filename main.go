package main

import (
	"ws-srv/handler"
	"ws-srv/ws"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	// 初始化 BucketManager
	connManager := ws.NewConnManager(16)

	// 初始化 HTTP 处理器
	var h handler.Handler

	// 创建并配置服务器
	s := g.Server()
	s.SetPort(8484)

	// 升级为 WebSocket 连接
	s.BindHandler("/", func(r *ghttp.Request) {
		h.Upgrade(r, connManager)
	})

	// 发送消息给指定客户端
	s.BindHandler("/SendMsgByClientId", func(r *ghttp.Request) {
		h.SendMsgByClientId(r, connManager)
	})

	// 发送消息给指定组
	s.BindHandler("/SendMsgByGroupId", func(r *ghttp.Request) {
		h.SendMsgByGroupId(r, connManager)
	})

	// 发送消息给所有客户端
	s.BindHandler("/SendMsgByAll", func(r *ghttp.Request) {
		h.SendMsgByAll(r, connManager)
	})

	// 统计连接数
	s.BindHandler("/summary", func(r *ghttp.Request) {
		h.Summary(r, connManager)
	})

	// 启动服务器
	s.Run()
}
