package provider

import (
	"context"
)

type TableRow struct {
	ID   string
	Data map[string]any
}

type TableColumn struct {
	ID               string
	Name             string
	Type             string
	Required         bool
	Default          any
	ReferenceTableID string
}

type TableFilter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
}

type TableProvider interface {
	InsertRow(context.Context, string, map[string]any) (*TableRow, error)
	GetRow(context.Context, string, string) (*TableRow, error)
	FindRows(context.Context, string, []TableFilter, int) ([]*TableRow, error)
	UpdateRow(context.Context, string, string, map[string]any) (*TableRow, error)
	DeleteRow(context.Context, string, string) error
	ListColumns(context.Context, string) ([]TableColumn, error)
	AddColumn(context.Context, string, TableColumn) (TableColumn, error)
	UpdateColumn(context.Context, string, string, TableColumn) (TableColumn, error)
	DeleteColumn(context.Context, string, string) error
}

type MockTableProvider struct{}

func (MockTableProvider) InsertRow(context.Context, string, map[string]any) (*TableRow, error) {
	return &TableRow{}, nil
}
func (MockTableProvider) GetRow(context.Context, string, string) (*TableRow, error) {
	return &TableRow{}, nil
}
func (MockTableProvider) FindRows(context.Context, string, []TableFilter, int) ([]*TableRow, error) {
	return nil, nil
}
func (MockTableProvider) UpdateRow(context.Context, string, string, map[string]any) (*TableRow, error) {
	return &TableRow{}, nil
}
func (MockTableProvider) DeleteRow(context.Context, string, string) error { return nil }
func (MockTableProvider) ListColumns(context.Context, string) ([]TableColumn, error) {
	return nil, nil
}
func (MockTableProvider) AddColumn(context.Context, string, TableColumn) (TableColumn, error) {
	return TableColumn{}, nil
}
func (MockTableProvider) UpdateColumn(context.Context, string, string, TableColumn) (TableColumn, error) {
	return TableColumn{}, nil
}
func (MockTableProvider) DeleteColumn(context.Context, string, string) error { return nil }
