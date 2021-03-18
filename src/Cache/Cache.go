package Cache

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type ChatCacheConfig struct {
	addr		string
}

type ChatCache struct {
	ChatCacheConfig
	conn 		redis.Conn
}

var ChatCacheInstance *ChatCache

func InitCache() *ChatCache {
	cacheConfig := ChatCacheConfig{
		addr: "127.0.0.1:6379",
	}
	catCache := &ChatCache{
		ChatCacheConfig:cacheConfig,
	}
	catCache.Connect()
	return catCache
}

func GetCacheDBInstance() *ChatCache {
	if ChatCacheInstance == nil {
		ChatCacheInstance = InitCache()
	}
	return ChatCacheInstance
}

func (c *ChatCache) Connect() {
	conn, err := redis.Dial("tcp", c.addr)
	if err != nil {
		fmt.Println("connect redis error!!")
		panic("redis error")
	}
	fmt.Println("connect redis succus!!!")
	c.conn = conn
}

