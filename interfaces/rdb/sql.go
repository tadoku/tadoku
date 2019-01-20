package rdb

// SQLHandler knows how to run queries against itself
type SQLHandler interface {
	Execute(string, ...interface{}) (Result, error)
	Query(string, ...interface{}) (Row, error)
}

// Result contains meta data of a query that was executed
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Row contains the data from a query
type Row interface {
	Scan(...interface{}) error
	Next() bool
	Close() error
}
