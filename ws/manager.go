package ws

import (
	"context"
	"fmt"

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
		buckets: buckets,
		count:   count,
	}
}

func (m *ConnManager) AddGroupConnection(groupId string, clientId string, conn *quickws.Conn) {
	if m.groups == nil {
		m.groups = make(map[string]*bucket)
		m.clientIndex = make(map[string]string)
	} else if m.existGroup(groupId) {
		bkt := m.groups[groupId]
		bkt.addConnection(clientId, conn)
		m.clientIndex[clientId] = groupId
		return
	}

	bkt := newBucket()
	bkt.addConnection(clientId, conn)
	m.groups[groupId] = bkt
	m.clientIndex[clientId] = groupId
}

func (m *ConnManager) existGroup(groupId string) bool {
	if _, ok := m.groups[groupId]; ok {
		return true
	}
	return false
}

func (m *ConnManager) getBucketIndex(clientId string) uint32 {
	// 这里使用简单的FNV-1a算法
	var h uint32 = 2166136261
	for i := 0; i < len(clientId); i++ {
		h ^= uint32(clientId[i])
		h *= 16777619
	}
	return h % uint32(m.count)
}

func (m *ConnManager) getBucket(clientId string) (*bucket, bool) {
	// 检查该连接是否属于某个 group
	if groupId, ok := m.clientIndex[clientId]; ok {
		return m.getBucketByGroupId(groupId), true
	}
	// 否则使用内置的 bucket
	index := m.getBucketIndex(clientId)
	return m.buckets[index], false
}

func (m *ConnManager) getBucketByGroupId(groupId string) *bucket {
	return m.groups[groupId]
}

func (m *ConnManager) delConn(clientId string) {
	bkt, isGroup := m.getBucket(clientId)
	bkt.delConnection(clientId)
	if isGroup {
		if bkt.count() == 0 {
			delete(m.groups, m.clientIndex[clientId])
		}
		delete(m.clientIndex, clientId)
	}
}

func (m *ConnManager) getConn(clientId string) (*connection, error) {
	bkt, _ := m.getBucket(clientId)
	return bkt.getConnection(clientId)
}

func (m *ConnManager) addConn(clientId string, conn *quickws.Conn) {
	bkt, _ := m.getBucket(clientId)
	bkt.addConnection(clientId, conn)
}

func (m *ConnManager) SendMsgByClientId(ctx context.Context, clientId string, content string) error {
	bkt, _ := m.getBucket(clientId)
	err := bkt.sendMsgByClientId(ctx, clientId, content)
	return err
}

func (m *ConnManager) SendMsgByGroupId(ctx context.Context, groupId string, content string) []string {
	bkt := m.getBucketByGroupId(groupId)
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

	groupTotal := 0
	for _, bkt := range m.groups {
		groupTotal += bkt.count()
	}
	return bktTotal, groupTotal, bktTotal + groupTotal
}

func (m *ConnManager) Summary() map[string]any {
	sm := make(map[string]any)

	bktTotal, groupTotal, total := m.Total()
	sm["total"] = total
	sm["bucket_total"] = bktTotal
	sm["group_total"] = groupTotal

	groupMap := make(map[string]int)
	for groupId, bkt := range m.groups {
		groupMap[fmt.Sprintf("%s", groupId)] = bkt.count()
	}

	bktMap := make(map[string]int)
	for i, bkt := range m.buckets {
		bktMap[fmt.Sprintf("bucket_%d", i)] = bkt.count()
	}

	sm["bucket_detail"] = bktMap
	sm["group_detail"] = groupMap
	return sm
}
