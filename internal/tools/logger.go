package tools

import (
	"awesomeProject/internal/config"
	"log/slog"
	"os"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

func InitLogger(config *config.Config) {
	once.Do(func() {
		var handler slog.Handler
		env := config.AppEnv

		if env == "production" {
			handler = slog.NewJSONHandler(os.Stdout, nil)
		} else {
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
			})
		}

		logger = slog.New(handler)
		slog.SetDefault(logger) // глобальный по умолчанию
	})
}
