package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pocwithmehul/common-go-lib"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg, err := commonlib.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := commonlib.NewLogger("stock-data-provider-service", cfg.Datadog)
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		logger.Error("mongo connect failed", map[string]interface{}{"error": err.Error()})
		log.Fatalf("mongo connect: %v", err)
	}

	collection := mongoClient.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection)
	reader := NewKafkaReader(cfg)

	go ConsumeStockEvents(reader, collection, logger)

	router := mux.NewRouter()
	router.HandleFunc("/v1/getstock", GetStockHandler(collection, cfg, logger)).Methods(http.MethodGet)

	port := cfg.Server.Port
	if port == 0 {
		port = 8082
	}
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Info("starting stock-data-provider-service", map[string]interface{}{"addr": addr})
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", map[string]interface{}{"error": err.Error()})
		log.Fatalf("server failed: %v", err)
	}
}
