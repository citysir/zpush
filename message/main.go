package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()

	InitConfig()

	BindRpcAddr()
}
