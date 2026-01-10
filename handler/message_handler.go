package handler

import (
	"context"
	"errors"
	"fmt"
	"time"
	"ws-srv/model"
	"ws-srv/ws"

	"github.com/antlabs/quickws"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gvalid"
)

type Message struct {
	ClientId string
	GroupId  string
	Content  string `p:"content"  v:"required#content不能为空"`
}

// Handler 处理 HTTP 请求
type Handler struct {
}

// Upgrade 处理 WebSocket 升级请求
func (h *Handler) Upgrade(r *ghttp.Request, connManager *ws.ConnManager) {
	// 获取客户端传递的 clientId 参数
	clientId := r.Get("clientId", "default_client_id").String()
	groupId := r.Get("groupId", "").String()

	// 创建 WebSocket 处理器
	ctx := context.Background()
	handler := ws.NewWebSocketHandler(ctx, clientId, groupId, connManager)

	// 升级为 WebSocket 连接
	c, err := quickws.Upgrade(
		r.Response.Writer,
		r.Request,
		quickws.WithServerReplyPing(),
		quickws.WithServerCallback(handler),
		quickws.WithServerReadTimeout(30*time.Second),
	)
	if err != nil {
		g.Log().Errorf(ctx, "WebSocket Upgrade fail, err: %v, clientId:%s\n", err, clientId)
		return
	}

	// 启动读取循环
	c.StartReadLoop()
}

// SendMsgByClientId 发送消息给指定客户端
func (h *Handler) SendMsgByClientId(r *ghttp.Request, connManager *ws.ConnManager) {
	var msg *Message

	// 解析并校验请求参数
	if err := r.Parse(&msg); err != nil {
		// 校验错误处理
		var v gvalid.Error
		if errors.As(err, &v) {
			r.Response.WriteJsonExit(model.Result{
				Code:    1,
				Message: v.FirstError().Error(),
			})
			return
		}

		// 其他错误处理
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: err.Error(),
		})
	}

	if msg.ClientId == "" {
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: "clientId不能为空",
		})
		return
	}

	ctx := context.TODO()

	// 发送消息给指定的客户端
	err := connManager.SendMsgByClientId(ctx, msg.ClientId, msg.Content)
	if err != nil {
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	// 发送成功响应
	r.Response.WriteJsonExit(model.Result{
		Code:    0,
		Message: "消息发送成功",
	})
}

// SendMsgByGroupId 发送消息给指定组，组内所有客户端都收到消息
func (h *Handler) SendMsgByGroupId(r *ghttp.Request, connManager *ws.ConnManager) {
	var msg *Message

	// 解析并校验请求参数
	if err := r.Parse(&msg); err != nil {
		// 校验错误处理
		var v gvalid.Error
		if errors.As(err, &v) {
			r.Response.WriteJsonExit(model.Result{
				Code:    1,
				Message: v.FirstError().Error(),
			})
			return
		}

		// 其他错误处理
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: err.Error(),
		})
	}

	if msg.GroupId == "" {
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: "groupId不能为空",
		})
		return
	}

	ctx := context.TODO()

	// 发送消息给指定的客户端
	allFailedClientIds := connManager.SendMsgByGroupId(ctx, msg.GroupId, msg.Content)
	if allFailedClientIds != nil {
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: fmt.Sprintf("部分客户端消息发送失败，clientIds: %v", allFailedClientIds),
		})
		return
	}

	// 发送成功响应
	r.Response.WriteJsonExit(model.Result{
		Code:    0,
		Message: "消息发送成功",
	})
}

// SendMsgByAll 发送消息给所有客户端
func (h *Handler) SendMsgByAll(r *ghttp.Request, connManager *ws.ConnManager) {
	var msg *Message

	// 解析并校验请求参数
	if err := r.Parse(&msg); err != nil {
		// 校验错误处理
		var v gvalid.Error
		if errors.As(err, &v) {
			r.Response.WriteJsonExit(model.Result{
				Code:    1,
				Message: v.FirstError().Error(),
			})
			return
		}

		// 其他错误处理
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: err.Error(),
		})
	}

	ctx := context.TODO()

	// 发送消息给所有的客户端
	allFailedClientIds := connManager.SendMsgByAll(ctx, msg.Content)
	if allFailedClientIds != nil {
		r.Response.WriteJsonExit(model.Result{
			Code:    1,
			Message: fmt.Sprintf("部分客户端消息发送失败，clientIds: %v", allFailedClientIds),
		})
		return
	}

	// 发送成功响应
	r.Response.WriteJsonExit(model.Result{
		Code:    0,
		Message: "消息发送成功",
	})
}

// Summary 统计连接数
func (h *Handler) Summary(r *ghttp.Request, connManager *ws.ConnManager) {
	// 发送消息给指定的客户端
	sm := connManager.Summary()

	// 发送成功响应
	r.Response.WriteJsonExit(model.Result{
		Code:    0,
		Message: "成功",
		Data:    sm,
	})
}
