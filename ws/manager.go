package ws

import (
	"context"
	"fmt"
	"sync"

	"github.com/antlabs/quickws"
)

const defaultConnManagerBucketCount = 16

type ConnManager struct {
	// 内置的 16 个 bucket
	buckets []*bucket
	// 通过 groupId 管理的一组连接，使用 groups 管理
	// 一个 groupId 对应一个 bucket
	groups map[string]*bucket
	// groups 管理的连接索引，key: clientId, value: groupId
	clientIndex map[string]string
	count       int
	// 读写锁，用于保护 groups 和 clientIndex 的并发访问
	sync.RWMutex
}

func NewConnManager(count int) *ConnManager {
	if count <= 0 {
		count = defaultConnManagerBucketCount
	}

	buckets := make([]*bucket, count)
	for i := 0; i < count; i++ {
		buckets[i] = newBucket()
	}

	return &ConnManager{
		buckets:     buckets,
		groups:      make(map[string]*bucket),
		clientIndex: make(map[string]string),
		count:       count,
	}
}

func (m *ConnManager) addConn(clientId string, groupId string, conn *quickws.Conn) {
	if groupId == "" {
		bkt := m.getBucketByIndex(clientId)
		bkt.addConnection(clientId, conn)
		return
	}

	groupIdExist := m.isGroupIdExist(groupId)
	if groupIdExist {
		bkt := m.getBucketByGroupId(groupId)
		bkt.addConnection(clientId, conn)
		// 更新 clientIndex，记录该客户端所属的组
		m.Lock()
		m.clientIndex[clientId] = groupId
		m.Unlock()
		return
	}

	m.Lock()
	bkt := newBucket()
	bkt.addConnection(clientId, conn)
	m.groups[groupId] = bkt
	m.clientIndex[clientId] = groupId
	m.Unlock()
}

func (m *ConnManager) delConn(clientId string) {
	// 先读取 client 所属的 bucket 和组信息
	bkt, fromGroup := m.getBucket(clientId)
	// 从 bucket 中删除连接
	bkt.delConnection(clientId)

	// 如果是组内连接，需要进一步更新 groups 和 clientIndex 结构
	if fromGroup {
		m.delClientIndex(clientId)
		if bkt.count() == 0 {
			m.delGroup(clientId)
		}
	}
}

func (m *ConnManager) delClientIndex(clientId string) {
	m.Lock()
	defer m.Unlock()
	delete(m.clientIndex, clientId)
}

func (m *ConnManager) delGroup(clientId string) {
	// 写锁：需要修改 groups 和 clientIndex 数据结构
	m.Lock()
	defer m.Unlock()

	// 如果该组的连接数已为 0，删除整个组
	if groupId, ok := m.clientIndex[clientId]; ok {
		delete(m.groups, groupId)
	}
}

func (m *ConnManager) getConn(clientId string) (*connection, error) {
	bkt, _ := m.getBucket(clientId)
	return bkt.getConnection(clientId)
}

func (m *ConnManager) getBucketByIndex(clientId string) *bucket {
	// 这里使用简单的FNV-1a算法
	var h uint32 = 2166136261
	for i := 0; i < len(clientId); i++ {
		h ^= uint32(clientId[i])
		h *= 16777619
	}
	idx := h % uint32(m.count)

	m.RLock()
	defer m.RUnlock()
	return m.buckets[idx]
}

func (m *ConnManager) getBucketByGroupId(groupId string) *bucket {
	m.RLock()
	defer m.RUnlock()
	return m.groups[groupId]
}

func (m *ConnManager) getBucket(clientId string) (*bucket, bool) {
	groupId := m.getGroupId(clientId)
	if groupId != "" {
		return m.getBucketByGroupId(groupId), true
	}
	return m.getBucketByIndex(clientId), false
}

func (m *ConnManager) isGroupIdExist(groupId string) bool {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.groups[groupId]; ok {
		return true
	}
	return false
}

func (m *ConnManager) getGroupId(clientId string) string {
	m.RLock()
	defer m.RUnlock()
	if groupId, ok := m.clientIndex[clientId]; ok {
		return groupId
	}
	return ""
}

func (m *ConnManager) SendMsgByClientId(ctx context.Context, clientId string, content string) error {
	bkt, _ := m.getBucket(clientId)
	err := bkt.sendMsgByClientId(ctx, clientId, content)
	return err
}

func (m *ConnManager) SendMsgByGroupId(ctx context.Context, groupId string, content string) []string {
	// 读锁：查询 groups 结构，获取对应组的 bucket
	bkt := m.getBucketByGroupId(groupId)

	// 如果组不存在，返回 nil
	if bkt == nil {
		return nil
	}

	failedClientIds := bkt.sendMsgByAll(ctx, content)
	return failedClientIds
}

func (m *ConnManager) SendMsgByAll(ctx context.Context, content string) []string {
	// 遍历所有 bucket 发送消息
	var allFailedClientIds []string
	for _, bkt := range m.buckets {
		failedClientIds := bkt.sendMsgByAll(ctx, content)
		allFailedClientIds = append(allFailedClientIds, failedClientIds...)
	}

	for _, bkt := range m.groups {
		failedClientIds := bkt.sendMsgByAll(ctx, content)
		allFailedClientIds = append(allFailedClientIds, failedClientIds...)
	}

	return allFailedClientIds
}

func (m *ConnManager) Total() (int, int, int) {
	bktTotal := 0
	for _, bkt := range m.buckets {
		bktTotal += bkt.count()
	}

	// 读锁：遍历 groups 结构
	m.RLock()
	groupTotal := 0
	for _, bkt := range m.groups {
		groupTotal += bkt.count()
	}
	m.RUnlock()

	return bktTotal, groupTotal, bktTotal + groupTotal
}

func (m *ConnManager) Summary() map[string]any {
	sm := make(map[string]any)

	bktTotal, groupTotal, total := m.Total()
	sm["total"] = total
	sm["bucket_total"] = bktTotal
	sm["group_total"] = groupTotal

	// 读锁：遍历 groups 结构获取组详细信息
	m.RLock()
	groupMap := make(map[string]int)
	for groupId, bkt := range m.groups {
		groupMap[fmt.Sprintf("%s", groupId)] = bkt.count()
	}
	m.RUnlock()

	bktMap := make(map[string]int)
	for i, bkt := range m.buckets {
		bktMap[fmt.Sprintf("bucket_%d", i)] = bkt.count()
	}

	sm["bucket_detail"] = bktMap
	sm["group_detail"] = groupMap
	return sm
}
