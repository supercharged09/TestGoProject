package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type producer interface {
	Produce() ([]string, error)
}

type presenter interface {
	Present([]string) error
}

type Service struct {
	prod producer
	pres presenter
	log  *slog.Logger
}

func NewService(prod producer, pres presenter, log *slog.Logger) *Service {
	return &Service{
		prod: prod,
		pres: pres,
		log:  log,
	}
}

// ToLower функция приводит содержимое к нижнему регистру
func (s *Service) ToLower(str string) string {
	lowerStr := strings.ToLower(str)
	return lowerStr
}

// ReplaceLink функция производит маскировку ссылки http://
func (s *Service) ReplaceLink(lowerStr string) string {
	s.log.Debug("ReplaceLink вход", "строка", lowerStr)

	// Ищем http:// (только HTTP, не HTTPS)
	httpIndex := strings.Index(lowerStr, "http://")

	// Если нет http://, возвращаем строку без изменений
	if httpIndex == -1 {
		return lowerStr
	}

	// Длина протокола "http://" = 7 символов
	protocolLen := 7
	protocolEnd := httpIndex + protocolLen

	// Получаем оставшуюся часть строки после http://
	remaining := lowerStr[protocolEnd:]

	// Ищем конец URL (пробел, таб, конец строки)
	// Убираем знаки пунктуации из поиска, чтобы они считались частью URL
	end := strings.IndexAny(remaining, " \t\n\r")

	if end == -1 {
		// URL до конца строки - маскируем всё после протокола
		result := lowerStr[:protocolEnd] + strings.Repeat("*", len(remaining))
		s.log.Debug("ReplaceLink результат (до конца)", "результат", result)
		return result
	}

	maskedPart := strings.Repeat("*", end)

	remainingAfterURL := remaining[end:]

	// result собирает результат: часть до http:// + http:// + маскированная часть + остаток строки после URL
	result := lowerStr[:protocolEnd] + maskedPart + remainingAfterURL

	s.log.Debug("ReplaceLink результат",
		"протокол", lowerStr[:protocolEnd],
		"маскировано", maskedPart,
		"остаток", remainingAfterURL,
		"результат", result)

	return result
}

// maskLine функция выполняет полную обработку одной строки
func (s *Service) maskLine(line string) string {
	lowerText := s.ToLower(line)
	return s.ReplaceLink(lowerText)
}

// worker читает задачи из входного канала и отправлячет результаты в выходной канал
func (s *Service) worker(ctx context.Context, wg *sync.WaitGroup, workerID int, tasks <-chan task, results chan<- result) {
	defer wg.Done()

	logger := s.log.With(slog.Group("worker",
		slog.Int("id", workerID),
	))

	logger.Debug("Воркер запущен")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Воркер завершает работу по сигналу контекста")
			return
		case t, ok := <-tasks:
			if !ok {
				logger.Debug("Канал задач закрыт, завершение работы")
				return
			}

			logger.Debug("Обработка строки",
				"index", t.index,
				"lenght", len(t.line),
			)

			// maskedText - маскирование для полученной строки
			maskedText := s.maskLine(t.line)

			// отправка результата
			select {
			case <-ctx.Done():
				logger.Info("Воркер завершает работу во время отправки результата")
				return
			case results <- result{
				index: t.index,
				line:  maskedText,
			}:
				logger.Debug("Строка обработана", "index", t.index)
			}
		}
	}
}

// task представляет задачу на обработку одной строки
type task struct {
	index int
	line  string
}

// result представляет результат обработки одной строки
type result struct {
	index int
	line  string
}

func (s *Service) Run(ctx context.Context) error {
	s.log.Info("Запуск сервиса маскирования ссылок")

	s.log.Debug("Чтение данных из продюсера")
	data, err := s.prod.Produce()
	if err != nil {
		return fmt.Errorf("получение данных: %w", err)
	}

	s.log.Info("Данные получены", "количество строк", len(data))

	// создание каналов для задач и результатов
	tasks := make(chan task, len(data))
	results := make(chan result, len(data))

	// создание WG для ожидания завершения всех воркеров
	var wg sync.WaitGroup

	//запуск воркеров (<=10 одновременно)
	numWorkers := 10
	if len(data) < numWorkers {
		numWorkers = len(data)
	}

	s.log.Info("Запуск воркеров", "количество", numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, i, tasks, results)
	}

	// отправка задач в канал
	s.log.Debug("Отправка задач воркерам")
	for i, line := range data {
		select {
		case <-ctx.Done():
			s.log.Warn("Отмена отправки задач по сигналу контекста")
			close(tasks)
			return ctx.Err()
		case tasks <- task{
			index: i,
			line:  line,
		}:
			// задача успешно отправлена
		}
	}
	close(tasks) // закрытие канала задач
	s.log.Debug("Все задачи отправлены")

	// запуск горутины для закрытия канала результатов после завершения всех воркеров
	go func() {
		wg.Wait()
		close(results)
		s.log.Debug("Все воркеры завершили работу, канал результатов закрыт")
	}()

	// сборка результатов в правильном порядке
	s.log.Debug("Сбор результатов")
	resultSlice := make([]string, len(data))

	for {
		select {
		case <-ctx.Done():
			s.log.Warn("Сбор результатов прерван по сигналу контекста")
			return ctx.Err()
		case r, ok := <-results:
			if !ok {
				// канал результатов закрыт
				s.log.Debug("Сбор результатов завершен")

				// проверка, все ли результаты получены
				collected := 0
				for i, line := range resultSlice {
					if line != "" {
						collected++
					} else {
						s.log.Warn("Результат для индекса не получен", "index", i)
					}
				}
				s.log.Info("Результаты собраны",
					"ожидалось", len(data),
					"получено", collected,
				)

				// сохранение результатов через презентер
				s.log.Info("Сохранение результатов в файл")
				if err := s.pres.Present(resultSlice); err != nil {
					return fmt.Errorf("сохранение результатов: %w", err)
				}
				return nil
			}
			resultSlice[r.index] = r.line
		}
	}
}
