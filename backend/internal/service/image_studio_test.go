package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImageStudioCredentialCannotAuthenticateExternally(t *testing.T) {
	svc := &APIKeyService{}
	_, err := svc.GetByKey(context.Background(), ImageStudioKeyPrefix+"secret")
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.Error(t, svc.ValidateCustomKey(ImageStudioKeyPrefix+"secret"))
}
func TestEphemeralModerationLogOmitsContentAndErrors(t *testing.T) {
	svc := &ContentModerationService{}
	entry := svc.buildLog(ContentModerationCheckInput{noContentPersistence: true}, defaultContentModerationConfig(), ContentModerationActionAllow, false, "", 0, nil, "private prompt", nil, nil, "upstream echoed private prompt")
	require.Empty(t, entry.InputExcerpt)
	require.Empty(t, entry.Error)
}
