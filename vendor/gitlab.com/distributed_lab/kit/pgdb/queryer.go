package pgdb

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/masterminds/squirrel"
	"github.com/pkg/errors"
)

type rawQueryer interface {
	sqlx.Execer
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

type queryer struct {
	raw rawQueryer
}

func newQueryer(raw rawQueryer) *queryer {
	return &queryer{
		raw: raw,
	}
}

func (q *queryer) Get(dest interface{}, query squirrel.Sqlizer) error {
	sql, args, err := build(query)
	if err != nil {
		return err
	}
	return q.GetRaw(dest, sql, args...)
}

func (q *queryer) GetRaw(dest interface{}, query string, args ...interface{}) error {
	query = rebind(query)
	err := q.raw.Get(dest, query, args...)

	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return err
	}

	return errors.Wrap(err, "failed to get raw")
}

func (q *queryer) Exec(query squirrel.Sqlizer) error {
	sql, args, err := build(query)
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	return q.ExecRaw(sql, args...)
}

func (q *queryer) ExecRaw(query string, args ...interface{}) error {
	query = rebind(query)
	_, err := q.raw.Exec(query, args...)
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return err
	}

	return errors.Wrap(err, "failed to exec query")
}

// Select runs `query`, setting the results found on `dest`.
func (q *queryer) Select(dest interface{}, query squirrel.Sqlizer) error {
	sql, args, err := build(query)
	if err != nil {
		return err
	}
	return q.SelectRaw(dest, sql, args...)
}

// SelectRaw runs `query` with `args`, setting the results found on `dest`.
func (q *queryer) SelectRaw(dest interface{}, query string, args ...interface{}) error {
	//r.clearSliceIfPossible(dest) // TODO wat?
	query = rebind(query)
	err := q.raw.Select(dest, query, args...)

	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return err
	}

	return errors.Wrap(err, "failed to select")
}

func rebind(stmt string) string {
	return sqlx.Rebind(sqlx.BindType("postgres"), stmt)
}

func build(b squirrel.Sqlizer) (sql string, args []interface{}, err error) {
	sql, args, err = b.ToSql()

	if err != nil {
		err = errors.Wrap(err, "failed to parse query")
	}
	return
}
