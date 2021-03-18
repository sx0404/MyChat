package Client

import (
	"encoding/binary"
	"fmt"
	"net"
	"test/src/Common"
	"test/src/ErrCode"
	ChatMsg "test/src/proto"
)

type ChatClient struct{
	UserName 			string
	PassWorld 			string
	conn 				net.Conn
	serverAddr 			string
	*Common.ProtoDeal
	buff 				[]byte
	buffIndex 			uint16
	handle 				map[string]func(interface{})
	target				string
}

func InitClient() {
	chatClient := ChatClient{
		serverAddr: "127.0.0.1:12588",
		handle: make(map[string]func(interface{})),
	}
	chatClient.ProtoDeal = Common.GetProtoDealInstance()
	chatClient.connect()
	chatClient.RegisterAll()
	chatClient.RegistSelfAll()
	chatClient.BehaviorTree()
	chatClient.loop()
}

func (this *ChatClient) loop() {
	for {
		data := make([]byte, 1024)
		dataLen, err := this.conn.Read(data)
		if err != nil {
			fmt.Println("get err wrong", err)
			this.DoExit()
		}
		this.buff = append(this.buff[:this.buffIndex], data...)
		this.buffIndex += uint16(dataLen)
		for {
			var msgLen uint16 //消息长度防止沾包
			msgLen = binary.BigEndian.Uint16(this.buff)
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
			//fmt.Println("msg ok,msg", msgName)
			//直接执行函数
			this.HandleMsg(msg)
			if this.buffIndex < 2 {
				break
			}
		}
	}
}

func (this *ChatClient) DoExit() {
	this.conn.Close()
}

func (this *ChatClient) connect() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", this.serverAddr)
	if err != nil {
		fmt.Println( "addr err", err.Error())
		panic("addr error")
	}
	this.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("tcp error", err.Error())
		panic("tcp error")
	}

	fmt.Println("connection success")
}

func (this *ChatClient) SendSock(i interface{}) {
	data := this.Marshal(i)
	this.conn.Write(data)
}

func (this *ChatClient) BehaviorTree() {
	this.DoLogin()
}

func (this *ChatClient) DoLogin() {
	fmt.Println("input UserName:")
	this.UserName = Common.KeyInput()
	fmt.Println("input PassWord:")
	this.PassWorld = Common.KeyInput()

	this.SendSock(&ChatMsg.CsLogin{UserName: this.UserName,PassWord: this.PassWorld})
}

func (this *ChatClient) CoutOnDesk(str string) {
	fmt.Println(str)
}

func (this *ChatClient) HandleMsg(msg interface{}) {
	msgName := Common.GetStructName(msg)
	f := this.handle[msgName]
	if f == nil {
		fmt.Println("HandleMsg wrong msg",msg)
	}
	f(msg)
}

func (this *ChatClient) ChooseTarget() {
	this.CoutOnDesk("input chat target:")
	targetName := Common.KeyInput()
	this.SendSock(&ChatMsg.CsChatTarget{UserName: targetName})
	this.target = targetName
}

func (this *ChatClient) Regist(msg interface{},f func(interface{})) {
	this.handle[Common.GetStructName(msg)] = f
}

func (this *ChatClient) RegistSelfAll() {
	this.Regist(ChatMsg.ScLogin{},this.ScLogin)
	this.Regist(ChatMsg.ScChatTarget{},this.ScChatTarget)
	this.Regist(ChatMsg.ScChat{},this.ScChat)
	this.Regist(ChatMsg.ScChatFrom{},this.ScChatFrom)
}

func (this *ChatClient) ScLogin(i interface{}) {
	msg := i.(*ChatMsg.ScLogin)
	switch msg.ErrCode {
	case ErrCode.OK:
		this.CoutOnDesk("login ok")
		this.ChooseTarget()
	case ErrCode.LoginPassWord:
		this.CoutOnDesk("login password error,login again")
		this.DoLogin()
	default:
		fmt.Println("ScLogin unknown errCode: ", msg.ErrCode)
		this.DoLogin()
	}
}

func (this *ChatClient) ScChatTarget(i interface{}) {
	msg := i.(*ChatMsg.ScChatTarget)
	switch msg.ErrCode {
	case ErrCode.UserNotExist:
		fmt.Println("chat target user not exist")
		this.ChooseTarget()
	case ErrCode.RoleOffline:
		fmt.Println("target is offline,send content will be shown when he online")
		this.DoChat()
	case ErrCode.OK:
		fmt.Println("target is online")
		this.DoChat()
	default :
		fmt.Println("ScChatTarget unkwnow errCode",msg.ErrCode)
	}
}

func (this *ChatClient) DoChat() {
	fmt.Println("input content will send to ",this.target)
	content := Common.KeyInput()
	this.SendSock(&ChatMsg.CsChat{Content: content})
}

func (this *ChatClient) ScChat(i interface{}) {
	msg := i.(*ChatMsg.ScChat)
	switch msg.ErrCode {
	case ErrCode.ChatTargetNotSet:
		fmt.Println("not set chat target")
		this.ChooseTarget()
	case ErrCode.RoleOffline:
		fmt.Println("chat ok,but user offline:",this.target)
		this.DoChat()
	case ErrCode.OK:
		fmt.Println("chat ok,user online:",this.target)
		go this.DoChat()
	}
}

func (this *ChatClient) ScChatFrom(i interface{}) {
	msg := i.(*ChatMsg.ScChatFrom)
	fmt.Println("From ",msg.FromName,":",msg.Content)
}