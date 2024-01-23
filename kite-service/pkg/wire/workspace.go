package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type Workspace struct {
	ID          string          `json:"id"`
	GuildID     string          `json:"guild_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Files       []WorkspaceFile `json:"files"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type WorkspaceFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WorkspaceGetResponse APIResponse[Workspace]

type WorkspaceListResponse APIResponse[[]Workspace]

type WorkspaceCreateRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Files       []WorkspaceFile `json:"files"`
}

type WorkspaceCreateResponse APIResponse[Workspace]

type WorkspaceUpdateRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Files       []WorkspaceFile `json:"files"`
}

type WorkspaceUpdateResponse APIResponse[Workspace]

type WorkspaceDeleteResponse APIResponse[struct{}]

func WorkspaceToWire(workspace *model.Workspace) Workspace {
	files := make([]WorkspaceFile, len(workspace.Files))
	for i, file := range workspace.Files {
		files[i] = WorkspaceFile{
			Path:    file.Path,
			Content: file.Content,
		}
	}

	return Workspace{
		ID:          workspace.ID,
		GuildID:     workspace.GuildID,
		Name:        workspace.Name,
		Description: workspace.Description,
		Files:       files,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}
