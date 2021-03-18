package Server

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"test/src/Common"
	db "test/src/DB"
	"test/src/ErrCode"
	"test/src/Formation"
	ChatMsg "test/src/proto"
)

type SelfRouteMsg struct{
	Code				string		//操作码
	Params				[]interface{}
}

type GateWayAgent struct {
	QueProcessor
	conn				net.Conn
	*Common.ProtoDeal
	userID				uint64
	roleLogic			*RoleProcessor
	buff				[]byte
	buffIndex			uint16
	lastHearTIme		int64
}

type MsgRaw struct {
	msgID      uint16
	msgRawData []byte
}

func InitGateAgent(conn net.Conn) {
	queProcessor := InitQueProcessor("")
	protoDeal := Common.GetProtoDealInstance()
	gateWayAgent := GateWayAgent{
		QueProcessor:queProcessor,
		conn: conn,
		ProtoDeal : protoDeal,
		userID: 0}
	gateWayAgent.RegisterAll()
	gateWayAgent.RunGateWay()
}

//启动一个agent来处理
func (this *GateWayAgent) RunGateWay() {
	defer this.conn.Close()
	this.buff = make([]byte, 1024)
	this.buffIndex  = 0
	this.LoopDoNetData()
}

func (this *GateWayAgent) LoopDoNetData() {
	for {
		data := make([]byte, 1024)
		dataLen, err := this.conn.Read(data)
		if err != nil {
			fmt.Println("get err wrong", err)
			break
		}
		this.buff = append(this.buff[:this.buffIndex], data...)
		this.buffIndex += uint16(dataLen)
		var msgLen uint16 //消息长度防止沾包
		msgLen = binary.BigEndian.Uint16(this.buff)
		for {
			if uint16(dataLen) < msgLen {
				break //等待后续的包
			}
			msg, msgName := this.Unmarshal(this.buff[2:msgLen])
			this.buffIndex -= msgLen
			this.buff = this.buff[msgLen:dataLen]
			if msgName == "" {
				Common.Debug("unmarshal message error: ")
				break
			}
			fmt.Println("msg ok,msg", msgName)
			//直接执行函数
			this.HandleMsg(msgName, msg)
			if this.buffIndex < 2 {
				break
			}
		}
	}
}

func (this *GateWayAgent) RegistSelfAll() {
	this.RegisterSelfInfo()
}

func (this *GateWayAgent) RegisterSelfInfo() {
	this.RegisterInfo(&ChatMsg.CsLogin{},this.CsLogin)
	this.RegisterInfo(&InfoRoleLogicSendSock{},this.SendSock)
}

func (this *GateWayAgent) SendSock(i interface{}) {
	C := this.Marshal(i)
	this.conn.Write(C)
}

func (this *GateWayAgent) GenAgentName() string{
	return "GateWay" + Common.ToString(int64((this.userID)))
}

func (this *GateWayAgent) CsHeart(i interface{}) {
	this.lastHearTIme = Common.GetNow()
}

func (this *GateWayAgent) CheckHearTime() {
	Now := Common.GetNow()
	//TODO 客户端暂时不发心跳，心跳短线后面再测
	if Now - this.lastHearTIme > math.MaxInt64 {
		this.Exit()
	}
}

func (this *GateWayAgent) Exit() {
	this.conn.Close()
}

func (this *GateWayAgent) CsLogin(i interface{}) {
	//检查登陆信息,login放在网关.登陆没有问题再启动逻辑
	msg,ok := i.(*ChatMsg.CsLogin)
	if !ok {
		fmt.Println("CsLogin struct error")
		return
	}
	roleInfo := db.GetUser(msg.UserName)
	if roleInfo.UserName == "" {
		//注册一个玩家信息
		roleInfo = this.CreateRole(msg.UserName,msg.PassWord)
	} else if roleInfo.Password != msg.PassWord {
		this.SendSock(&ChatMsg.ScLogin{ErrCode: ErrCode.LoginPassWord})
		return
	}
	//登陆成功
	this.userID = roleInfo.UserID
	//给网关关连一个名字并注册
	this.Name = this.GenAgentName()
	//检测在线的逻辑进程信息
	OnlineRoleInstance := GetOnlineRoleInstance()
	//网关这里只能用,
	roleLogicProcessor := OnlineRoleInstance.FindWithID(this.userID)
	//不存在逻辑进程则立即启动
	if roleLogicProcessor == nil {
		//启动一个逻辑进程
		roleProcessor := InitRoleProcessor(this,msg.UserName)
		this.roleLogic = roleProcessor
		this.roleLogic.RunRoleProcessor()
	}else{
		this.roleLogic = roleLogicProcessor
		//逻辑正关联一个网管则直接提离
		if this.roleLogic.gatewayAgent != nil {
			//直接断掉前网关的socket来关闭网关
			this.roleLogic.gatewayAgent.conn.Close()
		}
	}
	this.AddProcToM()
	//把当前逻辑的网管替换成自己
	this.roleLogic.gatewayAgent = this
	this.SendSock(&ChatMsg.ScLogin{ErrCode: ErrCode.OK})
}

func (this *GateWayAgent) GetUserID() uint64 {
	return this.userID
}

func (this *GateWayAgent) HandleMsg(msgName string,msg interface{}) {
	switch {
	case msgName == "CsLogin":
		this.CsLogin(msg)
	case msgName == "CsHeart" :
		this.CsHeart(msg)
	default:
		roleLogic := this.roleLogic
		this.Send(&(roleLogic.QueProcessor),msg)
	}
}

func (this *GateWayAgent) CreateRole(userName string,passWord string) Formation.RoleInfo {
	roleInfo := Formation.RoleInfo{
		UserID: Common.GenUserID(),
		UserName: userName,
		Password: passWord,
	}
	db.InserRoleInfo(roleInfo)
	return roleInfo
}