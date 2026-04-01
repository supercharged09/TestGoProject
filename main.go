package main

import (
	"TestGoProject/service"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "link-replacer",
		Usage: "Заменяет ссылки в текстовом файле",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Путь к входному файлу(по умолчанию в корне --input input.txt)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Путь к выходному файлу",
				Value:   "output.txt",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Уровень логирования (debug, info, warn, error)",
				Value:   "info",
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("Ошибка при запуске приложения", "error", err)
		os.Exit(1)
	}
}

// func run с настройкой логирования
func run(c *cli.Context) error {
	logLevel := c.String("log-level")
	setupLogger(logLevel)

	//контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// sigChan канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// горутина для обработки сигналов
	go func() {
		sig := <-sigChan
		slog.Info("Получен сигнал завершения", "signal", sig)
		cancel()
	}()

	// inputFile\outputFile - компоненты сервиса
	inputFile := c.String("input")
	outputFile := c.String("output")

	prod := service.NewFileProducer(inputFile)
	pres := service.NewFilePresenter(outputFile)
	serv := service.NewService(prod, pres, slog.Default())

	// запуск сервиса с контекстом
	if err := serv.Run(ctx); err != nil {
		slog.Error("Ошибка выполнения сервиса", "error", err)
		return err
	}

	slog.Info("Программа успешно завершена")
	return nil
}

func setupLogger(level string) {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Debug("Логирование настроено", "уровень", level)
}
