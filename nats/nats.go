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

	// Msg nats message
	Msg = nats.Msg

	// Status of nats connection
	Status = nats.Status

	// ConnHandler is used for asynchronous events such as
	// disconnected and closed connections.
	ConnHandler = nats.ConnHandler

	// MsgHandler is a callback function that processes messages delivered to
	// asynchronous subscribers.
	MsgHandler = nats.MsgHandler

	// Statistics tracks various stats received and sent on this connection,
	// including counts for messages and bytes.
	Statistics = nats.Statistics

	// A Subscription represents interest in a given subject.
	Subscription = nats.Subscription
)

var (
	// Module is default Nats client
	Module = module.Module{
		{Constructor: NewDefaultConfig},
		{Constructor: NewConnection},
	}

	// ErrEmptyConfig when given empty options
	ErrEmptyConfig = errors.New("nats empty config")

	// Name is an Option to set the client name.
	Name = nats.Name

	// Secure is an Option to enable TLS secure connections that skip server verification by default.
	// Pass a TLS Configuration for proper TLS.
	Secure = nats.Secure

	// RootCAs is a helper option to provide the RootCAs pool from a list of filenames. If Secure is
	// not already set this will set it as well.
	RootCAs = nats.RootCAs

	// NoReconnect is an Option to turn off reconnect behavior.
	NoReconnect = nats.NoReconnect

	// DontRandomize is an Option to turn off randomizing the server pool.
	DontRandomize = nats.DontRandomize

	// ReconnectWait is an Option to set the wait time between reconnect attempts.
	ReconnectWait = nats.ReconnectWait

	// MaxReconnects is an Option to set the maximum number of reconnect attempts.
	MaxReconnects = nats.MaxReconnects

	// ReconnectBufSize sets the buffer size of messages kept while busy reconnecting
	ReconnectBufSize = nats.ReconnectBufSize

	// Timeout is an Option to set the timeout for Dial on a connection.
	Timeout = nats.Timeout

	// DisconnectHandler is an Option to set the disconnected handler.
	DisconnectHandler = nats.DisconnectHandler

	// ReconnectHandler is an Option to set the reconnected handler.
	ReconnectHandler = nats.ReconnectHandler

	// ClosedHandler is an Option to set the closed handler.
	ClosedHandler = nats.ClosedHandler

	// DiscoveredServersHandler is an Option to set the new servers handler.
	DiscoveredServersHandler = nats.DiscoveredServersHandler

	// ErrorHandler is an Option to set the async error  handler.
	ErrorHandler = nats.ErrorHandler

	// UserInfo is an Option to set the username and password to
	// use when not included directly in the URLs.
	UserInfo = nats.UserInfo

	// Token is an Option to set the token to use when not included
	// directly in the URLs.
	Token = nats.Token

	// SetCustomDialer is an Option to set a custom dialer which will be
	// used when attempting to establish a connection. If both Dialer
	// and CustomDialer are specified, CustomDialer takes precedence.
	SetCustomDialer = nats.SetCustomDialer

	// UseOldRequestStyle is an Option to force usage of the old Request style.
	UseOldRequestStyle = nats.UseOldRequestStyle
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
