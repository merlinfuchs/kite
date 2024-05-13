package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.WorkspaceStore = (*Client)(nil)

func (c *Client) GetWorkspace(ctx context.Context, id string, appID string) (*model.Workspace, error) {
	workspace, err := c.Q.GetWorkspaceForApp(ctx, pgmodel.GetWorkspaceForAppParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	res, err := workspaceToModel(workspace)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) GetWorkspacesForApp(ctx context.Context, appID string) ([]model.Workspace, error) {
	workspaces, err := c.Q.GetWorkspacesForApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]model.Workspace, len(workspaces))
	for i, workspace := range workspaces {
		res[i], err = workspaceToModel(workspace)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, workspace model.Workspace) (*model.Workspace, error) {
	files, err := json.Marshal(workspace.Files)
	if err != nil {
		return nil, err
	}

	w, err := c.Q.CreateWorkspace(ctx, pgmodel.CreateWorkspaceParams{
		ID:          workspace.ID,
		AppID:       workspace.AppID,
		Type:        string(workspace.Type),
		Name:        workspace.Name,
		Description: workspace.Description,
		Files:       files,
		CreatedAt:   timeToTimestamp(workspace.CreatedAt),
		UpdatedAt:   timeToTimestamp(workspace.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}

	res, err := workspaceToModel(w)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, workspace model.Workspace) (*model.Workspace, error) {
	files, err := json.Marshal(workspace.Files)
	if err != nil {
		return nil, err
	}

	w, err := c.Q.UpdateWorkspace(ctx, pgmodel.UpdateWorkspaceParams{
		ID:          workspace.ID,
		AppID:       workspace.AppID,
		Name:        workspace.Name,
		Description: workspace.Description,
		Files:       files,
		UpdatedAt:   timeToTimestamp(workspace.UpdatedAt),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	res, err := workspaceToModel(w)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, id string, appID string) error {
	_, err := c.Q.DeleteWorkspace(ctx, pgmodel.DeleteWorkspaceParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func workspaceToModel(workspace pgmodel.Workspace) (model.Workspace, error) {
	files := []model.WorkspaceFile{}
	if err := json.Unmarshal(workspace.Files, &files); err != nil {
		return model.Workspace{}, err
	}

	return model.Workspace{
		ID:          workspace.ID,
		AppID:       workspace.AppID,
		Type:        model.WorkspaceType(workspace.Type),
		Name:        workspace.Name,
		Description: workspace.Description,
		Files:       files,
		CreatedAt:   workspace.CreatedAt.Time,
		UpdatedAt:   workspace.UpdatedAt.Time,
	}, nil
}
