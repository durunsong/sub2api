package service

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestImageStudioGrokPresetsBecomeNativeGeometry(t *testing.T) {
	for _, tc := range []struct{ size, resolution, aspect string }{{"1024x1024", "1k", "1:1"}, {"1536x1024", "2k", "3:2"}, {"1024x1536", "2k", "2:3"}} {
		body, err := applyGrokImagineImageGeometry([]byte(`{"model":"grok-imagine-image","prompt":"draw","size":"` + tc.size + `"}`))
		require.NoError(t, err)
		require.False(t, gjson.GetBytes(body, "size").Exists())
		require.Equal(t, tc.resolution, gjson.GetBytes(body, "resolution").String())
		require.Equal(t, tc.aspect, gjson.GetBytes(body, "aspect_ratio").String())
	}
}
