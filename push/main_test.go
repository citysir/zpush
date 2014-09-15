package main

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/citysir/zpush/rpc/push"
	"net"
	"os"
	"testing"
	"time"
)

// go test github.com/citysir/zpush/push -test.v
func Test_Push(t *testing.T) {
	startTime := currentTimeMillis()
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTSocket("127.0.0.1:19090")
	if err != nil {
		t.Log(os.Stderr, "error resolving address:", err)
		os.Exit(1)
	}

	useTransport := transportFactory.GetTransport(transport)
	client := push.NewPushServiceClientFactory(useTransport, protocolFactory)
	if err := transport.Open(); err != nil {
		t.Log(os.Stderr, "Error opening socket to 127.0.0.1:19090", " ", err)
		os.Exit(1)
	}
	defer transport.Close()

	for i := 0; i < 1000; i++ {
		paramMap := make(map[string]string)
		paramMap["name"] = "qinerg"
		paramMap["passwd"] = "123456"
		result, err := client.FunCall(currentTimeMillis(), "login", paramMap)
		t.Log(i, "Call->", result, err)
	}

	endTime := currentTimeMillis()
	t.Log("Program exit. time->", endTime, startTime, (endTime - startTime))
}

// 转换成毫秒
func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}
