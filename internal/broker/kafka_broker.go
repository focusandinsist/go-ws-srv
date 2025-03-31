package broker

import (
	"log"

	"github.com/IBM/sarama"
)

type KafkaBroker struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	topic    string
}

func NewKafkaBroker(brokers []string, topic string) (*KafkaBroker, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaBroker{
		producer: producer,
		consumer: consumer,
		topic:    topic,
	}, nil
}

func (kb *KafkaBroker) SendMessage(message string) error {
	msg := &sarama.ProducerMessage{
		Topic: kb.topic,
		Value: sarama.StringEncoder(message),
	}
	_, _, err := kb.producer.SendMessage(msg)
	return err
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
