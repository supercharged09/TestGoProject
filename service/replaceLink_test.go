package service

import "testing"

// ДОПОЛНИТЕЛЬНЫЙ ТЕСТ ДЛЯ ОТЛАДКИ:
func TestReplaceLinkDebug(t *testing.T) {
	s := &Service{}

	// Проверьте как работает ваш метод
	testCases := []struct {
		input    string
		expected string
	}{
		{"http://example.com text", "http://********** text"},
		{"http://site text", "http://**** text"},
		{"no link", "no link"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := s.ReplaceLink(tc.input)
		t.Logf("ReplaceLink(%q) = %q, expected %q", tc.input, result, tc.expected)
	}
}
