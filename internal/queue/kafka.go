package queue

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaProducer) PublishTask(ctx context.Context, key, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	return p.writer.WriteMessages(ctx, msg)
}
