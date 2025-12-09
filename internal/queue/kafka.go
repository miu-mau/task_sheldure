package queue

import (
	"context"
	"fmt"
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
			Balancer:     &kafka.Hash{},
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

func (p *KafkaProducer) PublishTaskToTopic(ctx context.Context, topic string, key, value []byte) error {
	writer := &kafka.Writer{
		Addr:         p.writer.Addr,
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	return writer.WriteMessages(ctx, msg)
}

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

func (c *KafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

type KafkaMultiConsumer struct {
	readers []*kafka.Reader
	current int
}

func NewKafkaMultiConsumer(brokers []string, topics []string, groupID string) *KafkaMultiConsumer {
	readers := make([]*kafka.Reader, len(topics))
	for i, topic := range topics {
		readers[i] = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		})
	}

	return &KafkaMultiConsumer{
		readers: readers,
		current: 0,
	}
}

func (c *KafkaMultiConsumer) Close() error {
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (c *KafkaMultiConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if len(c.readers) == 0 {
		return kafka.Message{}, fmt.Errorf("no readers configured")
	}

	type result struct {
		msg kafka.Message
		err error
	}

	results := make(chan result, len(c.readers))
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, reader := range c.readers {
		go func(r *kafka.Reader) {
			msg, err := r.ReadMessage(cancelCtx)
			select {
			case results <- result{msg: msg, err: err}:
			case <-cancelCtx.Done():

			}
		}(reader)
	}

	receivedCount := 0
	var lastErr error

	for receivedCount < len(c.readers) {
		select {
		case res := <-results:
			receivedCount++
			if res.err == nil {

				cancel()
				return res.msg, nil
			}

			lastErr = res.err
		case <-ctx.Done():
			cancel()
			return kafka.Message{}, ctx.Err()
		}
	}

	cancel()
	return kafka.Message{}, lastErr
}

func EnsureTopic(brokers []string, topic string, partitions int) error {
	conn, err := kafka.DialContext(context.Background(), "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}
	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {

		existingPartitions, err2 := conn.ReadPartitions(topic)
		if err2 == nil && len(existingPartitions) > 0 {

			if len(existingPartitions) >= partitions {

				return nil
			}

			return fmt.Errorf("topic exists but has %d partitions, need %d (delete and recreate)", len(existingPartitions), partitions)
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
