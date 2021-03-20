package Server

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"net"
	"test/src/Common"
	db "test/src/DB"
	"test/src/ErrCode"
	"test/src/Formation"
	ChatMsg "test/src/proto"
)

type SelfRouteMsg struct {
	Code   string //操作码
	Params []interface{}
}

type GateWayAgent struct {
	QueProcessor
	conn net.Conn
	*Common.ProtoDeal
	userID       uint64
	roleLogic    *RoleProcessor
	buff         []byte
	buffIndex    uint16
	lastHearTIme int64
}

type MsgRaw struct {
	msgID      uint16
	msgRawData []byte
}

func InitGateAgent(conn net.Conn) {
	queProcessor := InitQueProcessor("")
	protoDeal := Common.GetProtoDealInstance()
	gateWayAgent := GateWayAgent{
		QueProcessor: queProcessor,
		conn:         conn,
		ProtoDeal:    protoDeal,
		userID:       0}
	gateWayAgent.RunGateWay()
}

//启动一个agent来处理
func (a *GateWayAgent) RunGateWay() {
	defer a.conn.Close()
	a.buff = make([]byte, 1024)
	a.buffIndex = 0
	a.LoopDoNetData()
}

func (a *GateWayAgent) LoopDoNetData() {
	data := make([]byte, 1024)
	for {
		dataLen, err := a.conn.Read(data)
		if dataLen == 0 {
			continue
		}
		if err != nil {
			fmt.Println("get err wrong", err)
			break
		}
		a.buff = append(a.buff[:a.buffIndex], data...)
		a.buffIndex += uint16(dataLen)
		var msgLen uint16 //消息长度防止沾包
		msgLen = binary.BigEndian.Uint16(a.buff)
		for {
			if uint16(dataLen) < msgLen {
				break //等待后续的包
			}

			// id
			var pdNameLen uint16
			pdNameLen = binary.BigEndian.Uint16(a.buff[2:4]) //前两位作为proto名称的长度
			var msgName string
			msgName = string(a.buff[4 : 4+pdNameLen])
			msg := a.PdFactory(msgName)
			err = proto.Unmarshal(a.buff[4+pdNameLen:msgLen], msg)
			if err != nil {
				fmt.Println("proto unmarshal error!!!!")
				a.DoExit() //解码出错关闭网关
			}
			a.buffIndex -= msgLen
			a.buff = a.buff[msgLen:dataLen]
			if msgName == "" {
				Common.Debug("unmarshal message error: ")
				break
			}
			fmt.Println("msg ok,msg", msgName)
			//直接执行函数
			a.HandleMsg(msgName, msg)
			if a.buffIndex < 2 {
				break
			}
		}
	}
}

func (a *GateWayAgent) RegisterSelfAll() {
	a.RegisterSelfInfo()
}

func (a *GateWayAgent) RegisterSelfInfo() {
	a.RegisterInfo(&ChatMsg.CsLogin{}, a.CsLogin)
}

func (a *GateWayAgent) SendSock(i proto.Message) {
	C := a.Marshal(i)
	a.conn.Write(C)
}

func (a *GateWayAgent) GenAgentName() string {
	return "GateWay" + Common.ToString(int64((a.userID)))
}

func (a *GateWayAgent) CsHeart(i interface{}) {
	a.lastHearTIme = Common.GetNow()
}

func (a *GateWayAgent) CheckHearTime() {
	Now := Common.GetNow()
	//TODO 客户端暂时不发心跳，心跳短线后面再测
	if Now-a.lastHearTIme > math.MaxInt64 {
		a.Exit()
	}
}

func (a *GateWayAgent) Exit() {
	a.conn.Close()
	//logic一并退出
}

func (a *GateWayAgent) CsLogin(i interface{}) {
	//检查登陆信息,login放在网关.登陆没有问题再启动逻辑
	msg, ok := i.(*ChatMsg.CsLogin)
	if !ok {
		fmt.Println("CsLogin struct error")
		return
	}
	roleInfo := db.GetUser(msg.UserName)
	if roleInfo.UserName == "" {
		//注册一个玩家信息
		roleInfo = a.CreateRole(msg.UserName, msg.PassWord)
	} else if roleInfo.Password != msg.PassWord {
		a.SendSock(&ChatMsg.ScLogin{ErrCode: ErrCode.LoginPassWord})
		return
	}
	//登陆成功
	a.userID = roleInfo.UserID
	//给网关关连一个名字并注册
	a.Name = a.GenAgentName()
	//检测在线的逻辑进程信息
	OnlineRoleInstance := GetOnlineRoleInstance()
	//网关这里只能用,
	roleLogicProcessor := OnlineRoleInstance.FindWithID(a.userID)
	//不存在逻辑进程则立即启动
	if roleLogicProcessor == nil {
		//启动一个逻辑进程
		roleProcessor := InitRoleProcessor(a, msg.UserName)
		a.roleLogic = &roleProcessor
		a.roleLogic.RunRoleProcessor()
	} else {
		a.roleLogic = roleLogicProcessor
		//逻辑正关联一个网管则直接提离
		if a.roleLogic.gatewayAgent != nil {
			//直接断掉前网关的socket来关闭网关
			a.roleLogic.gatewayAgent.conn.Close()
		}
	}
	a.AddProcToM()
	//把当前逻辑的网管替换成自己
	a.roleLogic.gatewayAgent = a
	a.SendSock(&ChatMsg.ScLogin{ErrCode: ErrCode.OK})
}

func (a *GateWayAgent) GetUserID() uint64 {
	return a.userID
}

const CsLoginStr = "CsLogin"
const CsHeartStr = "CsHeart"

func (a *GateWayAgent) HandleMsg(msgName string, msg interface{}) {
	switch {
	case msgName == CsLoginStr:
		a.CsLogin(msg)
	case msgName == CsHeartStr:
		a.CsHeart(msg)
	default:
		roleLogic := a.roleLogic
		a.Send(&(roleLogic.QueProcessor), msg)
	}
}

func (a *GateWayAgent) CreateRole(userName string, passWord string) Formation.RoleInfo {
	roleInfo := Formation.RoleInfo{
		UserID:   Common.GenUserID(),
		UserName: userName,
		Password: passWord,
	}
	db.InserRoleInfo(roleInfo)
	return roleInfo
}
