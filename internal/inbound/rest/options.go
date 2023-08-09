package rest

import (
	"io"
	"time"
)

// Option provides a mechanism to optionally override default server
// configuration.
type Option interface {
	apply(*config)
}

// LogOutputOption overrides the default writer the server outputs logs to.
type LogOutputOption struct {
	W io.Writer
}

func (o *LogOutputOption) apply(cfg *config) {
	cfg.logOutput = o.W
}

// LogPrefixOptions overrides the default prefix the server uses when logging.
type LogPrefixOptions struct {
	Prefix string
}

func (o *LogPrefixOptions) apply(cfg *config) {
	cfg.logPrefix = o.Prefix
}

// LogFlagsOption overrides the default flags the server uses when logging.
type LogFlagsOption struct {
	Flags int
}

func (o *LogFlagsOption) apply(cfg *config) {
	cfg.logFlags = o.Flags
}

// ReadTimeoutOption overrides the server's default read timeout.
type ReadTimeoutOption struct {
	Timeout time.Duration
}

func (o *ReadTimeoutOption) apply(cfg *config) {
	cfg.readTimeout = o.Timeout
}

// WriteTimeoutOption overrides the server's default write timeout.
type WriteTimeoutOption struct {
	Timeout time.Duration
}

func (o *WriteTimeoutOption) apply(cfg *config) {
	cfg.writeTimeout = o.Timeout
}

// StackTraceOption overrides the server default for printing stack traces when
// recovering from panics.
type StackTraceOption struct {
	Enable bool
}

func (o *StackTraceOption) apply(cfg *config) {
	cfg.enableStackTrace = o.Enable
}

// JwtTtlOption overrides the server default for the time-to-live of JWTs.
type JwtTtlOption struct {
	TTL time.Duration
}

func (o *JwtTtlOption) apply(cfg *config) {
	cfg.jwtTTL = o.TTL
}
