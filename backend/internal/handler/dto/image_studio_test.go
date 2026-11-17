package dto

import (
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImageStudioAccountingIdentityNeverExposesCredential(t *testing.T) {
	key := &service.APIKey{ID: 42, Name: "Image Studio", Key: service.ImageStudioKeyPrefix + "private"}
	dto := APIKeyFromService(key)
	require.Empty(t, dto.Key)
	require.Equal(t, int64(42), dto.ID)
	require.Equal(t, "Image Studio", dto.Name)
	require.NotEmpty(t, key.Key)
}
