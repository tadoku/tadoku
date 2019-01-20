package infra

import (
	"database/sql"

	"github.com/tadoku/api/interfaces/rdb"

	"github.com/jmoiron/sqlx"
	// Postgres driver that's used to connect to the db
	_ "github.com/lib/pq"
	"github.com/srvc/fail"
)

// NewRDB creates a new relational database connection pool
func NewRDB(URL string, maxIdleConns, maxOpenConns int) (rdb.SQLHandler, error) {
	db, err := sqlx.Open("postgres", URL)
	if err != nil {
		return nil, fail.Wrap(err)
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)

	return &sqlHandler{db: db}, nil
}

type sqlHandler struct {
	db *sqlx.DB
}

func (handler *sqlHandler) Execute(statement string, args ...interface{}) (rdb.Result, error) {
	res := sqlResult{}
	result, err := handler.db.Exec(statement, args...)
	if err != nil {
		return res, err
	}
	res.Result = result

	return res, nil
}

func (handler *sqlHandler) Query(statement string, args ...interface{}) (rdb.Row, error) {
	row := new(sqlRow)
	rows, err := handler.db.Query(statement, args...)
	if err != nil {
		return row, err
	}
	row.Rows = rows

	return row, nil
}

type sqlResult struct {
	Result sql.Result
}

func (r sqlResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r sqlResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

type sqlRow struct {
	Rows *sql.Rows
}

func (r sqlRow) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r sqlRow) Next() bool {
	return r.Rows.Next()
}

func (r sqlRow) Close() error {
	return r.Rows.Close()
}
