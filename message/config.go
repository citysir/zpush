package main

import (
	"encoding/json"
	"flag"
	"github.com/citysir/golib/io/fileutil"
	"log"
	"math"
	"runtime"
)

var (
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./message.conf", "set message config file path")
}

type Config struct {
	PidFile   string `json:"PidFile"`
	Log       string `json:"Log"`
	RpcAddr   string `json:"RpcAddr"`
	PprofAddr string `json:"PprofAddr"`
	MaxCore   int    `json:"MaxCore"`
}

func InitConfig() {
	log.Printf("InitConfig %s\n", confFile)

	confText, err := fileutil.ReadText(confFile)
	if err != nil {
		panic(err)
	}

	log.Println(confText)

	var config Config
	err = json.Unmarshal([]byte(confText), &config)
	if err != nil {
		panic(err)
	}

	Conf = &config

	Conf.MaxCore = math.MaxInt32(Conf.MaxCore, runtime.NumCPU())
}
