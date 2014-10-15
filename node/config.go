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
	flag.StringVar(&confFile, "c", "./node.conf", "set node config file path")
}

type Config struct {
	PidFile       string `json:"PidFile"`
	Log           string `json:"Log"`
	TcpAddr       string `json:"TcpAddr"`
	RpcAddr       string `json:"RpcAddr"`
	PprofAddr     string `json:"PprofAddr"`
	StatAddr      string `json:"StatAddr"`
	MaxCore       int    `json:"MaxCore"`
	KetamaBase    int    `json:"KetamaBase"`
	ChannelBucket int    `json:"ChannelBucket"`

	MsgBufNum int `json:"MsgBufNum"`

	WriteBufferSize     int  `json:"WriteBufferSize"`
	ReadeBufferSize     int  `json:"ReadBufferSize"`
	BufioInstance       int  `json:"BufioInstance"`
	BufioNumPerInstance int  `json:"BufioNumPerInstance"`
	TcpKeepalive        bool `json:"TcpKeepalive"`

	ZookeeperAddr            string `json:"ZookeeperAddr"`
	ZookeeperTimeout         int    `json:"ZookeeperTimeout"`
	ZookeeperLocation        string `json:"ZookeeperLocation"`
	ZookeeperName            string `json:"ZookeeperName"`
	ZookeeperWeight          int    `json:"ZookeeperWeight"`
	ZookeeperMessageLocation string `json:"ZookeeperMessageLocation"`
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
	Conf.ChannelBucket = math.MaxInt32(Conf.ChannelBucket, runtime.NumCPU())
	Conf.BufioInstance = math.MaxInt32(Conf.BufioInstance, runtime.NumCPU())
}
