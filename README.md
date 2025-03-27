# go-ws-srv
 
q:有socketIO为啥还做这个？
a:那垃圾socketIo频繁更新还版本不兼容，谁用谁sb

这里是 WebSocket 推送系统 的 文件架构设计，按照 清晰的模块划分，保证 高并发、可扩展、分布式支持，目标是 Socket.IO 替代品。

📂 文件结构
websocket-server/
│── cmd/                  # 入口目录
│   ├── main.go           # 启动 WebSocket 服务器
│
│── config/               # 配置文件
│   ├── config.yaml       # 配置文件 (端口、Redis 地址等)
│   ├── config.go         # 解析配置文件
│
│── internal/             # 核心业务逻辑
│   ├── server/           # WebSocket 服务器
│   │   ├── server.go     # WebSocket 服务器入口
│   │   ├── connection.go # 连接管理（存储在线用户、心跳检测）
│   │   ├── handler.go    # 处理 WebSocket 事件
│   │   ├── message.go    # 消息管理（广播、私聊、频道）
│   │   ├── room.go       # 频道/房间管理
│   │   ├── auth.go       # 认证（JWT / Token 验证）
│   │
│   ├── storage/          # 存储层
│   │   ├── redis.go      # Redis 存储用户状态 / 消息队列
│   │   ├── db.go         # MongoDB / MySQL 存储历史消息
│   │
│   ├── broker/           # 消息分发层
│   │   ├── redis_broker.go # 使用 Redis Pub/Sub 进行消息同步
│   │   ├── kafka_broker.go # 使用 Kafka 进行消息同步
│
│── api/                  # 提供 HTTP API
│   ├── rest/             # REST API
│   │   ├── user.go       # 获取在线用户
│   │   ├── message.go    # 发送消息接口
│   │
│   ├── websocket/        # WebSocket API
│   │   ├── ws_handler.go # WebSocket 处理入口
│
│── pkg/                  # 通用工具包
│   ├── logger/           # 日志组件
│   │   ├── logger.go     # 日志封装
│   ├── utils/            # 常用工具
│   │   ├── json.go       # JSON 处理
│   │   ├── uuid.go       # 生成唯一 ID
│
│── test/                 # 测试
│   ├── load_test.go      # 压测脚本（模拟 10w+ 连接）
│
│── go.mod                # Go 依赖管理
│── Dockerfile            # Docker 部署
│── README.md             # 项目文档


📌 主要设计思路
WebSocket 服务器 (internal/server/)

server.go：监听 WebSocket 连接，管理 goroutine

connection.go：存储在线用户、心跳检测、断线恢复

message.go：处理广播、私聊、房间内聊天

auth.go：支持 JWT 认证，保证连接安全

room.go：支持频道管理，用户可以订阅频道

消息分发 (internal/broker/)

redis_broker.go：用 Redis Pub/Sub 进行跨服务器消息同步

kafka_broker.go：用 Kafka 进行大规模消息分发

存储 (internal/storage/)

redis.go：用 Redis 存储在线用户、短期消息

db.go：MongoDB / MySQL 存储历史消息，支持消息回放

HTTP API (api/rest/)

user.go：查询在线用户

message.go：发送 WebSocket 消息（REST API）


🚀 下一步
实现 WebSocket 服务器 (internal/server/)

WebSocket 事件处理 (internal/server/handler.go)

使用 Redis 进行分布式支持 (internal/broker/redis_broker.go)
