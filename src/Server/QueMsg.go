package Server

//所有内部路由用到的结构体
type Test struct {

}

type InfoLoadDB struct {

}

type InfoChangeGateWayToRoleLogic struct {
	Change *GateWayAgent
}

type InfoRoleLogicSendSock struct {
	I interface{}
}

type InfoRoleChat struct {
	FromName	string
	Content		string
}