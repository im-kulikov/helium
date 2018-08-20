package nats

import (
	"github.com/im-kulikov/helium/module"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	// Config alias
	Config = nats.Options

	// Client alias
	Client = nats.Conn
)

var (
	// Module is default Nats client
	Module = module.Module{
		{Constructor: NewDefaultConfig},
		{Constructor: NewConnection},
	}
	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("nats empty config")
)

// NewDefaultConfig default settings for connection
func NewDefaultConfig(v *viper.Viper) (*Config, error) {
	if !v.IsSet("nats") {
		return nil, ErrEmptyConfig
	}

	var servers []string
	if v.IsSet("nats.servers") {
		servers = v.GetStringSlice("nats.servers")
	}

	return &Config{
		Url:              v.GetString("nats.url"),
		Servers:          servers,
		NoRandomize:      v.GetBool("nats.no_randomize"),
		Name:             v.GetString("nats.name"),
		Verbose:          v.GetBool("nats.verbose"),
		Pedantic:         v.GetBool("nats.pedantic"),
		Secure:           v.GetBool("nats.secure"),
		AllowReconnect:   v.GetBool("nats.allow_reconnect"),
		MaxReconnect:     v.GetInt("nats.max_reconnect"),
		ReconnectWait:    v.GetDuration("nats.reconnect_wait"),
		Timeout:          v.GetDuration("nats.timeout"),
		FlusherTimeout:   v.GetDuration("nats.flusher_timeout"),
		PingInterval:     v.GetDuration("nats.ping_interval"),
		MaxPingsOut:      v.GetInt("nats.max_pings_out"),
		ReconnectBufSize: v.GetInt("nats.reconnect_buf_size"),
		SubChanLen:       v.GetInt("nats.sub_chan_len"),
		User:             v.GetString("nats.user"),
		Password:         v.GetString("nats.password"),
		Token:            v.GetString("nats.token"),
	}, nil
}

// New nats client
func NewConnection(opts *Config) (bus *Client, err error) {
	if opts == nil {
		return nil, ErrEmptyConfig
	}

	if bus, err = opts.Connect(); err != nil {
		return nil, err
	}

	return bus, nil
}
