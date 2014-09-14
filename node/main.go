package main

import (
	"flag"
	"fmt"
	"github.com/citysir/zpush/perf"
	"runtime"
)

func main() {
	flag.Parse()
	InitConfig()

	runtime.GOMAXPROCS(Conf.MaxCore)

	perf.BindAddr(Conf.PprofAddr)

	BindStatAddr(Conf.StatAddr)

	ChannelManager = NewChannelManager()
	defer ChannelManager.Close()

	BindTcpAddr(Conf.TcpAddr)
}
