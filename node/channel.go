package main

import (
	log "code.google.com/p/log4go"
	"errors"
	"fmt"
	"github.com/citysir/golib/hash"
	"net"
	"sync"
)

var (
	ErrChannelNotExist = errors.New("Channel not exist")
	ErrConnProto       = errors.New("Unknown connection protocol")
	UserChannel        *ChannelHome
	CometRing          *ketama.HashRing
)

// The subscriber interface.
type Channel interface {
	PushMsg(key string, m *myrpc.Message, expire uint) error
	AddConn(key string, conn *Connection) (*hlist.Element, error)
	RemoveConn(key string, e *hlist.Element) error
	Close() error
}

// Connection
type Connection struct {
	Conn net.Conn
	Msgs chan []byte
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
		}
	}()
}

// Write different message to client by different protocol
func (c *Connection) Write(key string, msg []byte) {
	select {
	case c.Buf <- msg:
	default:
		c.Conn.Close()
		log.Warn("user_key: \"%s\" discard message: \"%s\" and close connection", key, string(msg))
	}
}

// Channel bucket.
type ChannelBucket struct {
	Channels map[string]Channel
	mutex    *sync.Mutex
}

// ChannelBuckets.
type ChannelHome []*ChannelBucket

// Lock lock the bucket mutex.
func (c *ChannelBucket) Lock() {
	c.mutex.Lock()
}

// Unlock unlock the bucket mutex.
func (c *ChannelBucket) Unlock() {
	c.mutex.Unlock()
}

func NewChannelHome() *ChannelHome {
	channelHome := new(ChannelHome)
	// split hashmap to many bucket
	log.Debug("create %d ChannelHome", Conf.ChannelBucket)
	for i := 0; i < Conf.ChannelBucket; i++ {
		bucket := &ChannelBucket{
			Channels: map[string]Channel{},
			mutex:    &sync.Mutex{},
		}
		channelHome = append(channelHome, bucket)
	}
	return l
}

// Count get the bucket total channel count.
func (this *ChannelHome) Count() int {
	c := 0
	for i := 0; i < Conf.ChannelBucket; i++ {
		c += len(this.Channels)
	}
	return c
}

// bucket return a channelBucket use murmurhash3.
func (this *ChannelHome) bucket(key string) *ChannelBucket {
	h := hash.NewMurmur3C()
	h.Write([]byte(key))
	idx := uint(h.Sum32()) & uint(Conf.ChannelBucket-1)
	log.Debug("user_key:\"%s\" hit channel bucket index:%d", key, idx)
	return l.Channels[idx]
}

// bucketIdx return a channelBucket index.
func (this *ChannelHome) BucketIdx(key *string) uint {
	h := hash.NewMurmur3C()
	h.Write([]byte(*key))
	return uint(h.Sum32()) & uint(Conf.ChannelBucket-1)
}

// New create a user channel.
func (this *ChannelHome) New(key string) (Channel, error) {
	// get a channel bucket
	b := this.bucket(key)
	b.Lock()
	if c, ok := b.Data[key]; ok {
		b.Unlock()
		return c, nil
	} else {
		c = NewSeqChannel()
		b.Data[key] = c
		b.Unlock()
		return c, nil
	}
}

// Get a user channel from ChannelHome.
func (this *ChannelHome) Get(key string, newOne bool) (Channel, error) {
	// get a channel bucket
	b := this.bucket(key)
	b.Lock()
	if c, ok := b.Data[key]; !ok {
		if !Conf.Auth && newOne {
			c = NewSeqChannel()
			b.Data[key] = c
			b.Unlock()
			return c, nil
		} else {
			b.Unlock()
			log.Warn("user_key:\"%s\" channle not exists", key)
			return nil, ErrChannelNotExist
		}
	} else {
		b.Unlock()
		return c, nil
	}
}

// Delete a user channel from ChannleList.
func (l *ChannelHome) Delete(key string) (Channel, error) {
	// get a channel bucket
	b := l.bucket(key)
	b.Lock()
	if c, ok := b.Data[key]; !ok {
		b.Unlock()
		log.Warn("user_key:\"%s\" delete channle not exists", key)
		return nil, ErrChannelNotExist
	} else {
		delete(b.Data, key)
		b.Unlock()
		ChStat.IncrDelete()
		log.Info("user_key:\"%s\" delete channel", key)
		return c, nil
	}
}

// Close close all channel.
func (l *ChannelHome) Close() {
	log.Info("channel close")
	chs := make([]Channel, 0, l.Count())
	for _, c := range l.Channels {
		c.Lock()
		for _, c := range c.Data {
			chs = append(chs, c)
		}
		c.Unlock()
	}
	// close all channels
	for _, c := range chs {
		if err := c.Close(); err != nil {
			log.Error("c.Close() error(%v)", err)
		}
	}
}

// Migrate migrate portion of connections which don`t belong to this Comet
func (l *ChannelHome) Migrate() {
	// init ketama
	ring := ketama.NewRing(Conf.KetamaBase)
	for node, weight := range nodeWeightMap {
		ring.AddNode(node, weight)
	}
	ring.Bake()
	CometRing = ring

	// get all the channel lock
	channels := []Channel{}
	for i, c := range l.Channels {
		c.Lock()
		for k, v := range c.Data {
			hn := ring.Hash(k)
			if hn != Conf.ZookeeperCometNode {
				channels = append(channels, v)
				delete(c.Data, k)
				log.Debug("migrate delete channel key \"%s\"", k)
			}
		}
		c.Unlock()
		log.Debug("migrate channel bucket:%d finished", i)
	}
	// close all the migrate channels
	log.Info("close all the migrate channels")
	for _, channel := range channels {
		if err := channel.Close(); err != nil {
			log.Error("channel.Close() error(%v)", err)
			continue
		}
	}
	log.Info("close all the migrate channels finished")
}
