package database

// bunslog provides logging functionalities for Bun using slog.
// This package allows SQL queries issued by Bun to be displayed using slog.

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
)

// option is a function that configures a queryHook.
type option func(*queryHook)

// withLogger sets the *slog.Logger instance.
func withLogger(logger *slog.Logger) option {
	return func(h *queryHook) {
		h.logger = logger
	}
}

func withQueryLogLevel(level slog.Level) option {
	return func(h *queryHook) {
		h.queryLogLevel = level
	}
}

// queryHook is a hook for Bun that enables logging with slog.
// It implements bun.QueryHook interface.
type queryHook struct {
	logger             *slog.Logger
	queryLogLevel      slog.Level
	slowQueryLogLevel  slog.Level
	errorLogLevel      slog.Level
	slowQueryThreshold time.Duration
	logFormat          func(event *bun.QueryEvent) []slog.Attr
	now                func() time.Time
}

// newLogQueryHook initializes a new queryHook with the given options.
func newLogQueryHook(opts ...option) *queryHook {
	h := &queryHook{
		queryLogLevel:      slog.LevelDebug,
		slowQueryLogLevel:  slog.LevelWarn,
		errorLogLevel:      slog.LevelError,
		slowQueryThreshold: 3 * time.Second,
		now:                time.Now,
	}

	for _, opt := range opts {
		opt(h)
	}

	// use default format
	if h.logFormat == nil {
		h.logFormat = func(event *bun.QueryEvent) []slog.Attr {
			duration := h.now().Sub(event.StartTime)

			return []slog.Attr{
				slog.Any("error", event.Err),
				slog.String("operation", event.Operation()),
				slog.String("duration", duration.String()),
			}
		}
	}

	return h
}

// BeforeQuery is called before a query is executed.
func (h *queryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery is called after a query is executed.
// It logs the query based on its duration and whether it resulted in an error.
func (h *queryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	level := h.queryLogLevel
	duration := h.now().Sub(event.StartTime)
	if h.slowQueryThreshold > 0 && h.slowQueryThreshold <= duration {
		level = h.slowQueryLogLevel
	}

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		level = h.errorLogLevel
	}

	attrs := h.logFormat(event)
	if h.logger != nil {
		h.logger.LogAttrs(ctx, level, event.Query, attrs...)
		return
	}

	slog.LogAttrs(ctx, level, event.Query, attrs...)
}

var (
	_ bun.QueryHook = (*queryHook)(nil)
)
