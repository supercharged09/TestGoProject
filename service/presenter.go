package service

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// FilePresenter реализует интерфейс presenter для записи данных в текстовый файл, записывает слайс строк в файл, разделяя переводом строки
type FilePresenter struct {
	outputFile string
	log        *slog.Logger
}

// NewFilePresenter создает новый экземпляр FilePresenter, принимает путь к входному файлу.
// Если путь пустой, изпользуется значение по дефолту "output.txt"
func NewFilePresenter(outputFile string) *FilePresenter {
	if outputFile == "" {
		outputFile = "output.txt"
	}
	return &FilePresenter{
		outputFile: outputFile,
		log:        slog.Default().With("component", "presenter"),
	}
}

// Present записывает слайс строк в выходной файл
// При наличии ошибок создания файла или записи данных возвращает ошибку с описанием
func (p *FilePresenter) Present(lines []string) error {
	p.log.Debug("Начало записи в файл",
		"file", p.outputFile,
		"lines_count", len(lines),
	)

	file, err := os.Create(p.outputFile)
	if err != nil {
		p.log.Error("Ошибка создания файла",
			"file", p.outputFile,
			"error", err,
		)
		return fmt.Errorf("создание файла %s: %w", p.outputFile, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			p.log.Error("Ошибка закрытия файла",
				"file", p.outputFile,
				"error", err,
			)
		}
	}()

	// Объединение строк с переносами и добавление завершающего переноса строки
	data := []byte(strings.Join(lines, "\n") + "\n")
	bytesWritten, err := file.Write(data)
	if err != nil {
		p.log.Error("ошибка записи в файл",
			"file", p.outputFile,
			"error", err,
		)
		return fmt.Errorf("запись в файл %s: %w", p.outputFile, err)
	}

	p.log.Info("Файл успешно записан",
		"file", p.outputFile,
		"bytes_written", bytesWritten,
		"lines", len(lines),
	)
	return nil
}
