package main

import (
	"bufio"
)

// The Message struct
type Message struct {
	GroupId uint   `json:"gid"` // group id
	MsgId   int64  `json:"mid"` // message id
	Msg     string `json:"msg"` // message content
}

const (
	MinCmdNum = 1
	MaxCmdNum = 5

	CmdHeartBeat = "B"
	CmdSubscribe = "S"

	HeartbeatReply    = []byte("+B\r\n") // hearbeat
	AuthFailedReply   = []byte("-a\r\n") // auth failed reply
	ChannelErrorReply = []byte("-c\r\n") // channel not found reply
	ParamErrorReply   = []byte("-p\r\n") // param error reply
	NodeErrorReply    = []byte("-n\r\n") // node error reply
)

var (
	// cmd parse failed
	ErrProtocol = errors.New("cmd format error")
)

/**
redis protocol references:
http://redis.io/topics/protocol
http://redis.readthedocs.org/en/latest/topic/protocol.html
1 请求协议的格式如下
*<参数个数>\r\n
$<参数1的字节长度>\r\n
<数据>\r\n
...
$<参数N的字节长度>\r\n
<数据>\r\n

2 响应格式
“+”只返回一行数据
“-”发生了错误的提示信息
“:”返回整形数值
“$”返回一组数据（也可以理解为一团，一批数据……）
“*”返回多组数据
*/
func parseCmd(reader *bufio.Reader) ([]string, error) {
	// get argument number
	argNum, err := parseCmdSize(reader, '*')
	if err != nil {
		log.Error("tcp:cmd format error when find '*' (%v)", err)
		return nil, err
	}
	if argNum < MinCmdNum || argNum > MaxCmdNum {
		log.Error("tcp:cmd argument number length error")
		return nil, ErrProtocol
	}
	args := make([]string, 0, argNum)
	for i := 0; i < argNum; i++ {
		// get argument length
		dataLength, err := parseCmdSize(reader, '$')
		if err != nil {
			log.Error("tcp:parseCmdSize(reader, '$') error(%v)", err)
			return nil, err
		}
		// get argument data
		d, err := parseCmdData(reader, dataLength)
		if err != nil {
			log.Error("tcp:parseCmdData error(%v)", err)
			return nil, err
		}
		// append args
		args = append(args, string(d))
	}
	return args, nil
}

// parseCmdSize get the request protocol cmd size.
func parseCmdSize(reader *bufio.Reader, prefix byte) (int, error) {
	// get command size
	line, err := reader.ReadBytes('\n')
	if err != nil {
		log.Error("tcp:reader.ReadBytes('\\n') error(%v)", err)
		return 0, err
	}
	lineLength := len(line)
	if lineLength < 3 || line[0] != prefix || line[lineLength-2] != '\r' {
		log.Error("tcp:\"%v\"(%d) number format error, length error or prefix error or no \\r", line, lineLength)
		return 0, ErrProtocol
	}
	// skip the \r\n
	cmdSize, err := strconv.Atoi(string(line[1 : lineLength-1]))
	if err != nil {
		log.Error("tcp:\"%v\" number parse int error(%v)", line, err)
		return 0, ErrProtocol
	}
	return cmdSize, nil
}

// parseCmdData get the sub request protocol cmd data not included \r\n.
func parseCmdData(reader *bufio.Reader, dataLength int) ([]byte, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		log.Error("tcp:reader.ReadBytes('\\n') error(%v)", err)
		return 0, err
	}
	lineLength := len(line)
	// check last \r\n
	if lineLength < 3 || lineLength != dataLength+2 || line[lineLength-1] != '\r' {
		log.Error("tcp:\"%v\"(%d) number format error, length error or no \\r", line, lineLength)
		return nil, ErrProtocol
	}
	// skip last \r\n
	return line[0:dataLength], nil
}
