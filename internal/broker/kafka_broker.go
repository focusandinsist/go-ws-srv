package broker

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaBroker struct {
	producer sarama.AsyncProducer
	consumer sarama.Consumer
	topic    string
}

func NewKafkaBroker(brokers []string, topic string) (*KafkaBroker, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, err
	}

	// 处理成功和错误的返回
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

	return &KafkaBroker{
		producer: producer,
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (kb *KafkaBroker) SendMessage(message string) {
	msg := &sarama.ProducerMessage{
		Topic: kb.topic,
		Value: sarama.StringEncoder(message),
	}
	kb.producer.Input() <- msg
}

func (kb *KafkaBroker) ConsumeMessages(handler func(string)) {
	partitionConsumer, err := kb.consumer.ConsumePartition(kb.topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err)
	}

	defer partitionConsumer.Close()

	for message := range partitionConsumer.Messages() {
		handler(string(message.Value))
	}
}

func (kb *KafkaBroker) ConsumeUserMessages(userID string, handler func(string)) {
	partitionConsumer, err := kb.consumer.ConsumePartition(kb.topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err)
	}

	defer partitionConsumer.Close()

	for message := range partitionConsumer.Messages() {
		if string(message.Key) == userID {
			handler(string(message.Value))
		}
	}
}
