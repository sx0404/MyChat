package Server

//玩家在线后的一些基本信息

type OnlineRole struct {
	onlineID			map[uint64]*RoleProcessor
	onlineName			map[string]*RoleProcessor
}

var onlineRoleInstance *OnlineRole

func GetOnlineRoleInstance() *OnlineRole {
	if onlineRoleInstance == nil {
		onlineRoleInstance = &OnlineRole{
			onlineID: make(map[uint64]*RoleProcessor),
			onlineName: make(map[string]*RoleProcessor),
		}
	}
	return onlineRoleInstance
}

func (this *OnlineRole) Add(proc *RoleProcessor) {
	if proc == nil {
		return
	}
	this.onlineID[proc.userID] = proc
	this.onlineName[proc.userName] = proc
}

func (this *OnlineRole) Delete(proc *RoleProcessor) {
	this.onlineID[proc.userID] = nil
	this.onlineName[proc.userName] = nil
}

func (this *OnlineRole) FindWithID(userID uint64) *RoleProcessor {
	return this.onlineID[userID]
}

func (this *OnlineRole) FindWithUserName(userName string) *RoleProcessor {
	return this.onlineName[userName]
}

func RegisterOnline(proc *RoleProcessor) {
	Instance := GetOnlineRoleInstance()
	Instance.Add(proc)
}

func RegisterOffline(proc *RoleProcessor) {
	Instance := GetOnlineRoleInstance()
	Instance.Delete(proc)
}

func GetOnlineRoleByUserName(userName string) *RoleProcessor {
	instance := GetOnlineRoleInstance()
	return instance.FindWithUserName(userName)
}

func GetOnlineRoleByUserID(userID uint64) *RoleProcessor {
	instance := GetOnlineRoleInstance()
	return instance.FindWithID(userID)
}