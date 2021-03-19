package main

import (
	"fmt"
	"test/src/AStar"
	"test/src/BookDemo"
	"test/src/Cache"
	"test/src/Client"
	"test/src/Common"
	db "test/src/DB"
	"test/src/Server"
)

func main() {
	StartMod := Common.KeyInput()
	if StartMod == "s" {
		fmt.Println("server mode begin!!")
		Common.InitLog()
		Common.GetProtoDealInstance() //初始化编解码
		Cache.GetCacheDBInstance()
		db.GetDBInstance()
		Server.GetOnlineRoleInstance()
		Server.InitGateWay()
	} else if StartMod == "c" { //作为客户端启动
		fmt.Println("client mode begin!!")
		Client.InitClient()
	} else if StartMod == "d" { //测试书上示例的代码
		BookDemo.DoPanic()
	} else {
		AStar.CalcAStarTime(1)
	}
}
