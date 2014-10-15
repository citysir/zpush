package main

import (
	"github.com/citysir/zpush/rpc/message"
)

func NewMessageServiceProxy() (*MesageServiceProxy, error) {
	proxy := new(MessageServiceProxy)
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transport, err := thrift.NewTSocket(Conf.RpcAddr)
	if err != nil {
		return nil, err
	}
	useTransport := transportFactory.GetTransport(transport)
	proxy.client = message.NewMessageServiceClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		return nil, err
	}
	return proxy, nil
}

type MessageServiceProxy struct {
	client *MessageServiceClient
}

func (this *MessageServiceProxy) SavePrivateMessage(key string, message string, msgId int64, expire int64) (err error) {
	return client.SavePrivateMessage(key, message, msgId, expire)
}
