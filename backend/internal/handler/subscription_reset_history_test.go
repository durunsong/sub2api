package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type resetHistoryAPIRepo struct {
	resetCardAPIRepo
	historyCalls int
}

func (r *resetHistoryAPIRepo) ListResetHistory(_ context.Context, id int64, params pagination.PaginationParams) ([]service.SubscriptionResetHistory, *pagination.PaginationResult, error) {
	r.historyCalls++
	return []service.SubscriptionResetHistory{{ID: 9, Type: "reset_card", ResetAt: time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC), Daily: true, Weekly: true, Monthly: true, ValidityDays: 7}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func TestSubscriptionResetHistory_OwnershipAndResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		userID int64
		id     string
		status int
	}{
		{"owner", 7, "42", http.StatusOK},
		{"other user", 8, "42", http.StatusNotFound},
		{"missing", 7, "43", http.StatusNotFound},
		{"invalid id", 7, "bad", http.StatusBadRequest},
		{"negative id", 7, "-1", http.StatusBadRequest},
		{"unauthenticated", 0, "42", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &resetHistoryAPIRepo{resetCardAPIRepo: resetCardAPIRepo{subscriptions: map[int64]*service.UserSubscription{42: {ID: 42, UserID: 7}}}}
			h := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
			c, rec := resetCardAPIContext(http.MethodGet, "/api/v1/subscriptions/42/reset-history?page=2&page_size=999", nil, tc.userID)
			c.Params = gin.Params{{Key: "id", Value: tc.id}}
			if tc.userID == 0 {
				c.Keys = nil
			}
			h.GetResetHistory(c)
			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			if tc.status != http.StatusOK {
				require.Zero(t, repo.historyCalls)
				return
			}
			var body struct {
				Data struct {
					Items    []service.SubscriptionResetHistory `json:"items"`
					Page     int                                `json:"page"`
					PageSize int                                `json:"page_size"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, 2, body.Data.Page)
			require.Equal(t, 100, body.Data.PageSize)
			require.Equal(t, 7, body.Data.Items[0].ValidityDays)
			require.NotContains(t, rec.Body.String(), "source_reference")
			require.NotContains(t, rec.Body.String(), "idempotency")
		})
	}
}
