package service

import (
	"strings"
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
}

func NewService(prod producer, pres presenter) *Service {
	return &Service{
		prod: prod,
		pres: pres,
	}
}

func (s *Service) ToLower(str string) string {
	lowerStr := strings.ToLower(str)
	return lowerStr
}

func (s *Service) ReplaceLink(lowerStr string) string {
	index := strings.Index(lowerStr, "//")
	if index = strings.Index(lowerStr, "//"); index != -1 {
		start := index + 2
		end := strings.IndexAny(lowerStr[start:], " ")
		if end > 0 {
			return lowerStr[:start] + strings.Repeat("*", end) + lowerStr[start+end:]
		}
	}
	return lowerStr
}

func (s *Service) Run() error {
	data, err := s.prod.Produce()
	if err != nil {
		return err
	}

	result := make([]string, len(data))

	for i, line := range data {
		lowerText := s.ToLower(line)
		result[i] = s.ReplaceLink(lowerText)
	}
	return s.pres.Present(result)
}
