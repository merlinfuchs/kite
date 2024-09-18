package wire

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type Asset struct {
	ID            string      `json:"id"`
	AppID         string      `json:"app_id"`
	ModuleID      null.String `json:"module_id"`
	CreatorUserID string      `json:"creator_user_id"`
	URL           string      `json:"url"`
	Name          string      `json:"name"`
	ContentType   string      `json:"content_type"`
	ContentHash   string      `json:"content_hash"`
	ContentSize   int         `json:"content_size"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	ExpiresAt     null.Time   `json:"expires_at"`
}

type AssetCreateResponse = Asset

type AssetGetResponse = Asset

func AssetToWire(asset *model.Asset, apiPublicBaseURL string) *Asset {
	if asset == nil {
		return nil
	}

	return &Asset{
		ID:            asset.ID,
		AppID:         asset.AppID,
		ModuleID:      asset.ModuleID,
		CreatorUserID: asset.CreatorUserID,
		URL:           fmt.Sprintf("%s/v1/assets/%s", apiPublicBaseURL, asset.ID),
		Name:          asset.Name,
		ContentType:   asset.ContentType,
		ContentHash:   asset.ContentHash,
		ContentSize:   asset.ContentSize,
		CreatedAt:     asset.CreatedAt,
		UpdatedAt:     asset.UpdatedAt,
		ExpiresAt:     asset.ExpiresAt,
	}
}
