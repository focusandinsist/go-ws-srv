1.msg后续改为proto协议
    你可以开始时直接使用结构体定义消息，等到系统逐渐扩展时，转换到 Protobuf 或其他协议。
    甚至可以在系统中同时使用两者，内部通信使用 Protobuf，外部接口（如 WebSocket、HTTP）则使用 JSON 格式。