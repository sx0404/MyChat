package Server

import "sync"

//玩家在线后的一些基本信息

type OnlineRole struct {
	onlineID			sync.Map
	onlineName			sync.Map
}

var onlineRoleInstance *OnlineRole

func GetOnlineRoleInstance() *OnlineRole {
	if onlineRoleInstance == nil {
		onlineRoleInstance = &OnlineRole{
			onlineID: sync.Map{},
			onlineName: sync.Map{},
		}
	}
	return onlineRoleInstance
}

func (r *OnlineRole) Add(proc *RoleProcessor) {
	if proc == nil {
		return
	}
	r.onlineID.Store(proc.id,proc)
	r.onlineName.Store(proc.userName,proc)
}

func (r *OnlineRole) Delete(proc *RoleProcessor) {
	r.onlineID.Delete(proc.userID)
	r.onlineName.Delete(proc.userName)
}

func (r *OnlineRole) FindWithID(userID uint64) *RoleProcessor {
	c,ok := r.onlineID.Load(userID)
	if !ok {
		return nil
	}
	return c.(*RoleProcessor)
}

func (r *OnlineRole) FindWithUserName(userName string) *RoleProcessor {
	c,ok := r.onlineName.Load(userName)
	if !ok {
		return nil
	}
	return c.(*RoleProcessor)
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