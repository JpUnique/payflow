package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.Hash{},
	})

	return &KafkaProducer{writer: writer}
}

func (p *KafkaProducer) Publish(
	ctx context.Context,
	key string,
	event interface{},
) error {

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("[KAFKA] publish_failed err=%v", err)
		return err
	}

	log.Printf("[KAFKA] event_published key=%s", key)
	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
