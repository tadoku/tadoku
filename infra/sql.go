package infra

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/srvc/fail"
	"github.com/tadoku/api/interfaces/rdb"

	// Postgres driver that's used to connect to the db
	_ "github.com/lib/pq"
)

// -----------------------------------------------------------------
// SQLHandler
// -----------------------------------------------------------------

// NewSQLHandler creates an interface to run queries on a database
func NewSQLHandler(db *RDB) rdb.SQLHandler {
	return &sqlHandler{db: db}
}

type sqlHandler struct {
	db *RDB
}

func (handler *sqlHandler) Execute(statement string, args ...interface{}) (rdb.Result, error) {
	res := sqlResult{}
	result, err := handler.db.Exec(statement, args...)
	if err != nil {
		return res, fail.Wrap(err)
	}
	res.Result = result

	return res, nil
}

func (handler *sqlHandler) NamedExecute(statement string, arg interface{}) (rdb.Result, error) {
	res := sqlResult{}
	result, err := handler.db.NamedExec(statement, arg)
	if err != nil {
		return res, fail.Wrap(err)
	}
	res.Result = result

	return res, nil
}

func (handler *sqlHandler) Query(statement string, args ...interface{}) (rdb.Rows, error) {
	row := new(sqlRows)
	rows, err := handler.db.Queryx(statement, args...)
	if err != nil {
		return row, fail.Wrap(err)
	}
	row.Rows = rows

	return row, nil
}

func (handler *sqlHandler) QueryRow(statement string, args ...interface{}) rdb.Row {
	return sqlRow{Row: handler.db.QueryRowx(statement, args...)}
}

func (handler *sqlHandler) Get(dest interface{}, query string, args ...interface{}) error {
	return handler.db.Get(dest, query, args...)
}

func (handler *sqlHandler) Select(dest interface{}, query string, args ...interface{}) error {
	return handler.db.Select(dest, query, args...)
}

func (handler *sqlHandler) Begin() (rdb.TxHandler, error) {
	tx, err := handler.db.Beginx()
	if err != nil {
		return nil, fail.Wrap(err)
	}

	return txHandler{tx: tx}, nil
}

// -----------------------------------------------------------------
// Transaction - Tx
// -----------------------------------------------------------------
type txHandler struct {
	tx *sqlx.Tx
}

func (handler txHandler) Query(statement string, args ...interface{}) (rdb.Rows, error) {
	row := new(sqlRows)
	rows, err := handler.tx.Queryx(statement, args...)
	if err != nil {
		return row, fail.Wrap(err)
	}
	row.Rows = rows

	return row, nil
}

func (handler txHandler) QueryRow(statement string, args ...interface{}) rdb.Row {
	return sqlRow{Row: handler.tx.QueryRowx(statement, args...)}
}

func (handler txHandler) Get(dest interface{}, query string, args ...interface{}) error {
	return handler.Get(dest, query, args...)
}

func (handler txHandler) Select(dest interface{}, query string, args ...interface{}) error {
	return handler.Select(dest, query, args...)
}

func (handler txHandler) Commit() error {
	err := handler.tx.Commit()
	if err != nil {
		return fail.Wrap(err)
	}

	return nil
}

func (handler txHandler) Rollback() error {
	err := handler.tx.Rollback()
	if err != nil {
		return fail.Wrap(err)
	}

	return nil
}

// -----------------------------------------------------------------
// Result
// -----------------------------------------------------------------

type sqlResult struct {
	Result sql.Result
}

func (r sqlResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r sqlResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

// -----------------------------------------------------------------
// Rows
// -----------------------------------------------------------------

type sqlRows struct {
	Rows *sqlx.Rows
}

func (r sqlRows) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r sqlRows) StructScan(dest interface{}) error {
	return r.Rows.StructScan(dest)
}

func (r sqlRows) Next() bool {
	return r.Rows.Next()
}

func (r sqlRows) Close() error {
	return r.Rows.Close()
}

// -----------------------------------------------------------------
// Row
// -----------------------------------------------------------------

type sqlRow struct {
	Row *sqlx.Row
}

func (r sqlRow) Scan(dest ...interface{}) error {
	return r.Row.Scan(dest...)
}

func (r sqlRow) StructScan(dest interface{}) error {
	return r.Row.StructScan(dest)
}
