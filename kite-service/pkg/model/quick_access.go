package model

import "time"

type QuickAccessItemType string

const (
	QuickAccessItemTypeDeployment QuickAccessItemType = "DEPLOYMENT"
	QuickAccessItemTypeWorkspace  QuickAccessItemType = "WORKSPACE"
)

type QuickAccessItem struct {
	ID        string
	AppID     string
	Type      QuickAccessItemType
	Name      string
	UpdatedAt time.Time
}
