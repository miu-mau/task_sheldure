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
			Balancer:     &kafka.Hash{}, // Используем Hash для распределения по ключу (task_id)
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

// KafkaConsumer для чтения задач из Kafka
type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// ReadMessage читает одно сообщение из Kafka
func (c *KafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// EnsureTopic создаёт топик с несколькими партициями для распределения нагрузки
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
		// Проверяем, существует ли топик
		existingPartitions, err2 := conn.ReadPartitions(topic)
		if err2 == nil && len(existingPartitions) > 0 {
			// Топик существует, проверяем количество партиций
			if len(existingPartitions) >= partitions {
				// У топика достаточно партиций
				return nil
			}
			// Топик существует, но с меньшим количеством партиций
			// В продакшене нужно было бы увеличить партиции, но для демо можно оставить
			return fmt.Errorf("topic exists but has %d partitions, need %d (delete and recreate)", len(existingPartitions), partitions)
		}
		// Топик не существует, но не удалось создать
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
