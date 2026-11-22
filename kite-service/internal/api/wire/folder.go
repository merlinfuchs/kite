package wire

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gopkg.in/guregu/null.v4"
)

// FolderUpdateRequest moves commands, event listeners or messages into a
// folder, or out of any with a null or empty folder. Renaming a folder moves
// all of its items.
type FolderUpdateRequest struct {
	IDs    []string    `json:"ids"`
	Folder null.String `json:"folder"`
}

func (req FolderUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.IDs, validation.Required, validation.Length(1, 500), validation.Each(validation.Required)),
		validation.Field(&req.Folder, validation.Length(0, 100)),
	)
}

// NormalizedFolder trims the folder, a blank one is no folder.
func (req FolderUpdateRequest) NormalizedFolder() null.String {
	folder := strings.TrimSpace(req.Folder.String)
	return null.NewString(folder, req.Folder.Valid && folder != "")
}
