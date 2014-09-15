package main

import (
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/citysir/zpush/rpc/message"
	"os"
)

type MessageServiceImpl struct {
}

func (this *MessageServiceImpl) SavePrivateMessage(key string, message string, msgId int64, expire int64) (err error) {

	return
}

func bindRpcAddr() {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	serverTransport, err := thrift.NewTServerSocket(NetworkAddr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	handler := &MessageServiceImpl{}
	processor := message.NewMessageServiceProcessor(handler)

	server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	fmt.Println("thrift server in", NetworkAddr)
	server.Serve()
}
