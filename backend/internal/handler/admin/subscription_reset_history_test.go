package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type adminResetHistoryRepo struct {
	service.UserSubscriptionRepository
}

func (*adminResetHistoryRepo) GetByIDIncludeDeleted(_ context.Context, id int64) (*service.UserSubscription, error) {
	if id != 42 {
		return nil, service.ErrSubscriptionNotFound
	}
	now := time.Now()
	return &service.UserSubscription{ID: id, UserID: 7, DeletedAt: &now}, nil
}
func (*adminResetHistoryRepo) ListResetHistory(_ context.Context, _ int64, params pagination.PaginationParams) ([]service.SubscriptionResetHistory, *pagination.PaginationResult, error) {
	return []service.SubscriptionResetHistory{{ID: 1, Type: "admin", Weekly: true}}, &pagination.PaginationResult{Total: 1, Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func TestAdminSubscriptionResetHistoryIncludesRevokedSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubscriptionHandler(service.NewSubscriptionService(nil, &adminResetHistoryRepo{}, nil, nil, nil))
	for _, tc := range []struct {
		id     string
		status int
	}{{"42", 200}, {"43", 404}, {"0", 400}, {"bad", 400}} {
		t.Run(tc.id, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/subscriptions/"+tc.id+"/reset-history", nil)
			c.Params = gin.Params{{Key: "id", Value: tc.id}}
			h.GetResetHistory(c)
			require.Equal(t, tc.status, rec.Code)
			if tc.status == 200 {
				require.Contains(t, rec.Body.String(), `"type":"admin"`)
			}
		})
	}
}
