package pgdb

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/masterminds/squirrel"
	"github.com/pkg/errors"
)

type Opts struct {
	URL                string
	MaxOpenConnections int
	MaxIdleConnections int
}

func Open(opts Opts) (*DB, error) {
	db, err := sqlx.Connect("postgres", opts.URL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}
	return &DB{
		db:      db,
		Queryer: newQueryer(db),
	}, nil
}

type Execer interface {
	Exec(query squirrel.Sqlizer) error
	ExecRaw(query string, args ...interface{}) error
}

type Selecter interface {
	Select(dest interface{}, query squirrel.Sqlizer) error
	SelectRaw(dest interface{}, query string, args ...interface{}) error
}

type Getter interface {
	Get(dest interface{}, query squirrel.Sqlizer) error
	GetRaw(dest interface{}, query string, args ...interface{}) error
}

type TransactionFunc func() error

type Transactor interface {
	Transaction(transactionFunc TransactionFunc) (err error)
}

// Connection is yet another thin wrapper for sql.DB allowing to use squirrel queries directly
type Connection interface {
	Transactor
	Queryer
}

// Queryer overloads sqlx's interface name with different meaning, which is not cool.
type Queryer interface {
	Execer
	Selecter
	Getter
}
