package rdb

type sqlQueryHandler interface {
	Query(string, ...interface{}) (Rows, error)
	QueryRow(string, ...interface{}) Row

	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type sqlQueryExecuter interface {
	Execute(string, ...interface{}) (Result, error)
	NamedExecute(string, interface{}) (Result, error)
}

// SQLHandler knows how to run queries against itself
type SQLHandler interface {
	sqlQueryHandler
	sqlQueryExecuter

	Begin() (TxHandler, error)
}

// TxHandler is a SQLHandler in a transaction context
type TxHandler interface {
	sqlQueryHandler
	sqlQueryExecuter

	Commit() error
	Rollback() error
}

// Result contains meta data of a query that was executed
type Result interface {
	LastInsertID() (int64, error)
	RowsAffected() (int64, error)
}

// Row contains the data from a query
type Row interface {
	Scan(...interface{}) error
	StructScan(interface{}) error
}

// Rows contains the data from a query
type Rows interface {
	Scan(...interface{}) error
	StructScan(interface{}) error
	Next() bool
	Close() error
}
