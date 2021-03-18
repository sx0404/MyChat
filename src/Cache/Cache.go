package Cache

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type CacheConfig struct {
	addr		string
}

type ChatCache struct {
	CacheConfig
	conn 		redis.Conn
}

var CacheInstance *ChatCache

func InitCache() *ChatCache {
	cacheConfig := CacheConfig{
		addr: "127.0.0.1:6379",
	}
	catCache := &ChatCache{
		CacheConfig:cacheConfig,
	}
	catCache.Connect()
	return catCache
}

func GetCacheDBInstance() *ChatCache {
	if CacheInstance == nil {
		CacheInstance = InitCache()
	}
	return CacheInstance
}

func (this *ChatCache) Connect() {
	conn, err := redis.Dial("tcp",this.addr)
	if err != nil {
		fmt.Println("connect redis error!!")
		panic("redis error")
	}
	fmt.Println("connect redis succus!!!")
	this.conn = conn
}

