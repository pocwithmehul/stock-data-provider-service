package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	commontracer "github.com/pocwithmehul/common-go-lib/pkg/tracer"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"github.com/pocwithmehul/stock-data-provider-service/internal/db"
	"github.com/pocwithmehul/stock-data-provider-service/internal/handler"
	"github.com/pocwithmehul/stock-data-provider-service/internal/messaging"
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

func main() {
	commontracer.Start(commontracer.Config{Service: "stock-data-provider-service"})
	defer commontracer.Stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := commonlogger.NewLogger("stock-data-provider-service", cfg.Datadog)
	collectionProvider := &db.CollectionProvider{}
	go connectMongoWithRetry(cfg, collectionProvider, logger)

	router := muxtrace.NewRouter()
	router.HandleFunc("/healthz", handler.HealthHandler()).Methods(http.MethodGet)
	router.HandleFunc("/readyz", handler.ReadinessHandler(collectionProvider.Ready)).Methods(http.MethodGet)
	router.HandleFunc("/v1/getstock", handler.GetStockHandler(collectionProvider, cfg, logger)).Methods(http.MethodGet)

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

func connectMongoWithRetry(cfg *config.Config, collectionProvider *db.CollectionProvider, logger *commonlogger.Logger) {
	consumer, err := messaging.NewKafkaConsumer(cfg)
	if err != nil {
		log.Fatalf("create kafka consumer: %v", err)
	}

	for {
		mongoClient, err := db.ConnectMongo(context.Background(), cfg.Mongo.URI, cfg.Mongo.Database)
		if err != nil {
			logger.Error("mongo connect failed; retrying", map[string]interface{}{
				"error": err.Error(),
				"uri":   cfg.Mongo.URI,
			})
			time.Sleep(5 * time.Second)
			continue
		}

		collection := mongoClient.GetCollection(cfg.Mongo.Collection)
		collectionProvider.Set(collection)
		logger.Info("mongo connected", map[string]interface{}{
			"database":   cfg.Mongo.Database,
			"collection": cfg.Mongo.Collection,
		})

		go messaging.ConsumeStockEvents(context.Background(), consumer, collection, logger)
		return
	}
}
