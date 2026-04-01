package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моки для интерфейсов producer и presenter, mock.Mock для отслеживания вызовов методов
type MockProducer struct{ mock.Mock }
type MockPresenter struct{ mock.Mock }

// Реализуем метод Produce для мока
func (m *MockProducer) Produce() ([]string, error) {
	args := m.Called() // m.Called возвращает объект с аргументами, переданными при настройке мока
	// args.Get(0) получает первый возвращаемый параметр ([]string)
	// args.Error(1) получает второй параметр (error)
	return args.Get(0).([]string), args.Error(1)
}

// Реализация метода Present для мока
func (m *MockPresenter) Present(lines []string) error {
	// m.Called(lines) фиксирует вызов метода с конкретными аргументами
	// Error(0) возвращает первую ошибку из настроенных возвращаемых значений
	return m.Called(lines).Error(0)
}

// Основной тест
func TestService(t *testing.T) {
	// Экземпляр сервиса для тестирования методов ToLower и ReplaceLink
	s := &Service{}
	//Проверка преобразования строки в нижний регистр
	assert.Equal(t, "hello", s.ToLower("HELLO"))
	assert.Equal(t, "привет", s.ToLower("ПРИВЕТ"))
	assert.Equal(t, "мне 27 лет!", s.ToLower("МНЕ 27 ЛЕТ!"))
	assert.Equal(t, "", s.ToLower(""))
	assert.Equal(t, "my name john doe", s.ToLower("My NaMe JOhN dOe"))

	// Тестирование метода ReplaceLink
	assert.Equal(t, "http://********** text", s.ReplaceLink("http://mysite.com text"))
	// Тестирование текста без ссылки, со ссылкой, пустая строка
	assert.Equal(t, "text without link", s.ReplaceLink("text without link"))
	assert.Equal(t, "visit http://******** now", s.ReplaceLink("visit http://site.com now"))
	assert.Equal(t, "http://site.comwithoutspace", s.ReplaceLink("http://site.comwithoutspace"))
	assert.Equal(t, "", s.ReplaceLink(""))

	// Сценарий 1 - успешное выполнение метода Run
	t.Run("Run", func(t *testing.T) {
		//Создание мок объектов
		prod, pres := &MockProducer{}, &MockPresenter{}
		// Создание Service с зависимостями-моками
		svc := NewService(prod, pres)
		// При вызове prod.Produce() вернуть ["TEST"] и nil (без ошибки)
		prod.On("Produce").Return([]string{"TEST"}, nil)
		// При вызове pres.Present(["test"]) вернуть nil
		// "test" то "TEST" после преобразования ToLower
		pres.On("Present", []string{"test"}).Return(nil)

		// Выполняем тестируемый метод
		err := svc.Run()

		assert.NoError(t, err)
		// Все ожидаемые вызовы методов на prod были выполнены
		prod.AssertExpectations(t)
		// Все ожидаемые вызовы методов на pres были выполнены
		pres.AssertExpectations(t)
	})

	// Сценарий 2 - Ошибка в методе Run
	t.Run("Run error", func(t *testing.T) {
		prod, pres := &MockProducer{}, &MockPresenter{}
		svc := NewService(prod, pres)

		// prod.Produce() возвращает пустой массив и стандартную ошибку из testify
		prod.On("Produce").Return([]string{}, assert.AnError)

		//pres.Present() не настраивать, он не должен вызываться при ошибке producer

		// Метод вернул ошибку (т.к producer вернул ошибку)
		err := svc.Run()
		// Проверяем сохраненный результат
		assert.Error(t, err)
		prod.AssertExpectations(t)
		// Метод pres.Present() не был вызван, при ошибке producer работа должна прекратиться
		pres.AssertNotCalled(t, "Present", mock.Anything)
	})

}
