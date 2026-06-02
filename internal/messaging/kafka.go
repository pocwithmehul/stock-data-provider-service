package messaging

import (
	"context"
	"encoding/json"
	"time"

	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewKafkaReader(cfg *config.Config) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Kafka.Brokers,
		Topic:       cfg.Kafka.Topic,
		GroupID:     cfg.Kafka.GroupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
}

func ConsumeStockEvents(reader *kafka.Reader, collection *mongo.Collection, logger *commonlogger.Logger) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("kafka read failed", map[string]interface{}{"error": err.Error()})
			time.Sleep(2 * time.Second)
			continue
		}

		var event bson.M
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("failed parse kafka message", map[string]interface{}{"error": err.Error()})
			continue
		}

		_, err = collection.InsertOne(context.Background(), event)
		if err != nil {
			logger.Error("mongo insert failed", map[string]interface{}{"error": err.Error(), "event": event})
			continue
		}

		logger.Info("stored stock event", map[string]interface{}{"ticker": event["ticker"], "partition": msg.Partition})
	}
}
