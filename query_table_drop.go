package bun

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun/internal"
	"github.com/uptrace/bun/schema"
)

type DropTableQuery struct {
	baseQuery
	cascadeQuery

	ifExists bool
}

func NewDropTableQuery(db *DB) *DropTableQuery {
	q := &DropTableQuery{
		baseQuery: baseQuery{
			db:  db,
			dbi: db.DB,
		},
	}
	return q
}

func (q *DropTableQuery) Conn(db DBI) *DropTableQuery {
	q.setDBI(db)
	return q
}

func (q *DropTableQuery) Model(model interface{}) *DropTableQuery {
	q.setTableModel(model)
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) Table(tables ...string) *DropTableQuery {
	for _, table := range tables {
		q.addTable(schema.UnsafeIdent(table))
	}
	return q
}

func (q *DropTableQuery) TableExpr(query string, args ...interface{}) *DropTableQuery {
	q.addTable(schema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) IfExists() *DropTableQuery {
	q.ifExists = true
	return q
}

func (q *DropTableQuery) Restrict() *DropTableQuery {
	q.restrict = true
	return q
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	b = append(b, "DROP TABLE "...)
	if q.ifExists {
		b = append(b, "IF EXISTS "...)
	}

	b, err = q.appendTables(fmter, b)
	if err != nil {
		return nil, err
	}

	b = q.appendCascade(fmter, b)

	return b, nil
}

//------------------------------------------------------------------------------

func (q *DropTableQuery) Exec(ctx context.Context, dest ...interface{}) (res sql.Result, _ error) {
	if q.table != nil {
		if err := q.beforeDropTableHook(ctx); err != nil {
			return res, err
		}
	}

	bs := getByteSlice()
	defer putByteSlice(bs)

	queryBytes, err := q.AppendQuery(q.db.fmter, bs.b)
	if err != nil {
		return res, err
	}

	bs.update(queryBytes)
	query := internal.String(queryBytes)

	res, err = q.exec(ctx, q, query)
	if err != nil {
		return res, err
	}

	if q.table != nil {
		if err := q.afterDropTableHook(ctx); err != nil {
			return res, err
		}
	}

	return res, nil
}

func (q *DropTableQuery) beforeDropTableHook(ctx context.Context) error {
	if hook, ok := q.table.ZeroIface.(BeforeDropTableHook); ok {
		if err := hook.BeforeDropTable(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (q *DropTableQuery) afterDropTableHook(ctx context.Context) error {
	if hook, ok := q.table.ZeroIface.(AfterDropTableHook); ok {
		if err := hook.AfterDropTable(ctx, q); err != nil {
			return err
		}
	}
	return nil
}
