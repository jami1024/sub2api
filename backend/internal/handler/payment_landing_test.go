package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetLandingPackageShowcaseReturnsPublicPackageData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newPaymentHandlerLandingTestClient(t)
	ctx := t.Context()

	_, err := client.BalancePackage.Create().
		SetName("专属包-进阶级").
		SetPrice(100).
		SetCreditAmount(400).
		SetPackageScope(service.PackageScopeCodex).
		SetForSale(true).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Group.Create().
		SetName("gpt pro").
		SetPlatform(domain.PlatformOpenAI).
		SetStatus(domain.StatusActive).
		SetRateMultiplier(0.8).
		Save(ctx)
	require.NoError(t, err)

	configSvc := service.NewPaymentConfigService(client, nil, nil)
	h := NewPaymentHandler(nil, configSvc, nil)
	r := gin.New()
	r.GET("/api/v1/payment/public/landing-packages", h.GetLandingPackageShowcase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payment/public/landing-packages", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "专属包-进阶级")
	require.Contains(t, w.Body.String(), "综合低至 2 折")
	require.Contains(t, w.Body.String(), "gpt pro")
	require.Contains(t, w.Body.String(), "同样余额可多用约 25%")
}

func newPaymentHandlerLandingTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	db, err := sql.Open("sqlite", "file:payment_handler_landing?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
