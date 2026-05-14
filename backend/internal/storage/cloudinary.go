package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-viper/mapstructure/v2"
)

type CloudinaryProvider struct {
	cld *cloudinary.Cloudinary
}

// Matches the actual Cloudinary API response embedded in Response interface{}
type cloudinaryRawResponse struct {
	Plan        string `json:"plan"`
	LastUpdated string `json:"last_updated"`
	Credits     struct {
		Usage       float64 `json:"usage"`
		Limit       float64 `json:"limit"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"credits"`

	Bandwidth struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
	} `json:"bandwidth"`

	Storage struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
	} `json:"storage"`

	Transformations struct {
		Usage        int     `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
		Breakdown    struct {
			Transformation int `json:"transformation"`
		} `json:"breakdown"`
	} `json:"transformations"`

	Objects struct {
		Usage int `json:"usage"`
	} `json:"objects"`

	Impressions struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
	} `json:"impressions"`

	SecondsDelivered struct {
		Usage        int64   `json:"usage"`
		CreditsUsage float64 `json:"credits_usage"`
	} `json:"seconds_delivered"`

	MediaLimits struct {
		ImageMaxSizeBytes int `json:"image_max_size_bytes"`
		VideoMaxSizeBytes int `json:"video_max_size_bytes"`
		RawMaxSizeBytes   int `json:"raw_max_size_bytes"`
		ImageMaxPx        int `json:"image_max_px"`
		AssetMaxTotalPx   int `json:"asset_max_total_px"`
	} `json:"media_limits"`
}

type UsageResponse struct {
	Plan        string    `json:"plan"`
	LastUpdated string    `json:"last_updated"`
	DateFetched time.Time `json:"date_fetched"`

	Credits struct {
		Used        float64 `json:"used"`
		Limit       float64 `json:"limit"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"credits"`

	Storage struct {
		UsageBytes  int64   `json:"usage_bytes"`
		UsageMB     float64 `json:"usage_mb"`
		CreditsUsed float64 `json:"credits_used"`
	} `json:"storage"`

	Bandwidth struct {
		UsageBytes  int64   `json:"usage_bytes"`
		UsageGB     float64 `json:"usage_gb"`
		CreditsUsed float64 `json:"credits_used"`
	} `json:"bandwidth"`
}

func NewCloudinaryProvider() (*CloudinaryProvider, error) {
	name := os.Getenv("CLOUDINARY_CLOUD_NAME")
	api := os.Getenv("CLOUDINARY_API_KEY")
	secret := os.Getenv("CLOUDINARY_API_SECRET")

	if name == "" || api == "" || secret == "" {
		return nil, fmt.Errorf("missing cloudinary credentials")
	}

	cld, err := cloudinary.NewFromParams(name, api, secret)
	if err != nil {
		return nil, err
	}

	return &CloudinaryProvider{cld: cld}, nil
}

func (c *CloudinaryProvider) Name() string {
	return "cloudinary"
}

func (c *CloudinaryProvider) Upload(
	ctx context.Context,
	key string,
	data io.Reader,
	contentType string,
	size int64,
) (string, error) {

	resourceType := "raw"
	if strings.HasPrefix(contentType, "image/") {
		resourceType = "image"
	}

	result, err := c.cld.Upload.Upload(
		ctx,
		data,
		uploader.UploadParams{
			PublicID:     key,
			ResourceType: resourceType,
		},
	)

	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}

func (c *CloudinaryProvider) Delete(ctx context.Context, key, mimeType string) error {
	if strings.HasPrefix(mimeType, "image/") {
		_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
			PublicID:     key,
			ResourceType: "image",
		})
		return err
	}

	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     key,
		ResourceType: "raw",
	})

	return err
}

func (c *CloudinaryProvider) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *CloudinaryProvider) Ping(ctx context.Context) error {
	_, err := c.cld.Admin.Ping(ctx)
	return err
}

func (c *CloudinaryProvider) Usage(ctx context.Context) (any, error) {
	resp, err := c.cld.Admin.Usage(ctx, admin.UsageParams{})
	if err != nil {
		return nil, err
	}

	var raw cloudinaryRawResponse
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &raw,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}
	if err := decoder.Decode(resp.Response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &UsageResponse{
		Plan:        raw.Plan,
		LastUpdated: raw.LastUpdated,
	}

	result.Credits.Used = raw.Credits.Usage
	result.Credits.Limit = raw.Credits.Limit
	result.Credits.UsedPercent = raw.Credits.UsedPercent

	result.Storage.UsageBytes = raw.Storage.Usage
	result.Storage.UsageMB = float64(raw.Storage.Usage) / (1024 * 1024)
	result.Storage.CreditsUsed = raw.Storage.CreditsUsage

	result.Bandwidth.UsageBytes = raw.Bandwidth.Usage
	result.Bandwidth.UsageGB = float64(raw.Bandwidth.Usage) / (1024 * 1024 * 1024)
	result.Bandwidth.CreditsUsed = raw.Bandwidth.CreditsUsage

	return result, nil
}
