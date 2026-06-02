package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pocwithmehul/common-go-lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StockQueryResponse struct {
	StockData []bson.M `json:"stockData"`
}

func GetStockHandler(collection *mongo.Collection, cfg *commonlib.Config, logger *commonlib.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := commonlib.ParseBearerToken(r, &cfg.TokenAuth)
		if err != nil {
			logger.Error("token validation failed", map[string]interface{}{"error": err.Error(), "locale": commonlib.GetPreferredLocale(r)})
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
			"locale": commonlib.GetPreferredLocale(r),
		})

		cursor, err := collection.Find(context.Background(), bson.M{"ticker": symbol})
		if err != nil {
			logger.Error("mongo query failed", map[string]interface{}{"error": err.Error()})
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

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
