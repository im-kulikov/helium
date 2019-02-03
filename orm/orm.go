package orm

import (
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

	// Hook is a simple implementation of pg.QueryHook
	Hook struct {
		StartAt time.Time
		Before  func(*pg.QueryEvent)
		After   func(*pg.QueryEvent)
	}
)

// BeforeQuery callback
func (h *Hook) BeforeQuery(e *pg.QueryEvent) {
	h.StartAt = time.Now()

	if h.Before == nil {
		return
	}

	h.Before(e)
}

// AfterQuery callback
func (h Hook) AfterQuery(e *pg.QueryEvent) {
	if h.After == nil {
		return
	}

	h.After(e)
}

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
func NewConnection(opts *Config, l *zap.Logger) (db *pg.DB, err error) {
	if opts == nil {
		err = ErrEmptyConfig
		return
	}

	if l == nil {
		err = ErrEmptyLogger
		return
	}

	l.Debug("Connect to PostgreSQL",
		zap.String("address", opts.Addr),
		zap.String("user", opts.User),
		zap.String("password", opts.Password),
		zap.String("database", opts.Database),
		zap.Int("pool_size", opts.PoolSize))

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
		h := new(Hook)
		h.After = func(e *pg.QueryEvent) {
			query, qErr := e.FormattedQuery()
			l.Debug("pg query",
				zap.String("query", query),
				zap.Duration("query_time", time.Since(h.StartAt)),
				zap.Int("attempt", e.Attempt),
				zap.Any("params", e.Params),
				zap.Error(qErr))
		}
		db.AddQueryHook(h)
	}

	return
}
