package main

import (
	"TestGoProject/service"
	"fmt"
	"os"
)

func main() {

	input := "C:/Users/User/GolandProjects/TestGoProject/input.txt"
	output := "C:/Users/User/GolandProjects/TestGoProject/output.txt"

	prod := service.NewFileProducer(input)
	pres := service.NewFilePresenter(output)
	serv := service.NewService(prod, pres)

	err := serv.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Записано успешно")
	fmt.Println(output)

	content, err := os.ReadFile(output)
	if err != nil {
		fmt.Println("Ошибка чтения файла", err)
		return
	}
	fmt.Println(string(content))
}
