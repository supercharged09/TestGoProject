package service

import (
	"os"
	"strings"
)

type FilePresenter struct {
	outputFile string
}

func NewFilePresenter(outputFile string) *FilePresenter {
	if outputFile == "" {
		outputFile = "output.txt"
	}
	return &FilePresenter{outputFile: outputFile}
}

func (p *FilePresenter) Present(lines []string) error {
	file, err := os.Create(p.outputFile)
	if err != nil {
		return err
	}

	defer file.Close()

	data := []byte(strings.Join(lines, "\n") + "\n")
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
