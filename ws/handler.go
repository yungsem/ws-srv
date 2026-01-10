package ws

import (
	"context"

	"github.com/antlabs/quickws"
	"github.com/gogf/gf/v2/frame/g"
)

type WebSocketHandler struct {
	ctx         context.Context
	ClientId    string
	GroupId     string
	connManager *ConnManager
}

func NewWebSocketHandler(ctx context.Context, clientId string, groupId string, connManager *ConnManager) *WebSocketHandler {
	return &WebSocketHandler{
		ctx:         ctx,
		ClientId:    clientId,
		GroupId:     groupId,
		connManager: connManager,
	}
}

func (w *WebSocketHandler) OnOpen(c *quickws.Conn) {
	w.connManager.addConn(w.ClientId, w.GroupId, c)
	g.Log().Infof(w.ctx, "WebSocket OnOpen, clientId:%s\n", w.ClientId)
}

func (w *WebSocketHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	g.Log().Infof(w.ctx, "WebSocket OnMessage: %s, %v, clientId: %s\n", msg, op, w.ClientId)
}

func (w *WebSocketHandler) OnClose(c *quickws.Conn, err error) {
	g.Log().Infof(w.ctx, "WebSocket OnClose, clientId: %s, error: %v\n", w.ClientId, err)
	w.connManager.delConn(w.ClientId)
}
