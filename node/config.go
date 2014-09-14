package main

import (
	"encoding/json"
	"flag"
	"github.com/citysir/golib/io/fileutil"
	"log"
	"runtime"
)

var (
	Conf *Config
)

func init() {
	var confFile string
	flag.StringVar(&confFile, "c", "./node.conf", " set node config file path")
	flag.Parse()
	if err := initConfig(confFile); err != nil {
		panic(err)
	}
}

type Config struct {
	PidFile       string `json:"PidFile"`
	Log           string `json:"Log"`
	TCPBind       string `json:"TCPBind"`
	RPCBind       string `json:"RPCBind"`
	PprofBind     string `json:"PprofBind"`
	StatBind      string `json:"StatBind"`
	MaxProc       int    `json:"MaxProc"`
	KetamaBase    int    `json:"KetamaBase"`
	ChannelBucket int    `json:"ChannelBucket"`
}

func initConfig(confFile string) {
	Conf = &Config{
		// base
		PidFile:    "/tmp/zpush-node.pid",
		Log:        "./log/xml",
		MaxProc:    runtime.NumCPU(),
		TCPBind:    "localhost:6969",
		RPCBind:    "localhost:6970",
		PprofBind:  "localhost:6971",
		StatBind:   "localhost:6972",
		KetamaBase: 255,

		// channel
		ChannelBucket: runtime.NumCPU(),
	}

	confText, err := fileutil.ReadText(confFile)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		panic(err)
	}

	log.Println(confText)
}
