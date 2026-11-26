package service

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
)

func TestVideoStudioLastFrameIncludedInSafetyInputs(t *testing.T) {
	info := ParseGrokVideoRequest("application/json", []byte(`{"model":"grok-imagine-video","image":{"url":"first"},"last_frame":{"url":"last"}}`))
	require.Equal(t, []string{"first", "last"}, info.InputImageURLs)
}

func TestVideoStudioChunkedContentSignalsReadFailure(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/content", nil).WithContext(ctxkey.WithNoContentPersistence(context.Background()))
	resp := &http.Response{StatusCode: 200, Header: make(http.Header), ContentLength: -1, Body: io.NopCloser(io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(errors.New("broken stream"))))}
	require.Error(t, writeGrokMediaContentResponse(c, resp))
	require.Equal(t, "true", c.Writer.Header().Get("X-Sub2API-Content-Read-Error"))
}

func TestVideoStudioDoesNotChangeImageReferenceParsing(t *testing.T) {
	info := ParseGrokMediaRequest("application/json", []byte(`{"model":"grok-imagine-image","image":{"url":"first"},"last_frame":{"url":"last"}}`))
	require.Equal(t, []string{"first"}, info.InputImageURLs)
}
