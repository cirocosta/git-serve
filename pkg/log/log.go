package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

var SetLevel = logrus.SetLevel

type (
	Fields = logrus.Fields
	Logger = logrus.Entry
)

type ctxKey struct{}

var defaultLogger = logrus.NewEntry(logrus.StandardLogger())

func Verbose() {
	logrus.SetLevel(logrus.DebugLevel)
}

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	e := logger.WithContext(ctx)
	return context.WithValue(ctx, ctxKey{}, e)
}

func From(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(ctxKey{})
	if logger == nil {
		return defaultLogger.WithContext(ctx)
	}

	return logger.(*logrus.Entry)
}
