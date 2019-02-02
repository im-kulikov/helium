package nats

import (
	"github.com/im-kulikov/helium/module"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	// Config alias
	Config = nats.Options

	// StreamerConfig
	StreamerConfig struct {
		ClientID  string
		ClusterID string
		Options   []stan.Option
	}

	// Client alias
	Client = nats.Conn
)

var (
	// Module is default Nats client
	Module = module.Module{
		{Constructor: NewDefaultConfig},
		{Constructor: NewConnection},
		{Constructor: NewDefaultStreamerConfig},
		{Constructor: NewStreamer},
	}

	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("nats empty config")
	// ErrEmptyStreamerConfig when given empty options
	ErrEmptyStreamerConfig = errors.New("nats-streamer empty config")
	// ErrNoNatsConnection when empty nats.Conn
	ErrEmptyConnection = errors.New("nats connection empty")
	// ErrClusterIDEmpty when empty clusterID
	ErrClusterIDEmpty = errors.New("nats.cluster_id cannot be empty")
	// ErrClientIDEmpty when empty clientID
	ErrClientIDEmpty = errors.New("nats.client_id cannot be empty")
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

// NewDefaultConfig default settings for streaming connection
func NewDefaultStreamerConfig(v *viper.Viper, bus *Client) (*StreamerConfig, error) {
	if !v.IsSet("nats") {
		return nil, ErrEmptyConfig
	}

	var clusterID, clientID string
	if clusterID = v.GetString("nats.cluster_id"); clusterID == "" {
		return nil, ErrClusterIDEmpty
	}

	if clientID = v.GetString("nats.client_id"); clientID == "" {
		return nil, ErrClientIDEmpty
	}

	return &StreamerConfig{
		ClientID:  clientID,
		ClusterID: clusterID,
		Options:   []stan.Option{stan.NatsConn(bus)},
	}, nil
}

// NewConnection of nats client
func NewConnection(opts *Config) (bus *Client, err error) {
	if opts == nil {
		return nil, ErrEmptyConfig
	}

	if bus, err = opts.Connect(); err != nil {
		return nil, err
	}

	return bus, nil
}

// NewStreamer is nats-streamer client
func NewStreamer(opts *StreamerConfig) (stan.Conn, error) {
	if opts == nil {
		return nil, ErrEmptyStreamerConfig
	}

	if opts.Options == nil || len(opts.Options) == 0 {
		return nil, ErrEmptyConnection
	}

	return stan.Connect(opts.ClusterID, opts.ClientID, opts.Options...)
}
