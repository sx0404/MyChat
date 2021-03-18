package Server

import (
	"net"
	"test/src/Common"
)

type GateWay struct {
	maxConnectNum int32
	TCPAddr string
}

func InitGateWay() {
	gateWay := GateWay{100,"0.0.0.0:12588"}
	ln,err := net.Listen("tcp",gateWay.TCPAddr)
	if err != nil {
		panic("err tcp liston")
		Common.Error("tcp list wrong,check your port set!!!!")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			Common.Info("get client connection error: ", err.Error())
		}
		go InitGateAgent(conn)
	}
}