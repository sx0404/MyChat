package Cache

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"reflect"
	"test/src/Common"
	"test/src/Formation"
)

//加载信息到玩家内存的函数
func SetUser2(userID uint64, i interface{}) bool {
	this := GetCacheDBInstance()
	j,err := json.Marshal(i)
	if err != nil {
		fmt.Println("json marshal error",i)
		return false
	}
	tag := reflect.TypeOf(i).Name()
	_,err = this.conn.Do("SET", tag + Common.ToString(int64(userID)), j)
	if err != nil {
		fmt.Println("SetUser2 error",i,err)
		return false
	}
	return true
}

func GetRole2(userID uint64,i interface{})  {
	this := GetCacheDBInstance()
	data,err := redis.Bytes(this.conn.Do("GET", reflect.TypeOf(i).Name() + Common.ToString(int64(userID))))
	if err != nil {
		fmt.Println("GetRole2 error",i)
	}
	err = json.Unmarshal(data, i)
	if err != nil {
		fmt.Println("GetRole2 json Unmarshal",i)
	}
	return
}

func GetRoleInfo(userID uint64) *Formation.RoleInfo {
	p := new(Formation.RoleInfo)
	GetRole2(userID,p)
	return p
}

func SetRoleInfo(userID uint64,roleInfo Formation.RoleInfo) bool {
	return SetUser2(userID,roleInfo)
}

func GetRoleFriend(userID uint64) *Formation.RoleFriendInfo {
	p := new(Formation.RoleFriendInfo)
	GetRole2(userID,p)
	return p
}

func SetRoleFriend(userID uint64,roleFriend Formation.RoleFriendInfo) bool {
	return SetUser2(userID,roleFriend)
}

func GetRoleMoney(userID uint64) *Formation.RoleMoney {
	p := new(Formation.RoleMoney)
	GetRole2(userID,p)
	return p
}

func SetRoleMoney(userID uint64,roleMoney Formation.RoleMoney) bool {
	return SetUser2(userID,roleMoney)
}