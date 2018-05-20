package orm

import (
	"fmt"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/go-pg/pg/types"
	"github.com/im-kulikov/helium/logger"
	"github.com/pkg/errors"
)

type (
	// DB alias
	DB = pg.DB
	// Query alias
	Query = orm.Query
	// Config alias
	Config = struct {
		Addr     string
		User     string
		Password string
		Database string
		Debug    bool
		PoolSize int
	}
)

var (
	// InSlice alias
	InSlice = types.InSlice
	// ErrNoRows alias
	ErrNoRows = pg.ErrNoRows
	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("database empty config")
	// ErrEmptyLogger when logger not inited
	ErrEmptyLogger = errors.New("database empty logger")
)

// New database connection
func New(opts *Config) (db *DB, err error) {
	if opts == nil {
		err = ErrEmptyConfig
		return
	}

	if logger.G() == nil {
		err = ErrEmptyLogger
		return
	}

	db = pg.Connect(&pg.Options{
		Addr:     opts.Addr,
		User:     opts.User,
		Password: opts.Password,
		Database: opts.Database,
		PoolSize: opts.PoolSize,
	})

	if _, err = db.ExecOne("SELECT 1"); err != nil {
		return nil, err
	}

	if opts.Debug {
		db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, qErr := event.FormattedQuery()
			logger.G().Debugw(
				fmt.Sprintf("db query %s: \n\t%s", event.Func, query),
				"query_time", time.Since(event.StartTime),
				"error", qErr,
			)
		})
	}

	return
}
