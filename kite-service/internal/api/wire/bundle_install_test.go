package wire

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleInstallCreateRequestValidate(t *testing.T) {
	valid := func() BundleInstallCreateRequest {
		return BundleInstallCreateRequest{
			BundleID:      "moderation",
			BundleVersion: 1,
			Settings:      map[string]string{"channel_id": "123"},
			Items:         []string{"ban", "kick"},
		}
	}

	tests := []struct {
		name    string
		change  func(r *BundleInstallCreateRequest)
		wantErr bool
	}{
		{"valid", func(r *BundleInstallCreateRequest) {}, false},
		{"no settings", func(r *BundleInstallCreateRequest) { r.Settings = nil }, false},
		{"no items", func(r *BundleInstallCreateRequest) { r.Items = nil }, true},
		{"duplicate item", func(r *BundleInstallCreateRequest) { r.Items[1] = "ban" }, true},
		{"empty item", func(r *BundleInstallCreateRequest) { r.Items[0] = "" }, true},
		{"missing bundle", func(r *BundleInstallCreateRequest) { r.BundleID = "" }, true},
		{"missing version", func(r *BundleInstallCreateRequest) { r.BundleVersion = 0 }, true},
		{"long setting", func(r *BundleInstallCreateRequest) { r.Settings["channel_id"] = strings.Repeat("a", 4001) }, true},
		{"too many settings", func(r *BundleInstallCreateRequest) {
			for i := range 51 {
				r.Settings[strings.Repeat("k", i+1)] = "v"
			}
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid()
			tt.change(&req)
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
