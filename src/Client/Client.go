package Client

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
	"test/src/Common"
	"test/src/ErrCode"
	ChatMsg "test/src/proto"
)

type ChatClient struct {
	UserName   string
	PassWorld  string
	conn       net.Conn
	serverAddr string
	*Common.ProtoDeal
	buff      []byte
	buffIndex uint16
	handle    map[string]func(interface{})
	target    string
}

func InitClient() {
	chatClient := ChatClient{
		serverAddr: "127.0.0.1:12588",
		handle:     make(map[string]func(interface{})),
	}
	chatClient.ProtoDeal = Common.GetProtoDealInstance()
	chatClient.connect()
	chatClient.RegistSelfAll()
	chatClient.BehaviorTree()
	chatClient.LoopDoNetData()
}
func (a *ChatClient) LoopDoNetData() {
	data := make([]byte, 1024)
	for {
		dataLen, err := a.conn.Read(data)
		if err != nil {
			fmt.Println("get err wrong", err)
			break
		}
		if dataLen == 0 {
			continue
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
			a.HandleMsg(msg)
			if a.buffIndex < 2 {
				break
			}
		}
	}
}

func (a *ChatClient) DoExit() {
	a.conn.Close()
}

func (a *ChatClient) connect() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", a.serverAddr)
	if err != nil {
		fmt.Println("addr err", err.Error())
		panic("addr error")
	}
	a.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("tcp error", err.Error())
		panic("tcp error")
	}

	fmt.Println("connection success")
}

func (a *ChatClient) SendSock(msg proto.Message) {
	data, err := a.Marshal(msg)
	if err != nil {
		fmt.Println("marshal msg error,msg:", msg)
	}
	_, err = a.conn.Write(data)
	if err != nil {
		fmt.Println("send sock error,msg:", msg)
	}
}

func (a *ChatClient) BehaviorTree() {
	a.DoLogin()
}

func (a *ChatClient) DoLogin() {
	fmt.Println("input UserName:")
	a.UserName = Common.KeyInput()
	fmt.Println("input PassWord:")
	a.PassWorld = Common.KeyInput()

	a.SendSock(&ChatMsg.CsLogin{UserName: a.UserName, PassWord: a.PassWorld})
}

func (a *ChatClient) CoutOnDesk(str string) {
	fmt.Println(str)
}

func (a *ChatClient) HandleMsg(msg interface{}) {
	msgName := Common.GetStructName(msg)
	f := a.handle[msgName]
	if f == nil {
		fmt.Println("HandleMsg wrong msg", msg)
	}
	f(msg)
}

func (a *ChatClient) ChooseTarget() {
	a.CoutOnDesk("input chat target:")
	targetName := Common.KeyInput()
	a.SendSock(&ChatMsg.CsChatTarget{UserName: targetName})
	a.target = targetName
}

func (a *ChatClient) Regist(msg interface{}, f func(interface{})) {
	a.handle[Common.GetStructName(msg)] = f
}

func (a *ChatClient) RegistSelfAll() {
	a.Regist(ChatMsg.ScLogin{}, a.ScLogin)
	a.Regist(ChatMsg.ScChatTarget{}, a.ScChatTarget)
	a.Regist(ChatMsg.ScChat{}, a.ScChat)
	a.Regist(ChatMsg.ScChatFrom{}, a.ScChatFrom)
}

func (a *ChatClient) ScLogin(i interface{}) {
	msg := i.(*ChatMsg.ScLogin)
	switch msg.ErrCode {
	case ErrCode.OK:
		a.CoutOnDesk("login ok")
		a.ChooseTarget()
	case ErrCode.LoginPassWord:
		a.CoutOnDesk("login password error,login again")
		a.DoLogin()
	default:
		fmt.Println("ScLogin unknown errCode: ", msg.ErrCode)
		a.DoLogin()
	}
}

func (a *ChatClient) ScChatTarget(i interface{}) {
	msg := i.(*ChatMsg.ScChatTarget)
	switch msg.ErrCode {
	case ErrCode.UserNotExist:
		fmt.Println("chat target user not exist")
		a.ChooseTarget()
	case ErrCode.RoleOffline:
		fmt.Println("target is offline,send content will be shown when he online")
		a.DoChat()
	case ErrCode.OK:
		fmt.Println("target is online")
		a.DoChat()
	default:
		fmt.Println("ScChatTarget unknown errCode", msg.ErrCode)
	}
}

func (a *ChatClient) DoChat() {
	fmt.Println("input content will send to ", a.target)
	content := Common.KeyInput()
	a.SendSock(&ChatMsg.CsChat{Content: content})
}

func (a *ChatClient) ScChat(i interface{}) {
	msg := i.(*ChatMsg.ScChat)
	switch msg.ErrCode {
	case ErrCode.ChatTargetNotSet:
		fmt.Println("not set chat target")
		a.ChooseTarget()
	case ErrCode.RoleOffline:
		fmt.Println("chat ok,but user offline:", a.target)
		a.DoChat()
	case ErrCode.OK:
		fmt.Println("chat ok,user online:", a.target)
		go a.DoChat()
	}
}

func (a *ChatClient) ScChatFrom(i interface{}) {
	msg := i.(*ChatMsg.ScChatFrom)
	fmt.Println("From ", msg.FromName, ":", msg.Content)
}
