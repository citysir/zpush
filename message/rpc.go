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
	fmt.Println("SavePrivateMessage", key, message, msgId, expire)
	return
}

func BindRpcAddr(addr string) {
	go rpcListen(addr)
}

func rpcListen(addr string) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	serverTransport, err := thrift.NewTServerSocket(addr)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	handler := &MessageServiceImpl{}
	processor := message.NewMessageServiceProcessor(handler)

	server := thrift.NewTSimpleServer4(processor, serverTransport, transportFactory, protocolFactory)
	fmt.Println("thrift server in", addr)
	server.Serve()
}
