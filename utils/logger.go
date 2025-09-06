package utils

import (
	"channel-helper-go/utils/tint"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/cappuccinotm/slogx"
	"github.com/gookit/slog/rotatefile"
)

// ApplyHandler wraps slog.Handler as Middleware.
func applyHandler(handler slog.Handler) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			err := handler.Handle(ctx, rec)
			if err != nil {
				return err
			}

			return next(ctx, rec)
		}
	}
}

type Logger = *slog.Logger

func createLogsDir(name string) {
	err := os.MkdirAll(path.Join("logs", name), os.ModePerm)
	if err != nil {
		panic("failed to create logs directory: " + err.Error())
	}
}

func NewLogger(name string, module string) Logger {
	createLogsDir(name)
	level := slog.LevelDebug

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:  level,
		Prefix: fmt.Sprintf("[%s]", module),
	})

	var handlers []slogx.Middleware

	writer, err := rotatefile.NewConfig(
		path.Join("logs", name, fmt.Sprintf("%s.log", module)),
		func(c *rotatefile.Config) {
			c.MaxSize = 10 * 1024 * 1024 // 10 MB
			c.BackupNum = 5
			c.RotateTime = rotatefile.EveryMonth
			c.Compress = true
		},
	).Create()
	if err != nil {
		panic("failed to create log file: " + err.Error())
	}
	handlers = append(handlers,
		applyHandler(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})),
	)

	return slog.New(slogx.Accumulator(slogx.NewChain(consoleHandler, handlers...)))
}
