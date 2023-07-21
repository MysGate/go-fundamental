package pubsub

import (
	"fmt"

	"github.com/MysGate/go-fundamental/util"
	"github.com/cskr/pubsub"
)

const (
	capacity int = 100000
)

var manager Manager

type ManagerImpl struct {
	pubSub *pubsub.PubSub
}

type Manager interface {
	ShutDown()
	Close(string)
	AddSubscribe(string, chan interface{})
	UnSubscribe(string, chan interface{})
	TryPublish(string, interface{})
	TrySendOnce(ch chan interface{}, data interface{})
}

func InitSubscribeManager() Manager {
	manager = &ManagerImpl{
		pubSub: pubsub.New(capacity),
	}
	return manager
}

func GetSubscribeManager() Manager {
	return manager
}

func (c *ManagerImpl) ShutDown() {
	c.pubSub.Shutdown()
}

func (c *ManagerImpl) Close(key string) {
	c.pubSub.Close(key)
}

func (c *ManagerImpl) AddSubscribe(key string, ch chan interface{}) {
	c.pubSub.AddSub(ch, key)
}

func (c *ManagerImpl) UnSubscribe(key string, ch chan interface{}) {
	c.pubSub.Unsub(ch, key)
}

func (c *ManagerImpl) TryPublish(key string, data interface{}) {
	c.pubSub.TryPub(data, key)
}

func (c *ManagerImpl) TrySendOnce(ch chan interface{}, data interface{}) {
	select {
	case ch <- data:
	default:
		util.Logger().Error(fmt.Sprintf("chan blocking, drop data %v", data))
	}
}
