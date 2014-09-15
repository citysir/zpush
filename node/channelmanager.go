package main

import (
	log "code.google.com/p/log4go"
	"errors"
	"fmt"
	"github.com/citysir/golib/hash"
	"github.com/citysir/zpush/proto"
	"net"
	"sync"
)

var (
	ErrChannelNotExist = errors.New("Channel not exist")
	ErrConnProto       = errors.New("Unknown connection protocol")
	ChannelManager     *ChannelManagerStruct
	NodeRing           *hash.HashRing
)

// The subscriber interface.
type Channel interface {
	PushMsg(key string, m *proto.Message, expire uint) error
	AddConn(key string, conn *Connection) (*HlistNode, error)
	RemoveConn(key string, e *HlistNode) error
	Close() error
}

// Connection
type Connection struct {
	Conn    net.Conn
	Version string
	Msgs    chan []byte
}

// HandleWrite start a goroutine get msg from chan, then send to the conn.
func (c *Connection) HandleWrite(key string) {
	go func() {
		var (
			n   int
			err error
		)
		log.Debug("user_key: \"%s\" HandleWrite goroutine start", key)
		for {
			msg, ok := <-c.Msgs
			if !ok {
				log.Debug("user_key: \"%s\" HandleWrite goroutine stop", key)
				return
			}
			msg = []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), string(msg)))
			n, err = c.Conn.Write(msg)
			// update stat
			if err != nil {
				log.Error("user_key: \"%s\" conn.Write() error(%v)", key, err)
			} else {
				log.Debug("user_key: \"%s\" write \r\n========%s(%d)========", key, string(msg), n)
			}
		}
	}()
}

// Write different message to client by different protocol
func (c *Connection) Write(key string, msg []byte) {
	select {
	case c.Msgs <- msg:
	default:
		c.Conn.Close()
		log.Warn("user_key: \"%s\" discard message: \"%s\" and close connection", key, string(msg))
	}
}

// Channel bucket.
type ChannelBucket struct {
	channels map[string]Channel
	mutex    *sync.Mutex
}

// ChannelBuckets.
type ChannelManagerStruct struct {
	buckets []*ChannelBucket
}

// Lock lock the bucket mutex.
func (c *ChannelBucket) Lock() {
	c.mutex.Lock()
}

// Unlock unlock the bucket mutex.
func (c *ChannelBucket) Unlock() {
	c.mutex.Unlock()
}

func NewChannelManager() *ChannelManagerStruct {
	channelManager := new(ChannelManagerStruct)
	channelManager.buckets = []*ChannelBucket{}
	// split hashmap to many bucket
	log.Debug("create %d ChannelManagerStruct", Conf.ChannelBucket)
	for i := 0; i < Conf.ChannelBucket; i++ {
		bucket := &ChannelBucket{
			channels: map[string]Channel{},
			mutex:    &sync.Mutex{},
		}
		channelManager.buckets = append(channelManager.buckets, bucket)
	}
	return channelManager
}

// Count get the bucket total channel count.
func (this *ChannelManagerStruct) Count() int {
	c := 0
	for i := 0; i < Conf.ChannelBucket; i++ {
		c += len(this.buckets[i].channels)
	}
	return c
}

// bucket return a channelBucket use murmurhash3.
func (this *ChannelManagerStruct) bucket(key string) *ChannelBucket {
	h := hash.NewMurmur3C()
	h.Write([]byte(key))
	idx := uint(h.Sum32()) & uint(Conf.ChannelBucket-1)
	log.Debug("user_key:\"%s\" hit channel bucket index:%d", key, idx)
	return this.buckets[idx]
}

// bucketIdx return a channelBucket index.
func (this *ChannelManagerStruct) BucketIdx(key *string) uint {
	h := hash.NewMurmur3C()
	h.Write([]byte(*key))
	return uint(h.Sum32()) & uint(Conf.ChannelBucket-1)
}

// New create a user channel.
func (this *ChannelManagerStruct) New(key string) (Channel, error) {
	// get a channel bucket
	b := this.bucket(key)
	b.Lock()
	if c, ok := b.channels[key]; ok {
		b.Unlock()
		return c, nil
	} else {
		c = NewChannel()
		b.channels[key] = c
		b.Unlock()
		return c, nil
	}
}

// Get a user channel from ChannelManagerStruct.
// func (this *ChannelManagerStruct) Get(key string, newOne bool) (Channel, error) {
// 	// get a channel bucket
// 	b := this.bucket(key)
// 	b.Lock()
// 	if c, ok := b.Channels[key]; !ok {
// 		if !Conf.Auth && newOne {
// 			c = NewSeqChannel()
// 			b.Channels[key] = c
// 			b.Unlock()
// 			return c, nil
// 		} else {
// 			b.Unlock()
// 			log.Warn("user_key:\"%s\" channle not exists", key)
// 			return nil, ErrChannelNotExist
// 		}
// 	} else {
// 		b.Unlock()
// 		return c, nil
// 	}
// }

// // Delete a user channel from ChannleList.
// func (l *ChannelManagerStruct) Delete(key string) (Channel, error) {
// 	// get a channel bucket
// 	b := l.bucket(key)
// 	b.Lock()
// 	if c, ok := b.Channels[key]; !ok {
// 		b.Unlock()
// 		log.Warn("user_key:\"%s\" delete channle not exists", key)
// 		return nil, ErrChannelNotExist
// 	} else {
// 		delete(b.Channels, key)
// 		b.Unlock()
// 		ChStat.IncrDelete()
// 		log.Info("user_key:\"%s\" delete channel", key)
// 		return c, nil
// 	}
// }

// // Close close all channel.
// func (l *ChannelManagerStruct) Close() {
// 	log.Info("channel close")
// 	chs := make([]Channel, 0, l.Count())
// 	for _, c := range l.Channels {
// 		c.Lock()
// 		for _, c := range c.Channels {
// 			chs = append(chs, c)
// 		}
// 		c.Unlock()
// 	}
// 	// close all channels
// 	for _, c := range chs {
// 		if err := c.Close(); err != nil {
// 			log.Error("c.Close() error(%v)", err)
// 		}
// 	}
// }

// Migrate migrate portion of connections which don`t belong to this Comet
// func (l *ChannelManagerStruct) Migrate() {
// 	// init ketama
// 	ring := ketama.NewRing(Conf.KetamaBase)
// 	for node, weight := range nodeWeightMap {
// 		ring.AddNode(node, weight)
// 	}
// 	ring.Bake()
// 	NodeRing = ring

// 	// get all the channel lock
// 	channels := []Channel{}
// 	for i, c := range l.Channels {
// 		c.Lock()
// 		for k, v := range c.Channels {
// 			hn := ring.Hash(k)
// 			if hn != Conf.ZookeeperNode {
// 				channels = append(channels, v)
// 				delete(c.Channels, k)
// 				log.Debug("migrate delete channel key \"%s\"", k)
// 			}
// 		}
// 		c.Unlock()
// 		log.Debug("migrate channel bucket:%d finished", i)
// 	}
// 	// close all the migrate channels
// 	log.Info("close all the migrate channels")
// 	for _, channel := range channels {
// 		if err := channel.Close(); err != nil {
// 			log.Error("channel.Close() error(%v)", err)
// 			continue
// 		}
// 	}
// 	log.Info("close all the migrate channels finished")
// }
