package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"github.com/pocwithmehul/stock-data-provider-service/internal/db"
	"github.com/pocwithmehul/stock-data-provider-service/internal/handler"
	"github.com/pocwithmehul/stock-data-provider-service/internal/messaging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := commonlogger.NewLogger("stock-data-provider-service", cfg.Datadog)
	mongoClient, err := db.ConnectMongo(context.Background(), cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		logger.Error("mongo connect failed", map[string]interface{}{"error": err.Error()})
		log.Fatalf("mongo connect: %v", err)
	}

	collection := mongoClient.GetCollection(cfg.Mongo.Collection)
	reader := messaging.NewKafkaReader(cfg)

	go messaging.ConsumeStockEvents(reader, collection, logger)

	router := mux.NewRouter()
	router.HandleFunc("/v1/getstock", handler.GetStockHandler(collection, cfg, logger)).Methods(http.MethodGet)

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
