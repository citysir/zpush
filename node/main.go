package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	fmt.Println("Hello, 世界")
	if err := InitConfig(); err != nil {
		panic(err)
	}
}
