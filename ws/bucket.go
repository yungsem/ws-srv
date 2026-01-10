package ws

import (
	"context"
	"fmt"
	"sync"

	"github.com/antlabs/quickws"
	"github.com/gogf/gf/v2/frame/g"
)

type connection struct {
	clientId string
	conn     *quickws.Conn
}

func newConnection(clientId string, conn *quickws.Conn) *connection {
	return &connection{
		clientId: clientId,
		conn:     conn,
	}
}

type bucket struct {
	sync.RWMutex
	connectionMap map[string]*connection
}

func newBucket() *bucket {
	return &bucket{
		connectionMap: make(map[string]*connection),
	}
}

func (b *bucket) addConnection(clientId string, conn *quickws.Conn) {
	newConn := newConnection(clientId, conn)
	b.Lock()
	defer b.Unlock()
	b.connectionMap[clientId] = newConn
}

func (b *bucket) delConnection(clientId string) {
	b.Lock()
	defer b.Unlock()
	delete(b.connectionMap, clientId)
}

func (b *bucket) getConnection(clientId string) (*connection, error) {
	b.RLock()
	defer b.RUnlock()
	conn, ok := b.connectionMap[clientId]
	if !ok {
		return nil, fmt.Errorf("connection not found, clientId: %s", clientId)
	}
	return conn, nil
}

func (b *bucket) sendMsgByClientId(ctx context.Context, clientId string, msg string) error {
	conn, err := b.getConnection(clientId)
	if err != nil {
		return err
	}

	err = conn.conn.WriteMessage(quickws.Text, []byte(msg))
	if err != nil {
		g.Log().Errorf(ctx, "failed to send message to clientId %s: %v\n", clientId, err)
		// 发送失败，可能连接已经断开，删除该连接
		b.delConnection(clientId)
		return fmt.Errorf("failed to send message to clientId %s: %v", clientId, err)
	}

	return nil
}

func (b *bucket) sendMsgByAll(ctx context.Context, msg string) []string {
	// 创建连接副本，避免在遍历过程中持有锁
	b.RLock()
	// 复制连接信息到临时变量
	connCopy := make(map[string]*connection)
	for k, v := range b.connectionMap {
		connCopy[k] = v
	}
	b.RUnlock()

	var failedClientIds []string
	// 遍历副本进行消息发送
	for clientId, conn := range connCopy {
		err := conn.conn.WriteMessage(quickws.Text, []byte(msg))
		if err != nil {
			g.Log().Errorf(ctx, "failed to send message to clientId %s: %v\n", clientId, err)
			// 发送失败，可能连接已经断开，删除该连接
			b.delConnection(clientId)
			failedClientIds = append(failedClientIds, clientId)
		}
	}

	return failedClientIds
}

func (b *bucket) count() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.connectionMap)
}
