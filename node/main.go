package main

import "fmt"

func main() {
	fmt.Println("Hello, 世界")
	if err := InitConfig(); err != nil {
		panic(err)
	}
}
