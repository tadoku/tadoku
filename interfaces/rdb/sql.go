package rdb

// SQLHandler knows how to run queries against itself
type SQLHandler interface {
	Execute(string, ...interface{}) (Result, error)
	Query(string, ...interface{}) (Rows, error)
	QueryRow(string, ...interface{}) Row

	NamedExecute(string, interface{}) (Result, error)

	Get(dest interface{}, query string, args ...interface{}) error
}

// Result contains meta data of a query that was executed
type Result interface {
	LastInsertId() (int64, error)
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
