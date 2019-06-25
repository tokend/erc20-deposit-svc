package pgdb

import (
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Databaser interface {
	DB() *DB
}

type databaser struct {
	getter kv.Getter
	once   comfig.Once
}

func NewDatabaser(getter kv.Getter) Databaser {
	return &databaser{
		getter: getter,
	}
}

func (d *databaser) DB() *DB {
	return d.once.Do(func() interface{} {
		var config = struct {
			URL                string `fig:"url,required"`
			MaxOpenConnections int    `fig:"max_open_connection"`
			MaxIdleConnections int    `fig:"max_idle_connections"`
		}{
			MaxOpenConnections: 12,
			MaxIdleConnections: 12,
		}

		err := figure.Out(&config).
			From(kv.MustGetStringMap(d.getter, "db")).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to figure out"))
		}

		db, err := Open(Opts{
			URL:                config.URL,
			MaxOpenConnections: config.MaxOpenConnections,
			MaxIdleConnections: config.MaxIdleConnections,
		})
		if err != nil {
			panic(errors.Wrap(err, "failed to open database"))
		}

		return db
	}).(*DB)
}
