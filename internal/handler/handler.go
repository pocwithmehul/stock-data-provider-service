package handler

import (
	"context"
	"encoding/json"
	"net/http"

	commonlocale "github.com/pocwithmehul/common-go-lib/pkg/locale"
	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	commonmiddleware "github.com/pocwithmehul/common-go-lib/pkg/middleware"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"github.com/pocwithmehul/stock-data-provider-service/internal/db"
	"go.mongodb.org/mongo-driver/bson"
)

type StockQueryResponse struct {
	StockData []bson.M `json:"stockData"`
}

func GetStockHandler(collectionProvider *db.CollectionProvider, cfg *config.Config, logger *commonlogger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := commonmiddleware.ParseBearerToken(r, &cfg.TokenAuth)
		if err != nil {
			logger.Error("token validation failed", map[string]interface{}{"error": err.Error(), "locale": commonlocale.GetPreferredLocale(r)})
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
			"locale": commonlocale.GetPreferredLocale(r),
		})

		collection := collectionProvider.Get()
		if collection == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

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
