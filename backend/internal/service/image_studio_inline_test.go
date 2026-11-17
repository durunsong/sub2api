package service

import (
	"context"
	"encoding/base64"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestImageStudioForcesInlineOutputWithoutChangingAccountSettings(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := b64BackfillAccount(false)
	encoded := base64.StdEncoding.EncodeToString(b64BackfillPNGBytes)
	body := []byte(`{"data":[{"url":"data:image/png;base64,` + encoded + `"}]}`)
	result := svc.backfillOpenAIImagesB64JSON(ctxkey.WithNoContentPersistence(context.Background()), account, &OpenAIImagesRequest{ResponseFormat: "b64_json"}, body)
	require.Equal(t, encoded, gjson.GetBytes(result, "data.0.b64_json").String())
	require.False(t, ImagesURLToB64JSONEnabled(account))
}
