package broker

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"websocket-server/protocol"

	"github.com/IBM/sarama"
)

// KafkaBroker .
// KafkaBroker 结构对应一整套 producer 和 consumer group 实例。
// 生产者 (producer)：1 个对象，线程安全，多 goroutine 并行用也没问题。
// 消费者 (ConsumerGroup)：1 个对象，由它内部再分配给各个分区，不需要你手动维护多个消费者。
// 如果你有多个主题、多个消费者组，这时候你就可以再创建多个 KafkaBroker 实例，每个实例对应不同主题和消费者组。
type KafkaBroker struct {
	producer sarama.AsyncProducer // 整套 Broker 实例内部共享一份 producer。
	group    sarama.ConsumerGroup
	topic    string
}

// NewKafkaBroker .
func NewKafkaBroker(brokers []string, topic string) (*KafkaBroker, error) {
	// 1. 配置生产者
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	// 2. 创建消费者组
	// consumer, err := sarama.NewConsumer(brokers, nil)
	group, err := sarama.NewConsumerGroup(brokers, "my-group-id", nil)
	if err != nil {
		return nil, err
	}

	// 3. 启动 goroutine 来监听生产结果
	go func() {
		for {
			select {
			case msg := <-producer.Successes():
				log.Printf("Message sent to topic %s partition %d at offset %d\n", msg.Topic, msg.Partition, msg.Offset)
			case err := <-producer.Errors():
				log.Printf("Failed to send message: %v\n", err)
			}
		}
	}()

	// 4. 返回自定义 KafkaBroker 实例
	return &KafkaBroker{
		producer: producer,
		group:    group,
		topic:    topic,
	}, nil
	// TODO: 还没写订阅逻辑，所以你之后要给 consumer 实现消息订阅。
}

// MessageProduce .
func (kb *KafkaBroker) MessageProduce(message *protocol.Message) error {
	// 序列化成 JSON
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: kb.topic,
		Value: sarama.ByteEncoder(bytes),
	}
	kb.producer.Input() <- msg
	return nil
}

// MessageConsume .
func (kb *KafkaBroker) MessageConsume(handler func(string, int64)) {
	// 使用 ConsumerGroup 来自动拉取各分区消息
	ctx := context.Background()
	handlerImpl := &consumerGroupHandler{handler: handler}
	for {
		// 一直循环调用，保证 rebalance 时也能重新开始
		if err := kb.group.Consume(ctx, []string{kb.topic}, handlerImpl); err != nil {
			log.Printf("Error from consumer group: %v\n", err)
			time.Sleep(time.Second) // 简单重试策略
		}
	}
}

// 实现 ConsumerGroupHandler
type consumerGroupHandler struct {
	handler func(string, int64)
}

func (h *consumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error   { return nil } // 开始前调用，比如你可以初始化一些东西。
func (h *consumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error { return nil } // 结束前调用，比如清理资源。
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 重点，这个是你真正拿到一批消息的地方，claim.Messages() 会不断流出消息，你要在这里遍历处理。
	for msg := range claim.Messages() {
		h.handler(string(msg.Value), msg.Offset) // 对每个消息调用你自定义传入的 handler
		sess.MarkMessage(msg, "")                // 告诉消费者组，这个消息我已经处理完，可以提交 offset 了
	}
	return nil
}

// // MessageConsume .
// func (kb *KafkaBroker) MessageConsume(handler func(string)) {
// 	partitionConsumer, err := kb.consumer.ConsumePartition(kb.topic, 0, sarama.OffsetNewest) // 分区的消费者
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer partitionConsumer.Close()

// 	for message := range partitionConsumer.Messages() {
// 		handler(string(message.Value))
// 	}
// }
