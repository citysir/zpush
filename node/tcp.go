// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	log "code.google.com/p/log4go"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	minHearbeatSec            = 30
	delayHeartbeatSec         = 5
	firstPacketTimedoutSecond = time.Second * 5
	Second                    = int64(time.Second)
)

// tcpBuf cache.
type TcpBufferCache struct {
	instance []chan *bufio.Reader
	round    int
}

// NewTcpBufferCache return a new TcpBufferCache.
func NewTcpBufferCache() *TcpBufferCache {
	instance := make([]chan *bufio.Reader, 0, Conf.BufioInstance)
	log.Debug("create %d read buffer instance", Conf.BufioInstance)
	for i := 0; i < Conf.BufioInstance; i++ {
		instance = append(instance, make(chan *bufio.Reader, Conf.BufioNumPerInstance))
	}
	return &TcpBufferCache{instance: instance, round: 0}
}

// Get return a chan bufio.Reader (round-robin).
func (b *TcpBufferCache) Get() chan *bufio.Reader {
	readerChan := b.instance[b.round]
	// split requets to diff buffer chan
	if b.round++; b.round == Conf.BufioInstance {
		b.round = 0
	}
	return readerChan
}

// newBufioReader get a Reader by chan, if chan empty new a Reader.
func newBufioReader(c chan *bufio.Reader, r io.Reader) *bufio.Reader {
	select {
	case p := <-c:
		p.Reset(r)
		return p
	default:
		log.Warn("tcp bufioReader cache empty")
		return bufio.NewReaderSize(r, Conf.ReadBufferSize)
	}
}

// putBufioReader pub back a Reader to chan, if chan full discard it.
func putBufioReader(c chan *bufio.Reader, r *bufio.Reader) {
	r.Reset(nil)
	select {
	case c <- r:
	default:
		log.Warn("tcp bufioReader cache full")
	}
}

// StartTCP Start tcp listen.
func BindTcpAddr(tcpAddr string) {
	go tcpListen(tcpAddr)
}

func tcpListen(bind string) {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		log.Error("net.ResolveTCPAddr(\"tcp\"), %s) error(%v)", bind, err)
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	// free the listener resource
	defer func() {
		log.Info("tcp addr: \"%s\" close", bind)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	// init reader buffer instance
	tcpBufferCache := NewTcpBufferCache()
	for {
		log.Debug("start accept")
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Error("listener.AcceptTCP() error(%v)", err)
			continue
		}
		if err = conn.SetKeepAlive(Conf.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			conn.Close()
			continue
		}
		if err = conn.SetReadBuffer(Conf.ReadBufferSize); err != nil {
			log.Error("conn.SetReadBuffer(%d) error(%v)", Conf.ReadBufferSize, err)
			conn.Close()
			continue
		}
		if err = conn.SetWriteBuffer(Conf.WriteBufferSize); err != nil {
			log.Error("conn.SetWriteBuffer(%d) error(%v)", Conf.WriteBufferSize, err)
			conn.Close()
			continue
		}
		// first packet must sent by client in limited seconds
		if err = conn.SetReadDeadline(time.Now().Add(firstPacketTimedoutSecond)); err != nil {
			log.Error("conn.SetReadDeadLine() error(%v)", err)
			conn.Close()
			continue
		}
		readerChan := tcpBufferCache.Get()
		// one connection one routine
		go handleTcpConn(conn, readerChan)
		log.Debug("accept finished")
	}
}

// hanleTCPConn handle a long live tcp connection.
func handleTcpConn(conn net.Conn, readerChan chan *bufio.Reader) {
	addr := conn.RemoteAddr().String()
	log.Debug("<%s> handleTcpConn routine start", addr)
	reader := newBufioReader(readerChan, conn)
	if args, err := parseCmd(reader); err == nil {
		// return buffer bufio.Reader
		putBufioReader(readerChan, reader)
		switch args[0] {
		case CmdSubscribe:
			subscribeTcpHandle(conn, args[1:])
		default:
			conn.Write(ParamErrorReply)
			log.Warn("<%s> unknown cmd \"%s\"", addr, args[0])
		}
	} else {
		// return buffer bufio.Reader
		putBufioReader(readerChan, reader)
		log.Error("<%s> parseCmd() error(%v)", addr, err)
	}
	// close the connection
	if err := conn.Close(); err != nil {
		log.Error("<%s> conn.Close() error(%v)", addr, err)
	}
	log.Debug("<%s> handleTcpConn routine stop", addr)
}

// subscribeTcpHandle handle the subscribers's connection.
func subscribeTcpHandle(conn net.Conn, args []string) {
	argCount := len(args)
	addr := conn.RemoteAddr().String()
	if argCount < 2 {
		conn.Write(ParamErrorReply)
		log.Error("<%s> subscriber missing argument", addr)
		return
	}
	// key, heartbeat
	key := args[0]
	if key == "" {
		conn.Write(ParamErrorReply)
		log.Warn("<%s> key param error", addr)
		return
	}
	log.Debug("match node:%s hash node:%s", Conf.ZookeeperNode, NodeRing.Hash(key))
	if Conf.ZookeeperNode != NodeRing.Hash(key) {
		conn.Write(NodeErrorReply)
		log.Warn("<%s> key node(%s) unmatch", addr, CometRing.Hash(key))
		return
	}
	heartbeatStr := args[1]
	i, err := strconv.Atoi(heartbeatStr)
	if err != nil {
		conn.Write(ParamErrorReply)
		log.Error("<%s> user_key:\"%s\" heartbeat:\"%s\" argument error (%v)", addr, key, heartbeatStr, err)
		return
	}
	if i < minHearbeatSec {
		conn.Write(ParamErrorReply)
		log.Warn("<%s> user_key:\"%s\" heartbeat argument error, less than %d", addr, key, minHearbeatSec)
		return
	}
	heartbeat := i + delayHeartbeatSec
	token := ""
	if argCount > 2 {
		token = args[2]
	}
	version := ""
	if argCount > 3 {
		version = args[3]
	}
	log.Info("<%s> subscribe to key = %s, heartbeat = %d, token = %s, version = %s", addr, key, heartbeat, token, version)
	// fetch subscriber from the channel
	c, err := ChannelManager.Get(key, true)
	if err != nil {
		log.Warn("<%s> user_key:\"%s\" can't get a channel (%s)", addr, key, err)
		conn.Write(ChannelReply)
		return
	}
	// auth token
	if ok := c.AuthToken(key, token); !ok {
		conn.Write(AuthReply)
		log.Error("<%s> user_key:\"%s\" auth token \"%s\" failed", addr, key, token)
		return
	}
	// add a conn to the channel
	connElem, err := c.AddConn(key, &Connection{Conn: conn, Version: version})
	if err != nil {
		log.Error("<%s> user_key:\"%s\" add conn error(%v)", addr, key, err)
		return
	}
	// blocking wait client heartbeat
	cmd := []byte{' '}
	// reply := make([]byte, HeartbeatLen)
	begin := time.Now().UnixNano()
	end := begin + Second
	for {
		// more then 1 sec, reset the timer
		if end-begin >= Second {
			if err = conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(heartbeat))); err != nil {
				log.Error("<%s> user_key:\"%s\" conn.SetReadDeadLine() error(%v)", addr, key, err)
				break
			}
			begin = end
		}
		if _, err = conn.Read(cmd); err != nil {
			if err != io.EOF {
				log.Warn("<%s> user_key:\"%s\" conn.Read() failed, read heartbeat timedout error(%v)", addr, key, err)
			} else {
				// client connection close
				log.Warn("<%s> user_key:\"%s\" client connection close error(%v)", addr, key, err)
			}
			break
		}
		if string(cmd) == CmdHeartBeat {
			if _, err = conn.Write(HeartbeatReply); err != nil {
				log.Error("<%s> user_key:\"%s\" conn.Write() failed, write heartbeat to client error(%v)", addr, key, err)
				break
			}
			log.Debug("<%s> user_key:\"%s\" receive heartbeat (%s)", addr, key, cmd)
		} else {
			log.Warn("<%s> user_key:\"%s\" unknown heartbeat protocol (%s)", addr, key, cmd)
			break
		}
		end = time.Now().UnixNano()
	}
	// remove exists conn
	if err := c.RemoveConn(key, connElem); err != nil {
		log.Error("<%s> user_key:\"%s\" remove conn error(%v)", addr, key, err)
	}
	return
}
