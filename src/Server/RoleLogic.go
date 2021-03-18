package Server

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"test/src/Cache"
	"test/src/Common"
	db "test/src/DB"
	"test/src/ErrCode"
	"test/src/Formation"
	ChatMsg "test/src/proto"
)

//玩家的逻辑协程
type RoleProcessor struct {
	gatewayAgent *GateWayAgent	//与逻辑进程关联的网管协程
	QueProcessor
	userID		uint64
	userName	string
	chatTarget	string
	*Common.ProtoDeal
}

func GenRoleProcessorName(userID uint64) string {
	return "Role" + Common.ToString(int64(userID))
}

func InitRoleProcessor(gatewayAgent *GateWayAgent,userName string) *RoleProcessor{
	queProcessor := InitQueProcessor(GenRoleProcessorName(gatewayAgent.GetUserID()))
	roleProcessor := &RoleProcessor{
		gatewayAgent: gatewayAgent,
		QueProcessor: queProcessor,
		userID: gatewayAgent.GetUserID(),
		userName :userName,
		ProtoDeal : Common.GetProtoDealInstance(),
	}
	roleProcessor.RegistAllHandle()
	return roleProcessor
}

func (p *RoleProcessor) RegistAllHandle() {
	p.RegisterInfo(InfoLoadDB{}, p.InfoLoadDB)
	p.RegisterInfo(InfoChangeGateWayToRoleLogic{}, p.InfoChangeGateWayToRoleLogic)
	p.RegisterInfo(ChatMsg.CsChatTarget{}, p.CsChatTarget)
	p.RegisterInfo(ChatMsg.CsChat{}, p.CsChat)
	p.RegisterInfo(InfoRoleChat{}, p.InfoRoleChat)
}

func (p *RoleProcessor) RunRoleProcessor() {
	//异步请求加载玩家数据
	go func() {
		p.InfoLoadDB(p)
		//注册已经在线的玩家信息
		RegisterOnline(p)
		//延迟函数保证关闭后取消在线信息
		defer RegisterOffline(p)
		defer p.gatewayAgent.DoExit()
		p.RunQueProcessor()
	}()
}

func (p *RoleProcessor) InfoChangeGateWayToRoleLogic(i interface{}) {
	infoChangeGateWayToRoleLogic := i.(InfoChangeGateWayToRoleLogic)
	p.gatewayAgent = infoChangeGateWayToRoleLogic.Change
}

func (p *RoleProcessor) InfoLoadDB(i interface{}) {
	roleInfo := db.GetUserByUserID(p.userID)
	roleMoney := db.GetRoleMoney(p.userID)
	roleFriendInfo := db.GetRoleFriendInfo(p.userID)

	Cache.SetRoleInfo(p.userID,roleInfo)
	Cache.SetRoleMoney(p.userID,roleMoney)
	Cache.SetRoleFriend(p.userID,roleFriendInfo)
}

func (p *RoleProcessor) GetRoleInfo() *Formation.RoleInfo {
	roleInfo := Cache.GetRoleInfo(p.userID)
	return roleInfo
}

func (p *RoleProcessor) SendSock(i proto.Message) {
	if p.gatewayAgent == nil {
		fmt.Println("gateway not Found")
	}
	fmt.Println("send msg ",Common.GetStructName(i),i)
	C := p.Marshal(i)
	p.gatewayAgent.conn.Write(C)
}

func (p *RoleProcessor) CsChatTarget(i interface{}) {
	msg := i.(*ChatMsg.CsChatTarget)
	targetName := msg.GetUserName()
	//判断玩家是否在线
	target := GetOnlineRoleByUserName(targetName)
	if target == nil {
		//说明玩家不在线,通知给客户端
		targetID := db.GetUserIDByUserName(targetName)
		if targetID == 0 {
			//玩家不存在
			p.chatTarget = targetName
			p.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.UserNotExist})
			return
		}
		p.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.RoleOffline})
		return
	}
	p.chatTarget = targetName
	p.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.OK})
}

func (p *RoleProcessor) CsChat(i interface{}) {
	msg := i.(*ChatMsg.CsChat)
	if p.chatTarget == "" {
		p.SendSock(&ChatMsg.ScChat{ErrCode: ErrCode.ChatTargetNotSet})
		return
	}
	targetProc := GetOnlineRoleByUserName(p.chatTarget)
	if targetProc == nil {
		//通知客户端只能留言
		userID := db.GetUserIDByUserName(p.chatTarget)
		db.InsertChat(Formation.OfflineChat{
		SendID:   userID,
		FromID:   p.userID,
		FromName: p.userName,
		Content:  msg.Content,
		})
		p.SendSock(&ChatMsg.ScChat{ErrCode: ErrCode.RoleOffline})
		return
	}
	p.Send(&targetProc.QueProcessor,InfoRoleChat{FromName: p.userName,Content: msg.Content})
	p.SendSock(&ChatMsg.ScChat{ErrCode: ErrCode.OK})
}

//接受到聊天消息发给自己
func (p *RoleProcessor) InfoRoleChat(i interface{}) {
	info := i.(InfoRoleChat)
	p.SendSock(&ChatMsg.ScChatFrom{FromName: info.FromName,Content: info.Content})
}

