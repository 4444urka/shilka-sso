// Точка входа в приложение
// Для запуска приложения go run ./cmd/sso/main.go --config=config/<Название конфига>
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"shilka-sso/internal/app"
	"shilka-sso/internal/config"
	"shilka-sso/internal/lib/logger/handlers/slogpretty"
	"syscall"
)

// Уровни логирования
const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Загружаем конфиг
	cfg := config.MustLoad()

	// Загружаем логгер
	log := setUpLogger(cfg.Env)

	log.Info("starting sso")

	// Инициализизируем приложение
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// Запускаем сервер
	go application.GRPCServer.MustRun()

	// Делаем так называемый Graceful stop, приложение остановится только когда завершит последний запрос.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCServer.Stop()
}

// Создание логгера
func setUpLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

// Дизайн логгера
func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
