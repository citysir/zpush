package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	fmt.Println("Hello, node")
	InitConfig()
}
