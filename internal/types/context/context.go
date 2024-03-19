package context

import (
	"context"

	"github.com/intuit/naavik/pkg/logger"
)

type Context struct {
	Log     logger.Logger
	Context context.Context
}

func NewContextWithLogger() Context {
	return Context{
		Log:     logger.NewLogger(),
		Context: context.Background(),
	}
}

func Background() context.Context {
	return context.Background()
}

type Key string

func (c Key) String() string {
	return string(c)
}
