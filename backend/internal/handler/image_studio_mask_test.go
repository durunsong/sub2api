package handler

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageStudioNativeMaskPreservesSeparateImages(t *testing.T) {
	original := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	mask := image.NewNRGBA(original.Bounds())
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			original.SetNRGBA(x, y, color.NRGBA{R: 200, A: 255})
			mask.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	mask.SetNRGBA(1, 1, color.NRGBA{})
	var originalPNG, maskPNG bytes.Buffer
	require.NoError(t, png.Encode(&originalPNG, original))
	require.NoError(t, png.Encode(&maskPNG, mask))
	dataURL := func(data []byte) string { return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data) }
	model := imageStudioModel{Model: "gpt-image-2", Protocol: "images", MaxImages: 1, MaxReferences: 3, SupportsMask: true}
	input := imageStudioGenerateRequest{Prompt: "edit", References: []string{dataURL(originalPNG.Bytes())}, Mask: dataURL(maskPNG.Bytes())}
	body, path, contentType, err := buildImageStudioRequest(model, input)
	require.NoError(t, err)
	require.Equal(t, "/v1/images/edits", path)
	mediaType, params, err := mime.ParseMediaType(contentType)
	require.NoError(t, err)
	require.Equal(t, "multipart/form-data", mediaType)
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	parts := map[string][]byte{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		data, err := io.ReadAll(part)
		require.NoError(t, err)
		parts[part.FormName()] = data
	}
	require.Equal(t, originalPNG.Bytes(), parts["image[]"])
	require.Equal(t, maskPNG.Bytes(), parts["mask"])
	require.Equal(t, []byte("1"), parts["n"])
	model.SupportsMask = false
	_, _, _, err = buildImageStudioRequest(model, input)
	require.ErrorContains(t, err, "mask requires a supported model")
}
