package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	commonlogger "github.com/pocwithmehul/common-go-lib/pkg/logger"
	"github.com/pocwithmehul/stock-data-provider-service/internal/config"
	"github.com/pocwithmehul/stock-data-provider-service/internal/db"
)

func newTestLogger() *commonlogger.Logger {
	return commonlogger.NewLogger("test", commonlogger.DatadogConfig{})
}

func TestGetStockHandler_Unauthorized_NoToken(t *testing.T) {
	provider := &db.CollectionProvider{}
	cfg := &config.Config{}
	logger := newTestLogger()

	req := httptest.NewRequest(http.MethodGet, "/v1/getstock?symbol=AAPL", nil)
	w := httptest.NewRecorder()

	GetStockHandler(provider, cfg, logger)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetStockHandler_Unauthorized_InvalidToken(t *testing.T) {
	provider := &db.CollectionProvider{}
	cfg := &config.Config{}
	logger := newTestLogger()

	req := httptest.NewRequest(http.MethodGet, "/v1/getstock?symbol=AAPL", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	GetStockHandler(provider, cfg, logger)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
