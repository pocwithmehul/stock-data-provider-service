package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	commonkafka "github.com/pocwithmehul/common-go-lib/pkg/kafka"
	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewKafkaConsumer(cfg *config.Config) (*commonkafka.Consumer, error) {
	return commonkafka.NewConsumer(commonkafka.ConsumerConfig{
		Brokers:       cfg.Kafka.Brokers,
		Topic:         cfg.Kafka.Topic,
		GroupID:       cfg.Kafka.GroupID,
		InitialOffset: sarama.OffsetOldest,
	})
}

func ConsumeStockEvents(ctx context.Context, consumer *commonkafka.Consumer, collection *mongo.Collection, logger *commonlogger.Logger) {
	for {
		err := consumer.Consume(ctx, func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			var event bson.M
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Error("failed parse kafka message", map[string]interface{}{"error": err.Error()})
				return nil
			}

			if _, err := collection.InsertOne(ctx, event); err != nil {
				logger.Error("mongo insert failed", map[string]interface{}{"error": err.Error(), "event": event})
				return nil
			}

			logger.Info("stored stock event", map[string]interface{}{"ticker": event["ticker"], "partition": msg.Partition})
			return nil
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("kafka consume error", map[string]interface{}{"error": err.Error()})
			time.Sleep(2 * time.Second)
		}
	}
}
