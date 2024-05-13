package access

import (
	"context"
	"fmt"

	"github.com/merlinfuchs/dismod/distype"
)

type AppPermissions struct {
	UserIsOwner bool
}

func (m *AccessManager) GetAppPermissionsForUser(ctx context.Context, appID distype.Snowflake, userID distype.Snowflake) (*AppPermissions, error) {
	isOwner, err := m.apps.CheckUserIsOwnerOfApp(ctx, appID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user is owner of app: %w", err)
	}

	return &AppPermissions{
		UserIsOwner: isOwner,
	}, nil
}
