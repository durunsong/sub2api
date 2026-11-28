package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Muse-compatible IDs must never trigger GPT Responses image-tool conversion.
func IsMuseImageModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "muse-image", "muse-image-1.0":
		return true
	default:
		return false
	}
}

func validateMuseImagesRequest(req *OpenAIImagesRequest) error {
	if req.N != 1 || req.Stream || req.HasMask || req.HasNativeOptions {
		return fmt.Errorf("Muse images require n=1 and do not support streaming, masks or native GPT options")
	}
	if req.ResponseFormat != "" && req.ResponseFormat != "url" && req.ResponseFormat != "b64_json" {
		return fmt.Errorf("unsupported Muse response_format")
	}
	allowed := map[string]bool{"model": true, "prompt": true, "n": true, "size": true, "response_format": true}
	if req.IsEdits() {
		allowed["images"] = true
		allowed["image"] = req.Multipart
		allowed["image[]"] = req.Multipart
	}
	if req.Multipart {
		for i := range req.Uploads {
			if !isBackfillImageContent(req.Uploads[i].Data) {
				return fmt.Errorf("Muse multipart references must contain image bytes")
			}
		}
		seen := map[string]bool{}
		_, params, err := mime.ParseMediaType(req.ContentType)
		if err != nil {
			return err
		}
		reader := multipart.NewReader(bytes.NewReader(req.Body), params["boundary"])
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			name := part.FormName()
			isFile := part.FileName() != ""
			_ = part.Close()
			if !allowed[name] {
				return fmt.Errorf("unsupported Muse image field: %s", name)
			}
			if (name == "image" || name == "image[]") != isFile || name == "images" {
				return fmt.Errorf("Muse multipart references must be uploaded image files")
			}
			if !isFile && seen[name] {
				return fmt.Errorf("duplicate Muse image field")
			}
			seen[name] = true
		}
	} else {
		var unsupported string
		seen := map[string]bool{}
		gjson.ParseBytes(req.Body).ForEach(func(k, v gjson.Result) bool {
			if !allowed[k.String()] || seen[k.String()] {
				unsupported = k.String()
				return false
			}
			seen[k.String()] = true
			return true
		})
		if unsupported != "" {
			return fmt.Errorf("unsupported Muse image field: %s", unsupported)
		}
		if req.IsEdits() {
			for _, image := range gjson.GetBytes(req.Body, "images").Array() {
				if !image.IsObject() || image.Get("image_url").Type != gjson.String || strings.TrimSpace(image.Get("image_url").String()) == "" {
					return fmt.Errorf("Muse edits require images[].image_url")
				}
				if err := ValidateMuseImageReference(image.Get("image_url").String()); err != nil {
					return err
				}
				valid := true
				fields := 0
				image.ForEach(func(k, v gjson.Result) bool { fields++; valid = k.String() == "image_url" && fields == 1; return valid })
				if !valid {
					return fmt.Errorf("unsupported Muse image reference field")
				}
			}
		}
	}
	return nil
}

func (s *OpenAIGatewayService) normalizeMuseImageResponse(ctx context.Context, account *Account, body []byte) ([]byte, error) {
	items := gjson.GetBytes(body, "data")
	if !items.IsArray() || len(items.Array()) != 1 {
		return nil, fmt.Errorf("Muse returned no single image output")
	}
	item := items.Array()[0]
	encoded := item.Get("b64_json").String()
	if encoded == "" {
		var err error
		encoded, err = s.fetchOpenAIImageURLBase64(WithHTTPUpstreamRedirectsDisabled(ctx), account, item.Get("url").String())
		if err != nil {
			return nil, fmt.Errorf("Muse output dimensions could not be verified")
		}
		body, err = sjson.SetBytes(body, "data.0.b64_json", encoded)
		if err != nil {
			return nil, err
		}
	}
	actual := detectOpenAIImageResultSize(encoded)
	if actual == "" {
		return nil, fmt.Errorf("Muse output dimensions could not be verified")
	}
	// Muse2API echoes request.size; only bytes determine the billable output tier.
	return sjson.SetBytes(body, "data.0.size", actual)
}

// Output dimensions are provider-controlled, so all output tiers need an
// explicit price before creating a paid job (zero is an intentional free tier).
func ValidateMuseImagePricing(group *Group) error {
	if group == nil || !group.ImageCountBilling {
		return fmt.Errorf("Muse image price requires image-count billing")
	}
	for _, tier := range []string{"1K", "2K", "4K"} {
		price := group.GetImagePrice(tier)
		if price == nil || math.IsNaN(*price) || math.IsInf(*price, 0) || *price < 0 {
			return fmt.Errorf("Muse image price must be configured for %s", tier)
		}
	}
	return nil
}
