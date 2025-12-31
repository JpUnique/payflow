package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JpUnique/payflow/internal/events"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &KafkaConsumer{reader: reader}
}

func (c *KafkaConsumer) Start(
	ctx context.Context,
	handler func(events.PaymentCreatedEvent) error,
) error {

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("[KAFKA] read_error err=%v", err)
			continue
		}

		var event events.PaymentCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[KAFKA] unmarshal_error err=%v", err)
			continue
		}

		log.Printf(
			"[KAFKA] event_received type=%s tx_id=%s",
			event.EventType,
			event.TransactionID,
		)

		if err := handler(event); err != nil {
			log.Printf(
				"[PROCESSOR] handler_failed tx_id=%s err=%v",
				event.TransactionID,
				err,
			)
		}
	}
}
