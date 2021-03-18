package Server

import (
	"fmt"
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

func (this *RoleProcessor) RegistAllHandle() {
	this.RegisterInfo(InfoLoadDB{},this.InfoLoadDB)
	this.RegisterInfo(InfoChangeGateWayToRoleLogic{},this.InfoChangeGateWayToRoleLogic)
	this.RegisterInfo(ChatMsg.CsChatTarget{},this.CsChatTarget)
	this.RegisterInfo(ChatMsg.CsChat{},this.CsChat)
	this.RegisterInfo(InfoRoleChat{},this.InfoRoleChat)
}

func (this *RoleProcessor) RunRoleProcessor() {
	//异步请求加载玩家数据
	go func() {
		this.InfoLoadDB(this)
		//注册已经在线的玩家信息
		RegisterOnline(this)
		//延迟函数保证关闭后取消在线信息
		defer RegisterOffline(this)
		defer this.gatewayAgent.DoExit()
		this.RunQueProcessor()
	}()
}

func (this *RoleProcessor) InfoChangeGateWayToRoleLogic(i interface{}) {
	infoChangeGateWayToRoleLogic := i.(InfoChangeGateWayToRoleLogic)
	this.gatewayAgent = infoChangeGateWayToRoleLogic.Change
}

func (this *RoleProcessor) InfoLoadDB(i interface{}) {
	roleInfo := db.GetUserByUserID(this.userID)
	roleMoney := db.GetRoleMoney(this.userID)
	roleFriendInfo := db.GetRoleFriendInfo(this.userID)

	Cache.SetRoleInfo(this.userID,roleInfo)
	Cache.SetRoleMoney(this.userID,roleMoney)
	Cache.SetRoleFriend(this.userID,roleFriendInfo)
}

func (this *RoleProcessor) GetRoleInfo() *Formation.RoleInfo {
	roleInfo := Cache.GetRoleInfo(this.userID)
	return roleInfo
}

func (this *RoleProcessor) SendSock(i interface{}) {
	if this.gatewayAgent == nil {
		fmt.Println("gateway not Found")
	}
	fmt.Println("send msg ",Common.GetStructName(i),i)
	C := this.Marshal(i)
	this.gatewayAgent.conn.Write(C)
}

func (this *RoleProcessor) CsChatTarget(i interface{}) {
	msg := i.(*ChatMsg.CsChatTarget)
	targetName := msg.GetUserName()
	//判断玩家是否在线
	target := GetOnlineRoleByUserName(targetName)
	if target == nil {
		//说明玩家不在线,通知给客户端
		targetID := db.GetUserIDByUserName(targetName)
		if targetID == 0 {
			//玩家不存在
			this.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.UserNotExist})
			return
		}
		this.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.RoleOffline})
		return
	}
	this.chatTarget = targetName
	this.SendSock(&ChatMsg.ScChatTarget{ErrCode: ErrCode.OK})
}

func (this *RoleProcessor) CsChat(i interface{}) {
	msg := i.(*ChatMsg.CsChat)
	if this.chatTarget == "" {
		this.SendSock(&ChatMsg.ScChat{ErrCode:ErrCode.ChatTargetNotSet})
		return
	}
	targetProc := GetOnlineRoleByUserName(this.chatTarget)
	if targetProc == nil {
		//通知客户端只能留言
		userID := db.GetUserIDByUserName(this.chatTarget)
		db.InsertChat(Formation.OfflineChat{
		SendID: userID,
		FromID: this.userID,
		FromName: this.userName,
		Content: msg.Content,
		})
		this.SendSock(&ChatMsg.ScChat{ErrCode: ErrCode.RoleOffline})
		return
	}
	this.Send(&targetProc.QueProcessor,InfoRoleChat{FromName: this.userName,Content: msg.Content})
	this.SendSock(&ChatMsg.ScChat{ErrCode: ErrCode.OK})
}

//接受到聊天消息发给自己
func (this *RoleProcessor) InfoRoleChat(i interface{}) {
	info := i.(InfoRoleChat)
	this.SendSock(&ChatMsg.ScChatFrom{FromName: info.FromName,Content: info.Content})
}

