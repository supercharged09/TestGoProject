package main

import "fmt"

func main() {
	var num int
	fmt.Print("Введите оценку (0-100) :")
	fmt.Scan(&num)
	switch {
	case num >= 90:
		fmt.Println("A")
	case num >= 80:
		fmt.Println("B")
	case num >= 70:
		fmt.Println("C")
	case num >= 60:
		fmt.Println("D")
	default:
		fmt.Println("F")
	}
}
