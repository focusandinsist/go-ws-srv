# go-ws-srv
 
q:有socketIO为啥还做这个？
a:那垃圾socketIo频繁更新还版本不兼容，谁用谁sb

这里是 WebSocket 推送系统 的 文件架构设计，按照 清晰的模块划分，保证 高并发、可扩展、分布式支持，目标是 Socket.IO 替代品。

📂 文件结构
go-ws-srv/
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
│   ├── connection/           # WebSocket 服务器
│   │   ├── manager.go # 连接管理
│   │   ├── connection.go # 连接（存储在线用户、心跳检测）
│   ├── handler/           # WebSocket 服务器
│   │   ├── handler.go    # 处理 WebSocket 事件
│   ├── message/           # WebSocket 服务器
│   │   ├── manager.go # 消息管理
│   │   ├── message.go    # 消息（广播、私聊、频道）
│   ├── room/           # WebSocket 服务器
│   │   ├── room.go       # 频道/房间管理
│   ├── auth/           # WebSocket 服务器
│   │   ├── auth.go       # 认证（JWT / Token 验证）
│   ├── storage/          # 存储层
│   │   ├── redis.go      # Redis 存储用户状态 / 消息队列
│   │   ├── db.go         # MongoDB / MySQL 存储历史消息
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


📌 current features:
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





IM 系统的文件夹结构设计
一个典型的 IM 系统 的文件夹结构通常包括以下几部分：

/im-system
│
├── /cmd                     # 启动程序的入口
│   └── main.go              # 主程序入口
│
├── /config                  # 配置文件目录
│   └── config.yaml          # 配置文件，存储数据库、Redis、消息队列等配置信息
│
├── /internal                # 内部实现逻辑
│   ├── /auth                # 身份验证、注册、登录相关逻辑
│   ├── /message             # 消息相关的处理逻辑
│   ├── /user                # 用户信息的管理和存取
│   ├── /group               # 群组功能相关的逻辑
│   ├── /file                # 文件上传与下载处理
│   ├── /push                # 推送通知服务
│   └── /websocket           # WebSocket 连接管理
│
├── /pkg                     # 公共库文件，供不同模块使用
│   ├── /utils               # 工具函数
│   ├── /db                  # 数据库相关操作
│   ├── /redis               # Redis 连接及操作
│   └── /logger              # 日志记录
│
├── /api                     # 对外提供的 API 接口
│   ├── /v1                  # 第一版 API
│   ├── /v2                  # 第二版 API
│   └── /common              # 公共 API 和中间件
│
├── /scripts                 # 部署脚本、数据库迁移脚本等
│
└── /test                    # 测试代码
    ├── /unit                # 单元测试
    └── /integration         # 集成测试


