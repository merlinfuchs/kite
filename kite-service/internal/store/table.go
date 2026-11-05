package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type TableStore interface {
	TablesByApp(context.Context, string) ([]*model.Table, error)
	Table(context.Context, string, string) (*model.Table, error)
	CreateTable(context.Context, *model.Table) (*model.Table, error)
	UpdateTable(context.Context, *model.Table) (*model.Table, error)
	DeleteTable(context.Context, string, string) error
	TableRows(context.Context, string, string) ([]*model.TableRow, error)
	CreateTableRow(context.Context, *model.TableRow) (*model.TableRow, error)
	UpdateTableRow(context.Context, *model.TableRow) (*model.TableRow, error)
	DeleteTableRow(context.Context, string, string) error
	CreateTableAudit(context.Context, *model.TableAudit) error
	TableAudits(context.Context, string, string, int) ([]*model.TableAudit, error)
	TableAudit(context.Context, string, string, string) (*model.TableAudit, error)
	MarkTableAuditUndone(context.Context, string, string, string, time.Time) error
}
