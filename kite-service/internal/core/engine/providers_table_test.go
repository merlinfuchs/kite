package engine

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type tableStoreStub struct {
	table  *model.Table
	audits []*model.TableAudit
}

func (s *tableStoreStub) TablesByApp(context.Context, string) ([]*model.Table, error) {
	return []*model.Table{s.table}, nil
}
func (s *tableStoreStub) Table(context.Context, string, string) (*model.Table, error) {
	if s.table == nil {
		return nil, store.ErrNotFound
	}
	return s.table, nil
}
func (s *tableStoreStub) CreateTable(context.Context, *model.Table) (*model.Table, error) {
	return s.table, nil
}
func (s *tableStoreStub) UpdateTable(_ context.Context, table *model.Table) (*model.Table, error) {
	s.table = table
	return table, nil
}
func (s *tableStoreStub) DeleteTable(context.Context, string, string) error { return nil }
func (s *tableStoreStub) TableRows(context.Context, string, string) ([]*model.TableRow, error) {
	return nil, nil
}
func (s *tableStoreStub) CreateTableRow(_ context.Context, row *model.TableRow) (*model.TableRow, error) {
	return row, nil
}
func (s *tableStoreStub) UpdateTableRow(_ context.Context, row *model.TableRow) (*model.TableRow, error) {
	return row, nil
}
func (s *tableStoreStub) DeleteTableRow(context.Context, string, string) error { return nil }
func (s *tableStoreStub) CreateTableAudit(_ context.Context, audit *model.TableAudit) error {
	s.audits = append(s.audits, audit)
	return nil
}
func (s *tableStoreStub) TableAudits(context.Context, string, string, int) ([]*model.TableAudit, error) {
	return s.audits, nil
}
func (s *tableStoreStub) TableAudit(context.Context, string, string, string) (*model.TableAudit, error) {
	return nil, store.ErrNotFound
}
func (s *tableStoreStub) MarkTableAuditUndone(context.Context, string, string, string, time.Time) error {
	return nil
}

func TestTableProviderColumnMutationsCreateAudits(t *testing.T) {
	stub := &tableStoreStub{table: &model.Table{
		ID: "table", AppID: "app", Columns: []model.TableColumn{
			{ID: "name", Name: "Name", Type: "text"},
		},
	}}
	tableProvider := NewTableProvider("app", stub)

	if _, err := tableProvider.AddColumn(context.Background(), "table", provider.TableColumn{
		ID: "status", Name: "Status", Type: "text",
	}); err != nil {
		t.Fatal(err)
	}
	if len(stub.audits) != 1 || stub.audits[0].Action != "table_update" {
		t.Fatalf("expected one table audit, got %#v", stub.audits)
	}
	if string(stub.audits[0].BeforeData) == string(stub.audits[0].AfterData) {
		t.Fatal("column addition audit did not capture a change")
	}
}
