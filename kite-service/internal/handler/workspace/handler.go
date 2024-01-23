package workspace

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/util"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type WorkspaceHandler struct {
	workspaces store.WorkspaceStore
}

func NewHandler(workspaces store.WorkspaceStore) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaces: workspaces,
	}
}

func (h *WorkspaceHandler) HandleWorkspaceCreate(c *fiber.Ctx, req wire.WorkspaceCreateRequest) error {
	files := make([]model.WorkspaceFile, len(req.Files))
	for i, file := range req.Files {
		files[i] = model.WorkspaceFile{
			Path:    file.Path,
			Content: file.Content,
		}
	}

	workspace, err := h.workspaces.CreateWorkspace(c.Context(), model.Workspace{
		ID:          util.UniqueID(),
		GuildID:     c.Params("guildID"),
		Name:        req.Name,
		Description: req.Description,
		Files:       files,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return c.JSON(wire.WorkspaceCreateResponse{
		Success: true,
		Data:    wire.WorkspaceToWire(workspace),
	})
}

func (h *WorkspaceHandler) HandleWorkspaceUpdate(c *fiber.Ctx, req wire.WorkspaceUpdateRequest) error {
	files := make([]model.WorkspaceFile, len(req.Files))
	for i, file := range req.Files {
		files[i] = model.WorkspaceFile{
			Path:    file.Path,
			Content: file.Content,
		}
	}

	workspace, err := h.workspaces.UpdateWorkspace(c.Context(), model.Workspace{
		ID:          c.Params("workspaceID"),
		GuildID:     c.Params("guildID"),
		Name:        req.Name,
		Description: req.Description,
		Files:       files,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_workspace", "Workspace not found")
		}
		return err
	}

	return c.JSON(wire.WorkspaceCreateResponse{
		Success: true,
		Data:    wire.WorkspaceToWire(workspace),
	})
}

func (h *WorkspaceHandler) HandleWorkspaceGetForGuild(c *fiber.Ctx) error {
	workspace, err := h.workspaces.GetWorkspace(c.Context(), c.Params("workspaceID"), c.Params("guildID"))
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_workspace", "Workspace not found")
		}
		return err
	}

	return c.JSON(wire.WorkspaceGetResponse{
		Success: true,
		Data:    wire.WorkspaceToWire(workspace),
	})
}

func (h *WorkspaceHandler) HandleWorkspaceListForGuild(c *fiber.Ctx) error {
	workspaces, err := h.workspaces.GetWorkspacesForGuild(c.Context(), c.Params("guildID"))
	if err != nil {
		return err
	}

	res := make([]wire.Workspace, len(workspaces))
	for i, workspace := range workspaces {
		res[i] = wire.WorkspaceToWire(&workspace)
	}

	return c.JSON(wire.WorkspaceListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *WorkspaceHandler) HandleWorkspaceDelete(c *fiber.Ctx) error {
	err := h.workspaces.DeleteWorkspace(c.Context(), c.Params("workspaceID"), c.Params("guildID"))
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_workspace", "Workspace not found")
		}
		return err
	}

	return c.JSON(wire.WorkspaceDeleteResponse{
		Success: true,
	})
}
