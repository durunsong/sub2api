package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
)

// Three 10 MiB references expand to 40 MiB base64, before prompt/JSON overhead.
const imageStudioMaxBody = 42 << 20
const imageStudioMaxImage = 10 << 20
const imageStudioMaxResponse = 80 << 20

type imageStudioGenerateRequest struct {
	ModelID     string   `json:"model_id"`
	Prompt      string   `json:"prompt"`
	Size        string   `json:"size"`
	Quality     string   `json:"quality"`
	N           int      `json:"n"`
	AspectRatio string   `json:"aspect_ratio"`
	Resolution  string   `json:"resolution"`
	References  []string `json:"references"`
	Mask        string   `json:"mask"`
}
type imageStudioOutput struct {
	Images []imageStudioImage `json:"images"`
	Text   string             `json:"text,omitempty"`
}
type imageStudioImage struct {
	B64JSON  string `json:"b64_json"`
	MimeType string `json:"mime_type"`
}

func decodeStudioImage(raw string) ([]byte, string, error) {
	if len(raw) > base64.StdEncoding.EncodedLen(imageStudioMaxImage)+64 {
		return nil, "", errors.New("reference image is too large")
	}
	prefix, data, ok := strings.Cut(raw, ",")
	if !ok || !strings.HasPrefix(prefix, "data:image/") || !strings.HasSuffix(prefix, ";base64") {
		return nil, "", errors.New("reference must be a base64 image data URL")
	}
	mime := strings.TrimSuffix(strings.TrimPrefix(prefix, "data:"), ";base64")
	if mime != "image/png" && mime != "image/jpeg" && mime != "image/webp" {
		return nil, "", errors.New("only PNG, JPEG and WebP images are supported")
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(data)
	if err != nil || len(decoded) == 0 || len(decoded) > imageStudioMaxImage {
		return nil, "", errors.New("invalid image encoding")
	}
	if http.DetectContentType(decoded) != mime {
		return nil, "", errors.New("image content does not match its MIME type")
	}
	return decoded, mime, nil
}

// GPT Image 2 accepts dimensions divisible by 16, within the documented
// pixel budget and 3:1 aspect ratio. Bound each dimension before multiplication.
func validStudioCustomSize(size string) bool {
	width, height, ok := strings.Cut(size, "x")
	if !ok {
		return false
	}
	w, err := strconv.Atoi(width)
	if err != nil || w <= 0 || w > 3840 || w%16 != 0 {
		return false
	}
	h, err := strconv.Atoi(height)
	if err != nil || h <= 0 || h > 3840 || h%16 != 0 {
		return false
	}
	return w <= h*3 && h <= w*3 && w*h >= 655360 && w*h <= 8294400
}

func buildImageStudioRequest(model imageStudioModel, input imageStudioGenerateRequest) ([]byte, string, string, error) {
	fail := func(msg string) ([]byte, string, string, error) { return nil, "", "", errors.New(msg) }
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" || len([]rune(input.Prompt)) > 16000 {
		return fail("prompt must contain 1 to 16000 characters")
	}
	if input.N == 0 {
		input.N = 1
	}
	if input.N < 1 || input.N > model.MaxImages {
		return fail("unsupported image count")
	}
	if len(input.References) > model.MaxReferences {
		return fail("too many reference images")
	}
	if input.Mask != "" && (!model.SupportsMask || len(input.References) == 0) {
		return fail("mask requires a supported model and a reference image")
	}
	if input.Quality != "" && !slices.Contains(model.Qualities, input.Quality) {
		return fail("unsupported image quality")
	}
	if input.Size != "" && !slices.Contains(model.Sizes, input.Size) && !(model.SupportsCustomSize && validStudioCustomSize(input.Size)) {
		return fail("unsupported image size")
	}
	refs := make([][]byte, 0, len(input.References))
	mimes := make([]string, 0, len(input.References))
	for _, raw := range input.References {
		data, mime, err := decodeStudioImage(raw)
		if err != nil {
			return nil, "", "", err
		}
		refs = append(refs, data)
		mimes = append(mimes, mime)
	}
	if model.Protocol == "gemini" {
		resolution := input.Resolution
		if resolution == "" {
			resolution = input.Size
		}
		if resolution != "" && !slices.Contains(model.Sizes, resolution) {
			return fail("unsupported image resolution")
		}
		if input.AspectRatio != "" && !slices.Contains([]string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}, input.AspectRatio) {
			return fail("unsupported aspect ratio")
		}
		parts := []any{map[string]any{"text": input.Prompt}}
		for i, data := range refs {
			parts = append(parts, map[string]any{"inlineData": map[string]string{"mimeType": mimes[i], "data": base64.StdEncoding.EncodeToString(data)}})
		}
		generation := map[string]any{"responseModalities": []string{"TEXT", "IMAGE"}}
		imageConfig := map[string]string{}
		// 1K is the provider default; omit imageSize for legacy fixed-size models.
		if resolution != "" && resolution != "1K" {
			imageConfig["imageSize"] = resolution
		}
		if input.AspectRatio != "" {
			imageConfig["aspectRatio"] = input.AspectRatio
		}
		if len(imageConfig) > 0 {
			generation["imageConfig"] = imageConfig
		}
		payload := map[string]any{"contents": []any{map[string]any{"role": "user", "parts": parts}}, "generationConfig": generation}
		body, err := json.Marshal(payload)
		return body, "/v1beta/models/" + model.Model + ":generateContent", "application/json", err
	}
	if input.Resolution != "" || input.AspectRatio != "" {
		return fail("this model uses size rather than resolution or aspect ratio")
	}
	path := "/v1/images/generations"
	if len(refs) == 0 {
		payload := map[string]any{"model": model.Model, "prompt": input.Prompt, "n": input.N, "response_format": "b64_json"}
		if input.Size != "" {
			payload["size"] = input.Size
		}
		if input.Quality != "" {
			payload["quality"] = input.Quality
		}
		body, err := json.Marshal(payload)
		return body, path, "application/json", err
	}
	path = "/v1/images/edits"
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for _, field := range [][2]string{{"model", model.Model}, {"prompt", input.Prompt}, {"n", fmt.Sprint(input.N)}, {"response_format", "b64_json"}, {"size", input.Size}, {"quality", input.Quality}} {
		if field[1] != "" {
			if err := writer.WriteField(field[0], field[1]); err != nil {
				return nil, "", "", err
			}
		}
	}
	writeImage := func(name string, data []byte, mime string) error {
		ext := "png"
		if mime == "image/jpeg" {
			ext = "jpg"
		}
		if mime == "image/webp" {
			ext = "webp"
		}
		hdr := textproto.MIMEHeader{}
		hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="reference.%s"`, name, ext))
		hdr.Set("Content-Type", mime)
		part, err := writer.CreatePart(hdr)
		if err != nil {
			return err
		}
		_, err = part.Write(data)
		return err
	}
	for i, data := range refs {
		if err := writeImage("image[]", data, mimes[i]); err != nil {
			return nil, "", "", err
		}
	}
	if input.Mask != "" {
		data, mime, err := decodeStudioImage(input.Mask)
		if err != nil {
			return nil, "", "", err
		}
		if mime != "image/png" {
			return fail("mask must be a PNG image")
		}
		if err = writeImage("mask", data, mime); err != nil {
			return nil, "", "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), path, writer.FormDataContentType(), nil
}

func normalizeImageStudioResponse(body []byte) (imageStudioOutput, error) {
	var payload struct {
		Data []struct {
			B64 string `json:"b64_json"`
			URL string `json:"url"`
		} `json:"data"`
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text       string `json:"text"`
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	out := imageStudioOutput{Images: []imageStudioImage{}}
	if err := json.Unmarshal(body, &payload); err != nil {
		return out, errors.New("invalid image response")
	}
	appendImage := func(encoded, mime string) error {
		data, err := base64.StdEncoding.Strict().DecodeString(encoded)
		if err != nil || len(data) == 0 {
			return errors.New("invalid generated image")
		}
		detected := http.DetectContentType(data)
		if detected != "image/png" && detected != "image/jpeg" && detected != "image/webp" {
			return errors.New("unsupported generated image type")
		}
		if mime != "" && mime != detected {
			return errors.New("generated image MIME mismatch")
		}
		out.Images = append(out.Images, imageStudioImage{B64JSON: encoded, MimeType: detected})
		return nil
	}
	for _, item := range payload.Data {
		if item.B64 == "" {
			return out, errors.New("provider did not return inline images; generation may already be billed")
		}
		if err := appendImage(item.B64, ""); err != nil {
			return out, err
		}
	}
	for _, candidate := range payload.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil {
				if err := appendImage(part.InlineData.Data, part.InlineData.MimeType); err != nil {
					return out, err
				}
			}
			if part.Text != "" {
				if out.Text != "" {
					out.Text += "\n"
				}
				out.Text += part.Text
			}
		}
	}
	if len(out.Images) == 0 {
		return out, errors.New("provider returned no inline images; generation may already be billed")
	}
	return out, nil
}

type imageStudioResponseWriter struct {
	header   http.Header
	body     bytes.Buffer
	status   int
	overflow bool
}

func (w *imageStudioResponseWriter) Header() http.Header { return w.header }
func (w *imageStudioResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *imageStudioResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	if w.body.Len()+len(p) > imageStudioMaxResponse {
		w.overflow = true
		return 0, io.ErrShortBuffer
	}
	return w.body.Write(p)
}
func (w *imageStudioResponseWriter) Flush() {}
