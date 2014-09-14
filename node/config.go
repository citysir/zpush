package main

import (
	"encoding/json"
	"flag"
	"github.com/citysir/golib/io/fileutil"
	"log"
	"runtime"
)

var (
	Conf     *Config
	confFile string
)

func init() {
	var confFile string
	flag.StringVar(&confFile, "c", "./node.conf", "set node config file path")
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

func InitConfig() error {
	flag.Parse()
	log.Printf("conf %s\n", confFile)
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
		return err
	}

	var config Config
	err = json.Unmarshal([]byte(confText), &config)
	if err != nil {
		return err
	}

	log.Println(confText)
	return nil
}
