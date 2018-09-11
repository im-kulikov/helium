package orm

import (
	"fmt"
	"time"

	"github.com/go-pg/pg"
	"github.com/im-kulikov/helium/module"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type (
	// Config alias
	Config = struct {
		Addr     string
		User     string
		Password string
		Database string
		Debug    bool
		PoolSize int
		Logger   *zap.SugaredLogger
	}
)

var (
	// Module is default connection to PostgreSQL
	Module = module.Module{
		{Constructor: NewDefaultConfig},
		{Constructor: NewConnection},
	}

	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("database empty config")
	// ErrEmptyLogger when logger not initialized
	ErrEmptyLogger = errors.New("database empty logger")
)

// NewDefaultConfig returns connection config
func NewDefaultConfig(v *viper.Viper) (*Config, error) {
	if !v.IsSet("postgres") {
		return nil, ErrEmptyConfig
	}

	return &Config{
		Addr:     v.GetString("postgres.address"),
		User:     v.GetString("postgres.username"),
		Password: v.GetString("postgres.password"),
		Database: v.GetString("postgres.database"),
		Debug:    v.GetBool("postgres.debug"),
		PoolSize: v.GetInt("postgres.pool_size"),
	}, nil
}

// NewConnection returns database connection
func NewConnection(opts *Config, l *zap.SugaredLogger) (db *pg.DB, err error) {
	if opts == nil {
		err = ErrEmptyConfig
		return
	}

	if l == nil {
		err = ErrEmptyLogger
		return
	}

	l.Debugw("Connect to PostgreSQL",
		"address", opts.Addr,
		"user", opts.User,
		"password", opts.Password,
		"database", opts.Database,
		"pool_size", opts.PoolSize)

	db = pg.Connect(&pg.Options{
		Addr:     opts.Addr,
		User:     opts.User,
		Password: opts.Password,
		Database: opts.Database,
		PoolSize: opts.PoolSize,
	})

	if _, err = db.ExecOne("SELECT 1"); err != nil {
		return nil, errors.Wrap(err, "can't connect to postgres")
	}

	if opts.Debug {
		db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, qErr := event.FormattedQuery()
			l.Debugw(
				fmt.Sprintf("db query %s: \n\t%s", event.Func, query),
				"query_time", time.Since(event.StartTime),
				"error", qErr,
			)
		})
	}

	return
}
