package tty

import (
	"context"
)

type options struct {
	ctx           context.Context
	workdir       *string
	extraEnv      []string
	useCurrentEnv bool
}

type OptionFunc func(option *options)

func newOption(fs ...OptionFunc) *options {
	opt := &options{
		ctx:           context.Background(),
		useCurrentEnv: true,
	}

	for _, f := range fs {
		f(opt)
	}

	return opt
}

func WithContext(ctx context.Context) OptionFunc {
	return func(opt *options) {
		opt.ctx = ctx
	}
}

// WithWorkdir sets the working directory for the command.
func WithWorkdir(workdir string) OptionFunc {
	return func(opt *options) {
		opt.workdir = &workdir
	}
}

// WithExtraEnv sets the extra environment variables for the command. format: key=value
func WithExtraEnv(extraEnv ...string) OptionFunc {
	return func(opt *options) {
		opt.extraEnv = extraEnv
	}
}

// WithCurrentEnv sets whether to use the current environment variables for the command.
func WithEmptyEnv() OptionFunc {
	return func(opt *options) {
		opt.useCurrentEnv = false
	}
}
