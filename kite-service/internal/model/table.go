package model

import "time"

type TableColumn struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Required         bool   `json:"required,omitempty"`
	Default          any    `json:"default,omitempty"`
	ReferenceTableID string `json:"reference_table_id,omitempty"`
}

type Table struct {
	ID        string
	AppID     string
	Name      string
	Columns   []TableColumn
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TableRow struct {
	ID        string
	TableID   string
	Data      map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
