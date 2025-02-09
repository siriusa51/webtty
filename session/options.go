package session

import (
	"log/slog"
	"os"
)

type options struct {
	logHandler slog.Handler
}

type OptionFunc func(*options)

func newOptions(optfs ...OptionFunc) *options {
	opt := &options{
		logHandler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
	}

	for _, optf := range optfs {
		optf(opt)
	}

	return opt
}

// WithLogHandler sets the log handler for the session.
func WithLogHandler(logHandler slog.Handler) OptionFunc {
	return func(o *options) {
		o.logHandler = logHandler
	}
}
