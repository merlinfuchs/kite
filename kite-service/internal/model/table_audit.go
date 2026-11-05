package model

import "time"

type TableAudit struct {
	ID          string
	AppID       string
	TableID     string
	ActorUserID string
	Action      string
	RowID       string
	BeforeData  []byte
	AfterData   []byte
	CreatedAt   time.Time
	UndoneAt    *time.Time
}
