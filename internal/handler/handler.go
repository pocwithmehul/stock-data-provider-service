package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	commonmiddleware "github.com/pocwithmehul/common-go-lib/pkg/middleware"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StockQueryResponse struct {
	StockData []bson.M `json:"stockData"`
}

func GetStockHandler(collection *mongo.Collection, cfg *config.Config, logger *commonlogger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := commonmiddleware.ParseBearerToken(r, &cfg.TokenAuth)
		if err != nil {
			logger.Error("token validation failed", map[string]interface{}{"error": err.Error(), "locale": preferredLocale(r)})
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			http.Error(w, "missing symbol", http.StatusBadRequest)
			return
		}

		logger.Info("getstock request", map[string]interface{}{
			"symbol": symbol,
			"user":   userInfo,
			"locale": preferredLocale(r),
		})

		cursor, err := collection.Find(context.Background(), bson.M{"ticker": symbol})
		if err != nil {
			logger.Error("mongo query failed", map[string]interface{}{"error": err.Error()})
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		defer func() {
			if closeErr := cursor.Close(context.Background()); closeErr != nil {
				logger.Error("mongo cursor close failed", map[string]interface{}{"error": closeErr.Error()})
			}
		}()

		var results []bson.M
		if err := cursor.All(context.Background(), &results); err != nil {
			logger.Error("decode mongo results failed", map[string]interface{}{"error": err.Error()})
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(StockQueryResponse{StockData: results})
	}
}

func preferredLocale(r *http.Request) string {
	accept := r.Header.Get("Accept-Language")
	if accept == "" {
		return "en-US"
	}

	parts := strings.Split(accept, ",")
	if len(parts) == 0 {
		return "en-US"
	}

	locale := strings.TrimSpace(parts[0])
	if locale == "" {
		return "en-US"
	}

	return locale
}
