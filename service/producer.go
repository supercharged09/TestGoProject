package service

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

// FileProducer реализует интерфейс producer для чтения данных из файла, читает файл построчно и возвращает строки в виде слайса строк
type FileProducer struct {
	inputFile string
	log       *slog.Logger
}

// NewFileProducer создает новый экземпляр FileProducer, принимает путь к входному файлу и возвращает настроенный producer
func NewFileProducer(inputFile string) *FileProducer {
	return &FileProducer{
		inputFile: inputFile,
		log:       slog.Default().With("component", "producer"),
	}
}

// Produce читает файл построчно и возвращает все строки в виде слайса, при чтении каждые 100 строк логируется прогресс на уровне debug
// Возвращает []string либо error
func (p *FileProducer) Produce() ([]string, error) {
	p.log.Debug("Начало чтения файла", "file", p.inputFile)

	file, err := os.Open(p.inputFile)
	if err != nil {
		p.log.Error("Ошибка открытия файла",
			"file", p.inputFile,
			"error", err,
		)
		return nil, fmt.Errorf("открытие файла %s: %w", p.inputFile, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			p.log.Error("Ошибка закрытия файла",
				"file", p.inputFile,
				"error", err,
			)
		}
	}()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		lineCount++

		if lineCount%100 == 0 {
			p.log.Debug("Прочитано строк", "count", lineCount)
		}
	}

	if err := scanner.Err(); err != nil {
		p.log.Error("Ошибка сканирования файла",
			"file", p.inputFile,
			"error", err,
		)
		return nil, fmt.Errorf("Чтение файла %s: %w", p.inputFile, err)
	}

	p.log.Info("Файл кспешно прочитан",
		"file", p.inputFile,
		"lines", lineCount,
	)
	return lines, nil
}
