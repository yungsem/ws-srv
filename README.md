# ws-srv - 高性能 WebSocket 服务

基于 Go 语言开发的 WebSocket 服务，支持客户端管理、组消息广播、心跳检测等功能，具备高性能和高并发处理能力。

## 功能特性

- ✅ **WebSocket 连接管理**: 高效管理客户端连接
- ✅ **单客户端消息发送**: 支持向特定客户端发送消息
- ✅ **组消息广播**: 支持向特定组内所有客户端广播消息
- ✅ **全客户端广播**: 支持向所有连接的客户端发送消息
- ✅ **连接统计**: 实时统计连接数量和分布
- ✅ **分组管理**: 客户端可以分组，实现灵活的消息推送策略
- ✅ **心跳检测**: 内置 Ping/Pong 机制，维护连接状态
- ✅ **高性能架构**: 使用 bucket 分桶管理连接，支持高并发

## 技术栈

- **Go 语言**: 高性能、并发友好的开发语言
- **gogf 框架**: 现代化 Go 语言 Web 框架，提供便捷的 HTTP 服务和路由管理
- **quickws**: 高性能 WebSocket 库，支持快速的 WebSocket 升级和消息处理

## 项目结构

```
ws-srv/
├── client/              # 客户端示例文件
│   └── client.html      # 简单的 WebSocket 客户端示例
├── config/              # 配置文件目录
├── handler/             # HTTP 请求处理层
│   └── message_handler.go # 消息处理函数实现
├── model/               # 数据模型层
│   └── result.go       # 统一响应格式定义
├── ws/                  # WebSocket 核心业务逻辑层
│   ├── bucket.go       # 连接桶管理，用于并发处理
│   ├── handler.go      # WebSocket 事件处理
│   └── manager.go      # 连接管理
├── main.go             # 服务启动入口文件
├── go.mod              # Go 模块依赖管理
└── go.sum              # 依赖包版本锁定
```

## 快速开始

### 环境要求

- Go 1.16+ 环境

### 安装依赖

```bash
go mod download
```

### 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8484` 启动

### 测试连接

可以使用浏览器打开 `client/client.html` 文件，输入客户端 ID 和可选的组 ID 进行测试。

## API 文档

### WebSocket 连接端点

- **地址**: `ws://localhost:8484/`
- **参数**: 
  - `clientId`: 客户端唯一标识（可选，默认: "default_client_id"）
  - `groupId`: 组 ID（可选，用于分组消息发送）

### HTTP REST 端点

#### 1. 向指定客户端发送消息

**地址**: `POST /SendMsgByClientId`

**请求体**:
```json
{
  "ClientId": "client123",
  "Content": "Hello from server"
}
```

#### 2. 向指定组发送消息

**地址**: `POST /SendMsgByGroupId`

**请求体**:
```json
{
  "GroupId": "group1",
  "Content": "Hello to group members"
}
```

#### 3. 向所有客户端发送消息

**地址**: `POST /SendMsgByAll`

**请求体**:
```json
{
  "Content": "Hello to all connected clients"
}
```

#### 4. 获取连接统计信息

**地址**: `GET /summary`

**响应**:
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "total": 100,
    "bucket_total": 80,
    "group_total": 20,
    "bucket_detail": {
      "bucket_0": 10,
      "bucket_1": 15,
      ...
    },
    "group_detail": {
      "group1": 15,
      "group2": 5
    }
  }
}
```

## 使用示例

### 建立 WebSocket 连接

```javascript
const ws = new WebSocket('ws://localhost:8484/?clientId=myClientId&groupId=myGroup');

ws.onopen = function(event) {
    console.log('连接已建立');
};

ws.onmessage = function(event) {
    console.log('收到消息:', event.data);
};
```

### 通过 HTTP API 发送消息

```bash
curl -X POST http://localhost:8484/SendMsgByClientId \
  -H "Content-Type: application/json" \
  -d '{"ClientId":"myClientId","Content":"Hello from curl"}'
```

## 核心设计理念

### 连接管理

服务使用 `ConnManager` 结构体管理所有客户端连接，通过 bucket 分桶技术实现高效的并发处理。每个 bucket 内部使用读写锁保护连接数据，确保线程安全。

### 分组系统

支持将客户端分组管理，通过 `groupId` 参数可以将客户端加入特定组。这样可以方便地向组内所有成员同时发送消息。

### 心跳机制

内置 Ping/Pong 机制，自动维护连接状态，防止僵死连接占用资源。

## 性能优化

- 使用 bucket 分桶管理连接，减少锁竞争
- 使用读写锁区分读/写操作，提高并发性能
- 自动清理失效连接，避免资源泄漏
- 高效的哈希算法分配连接到不同桶

## 开发说明

### 项目依赖

查看 `go.mod` 文件了解所有依赖项。

### 日志

项目使用 `gogf/gf/v2/frame/g.Log()` 进行日志记录。

### 错误处理

所有 API 都有完善的错误处理，并返回统一的 JSON 错误格式。

## 部署

可以将项目编译为二进制文件直接部署，或使用 Docker 容器化部署。

### 编译

```bash
go build -o ws-srv main.go
```

### 运行

```bash
./ws-srv
```

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。