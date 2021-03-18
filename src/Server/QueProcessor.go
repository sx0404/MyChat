package Server

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"test/src/Common"
	"test/src/Route"
	"time"
)

type QueMsgM struct {
	QueMsgInfo		map[string]func(interface{})
	QueMsgCall		map[string]func(interface{}) CallReturnMsg
	procPtr			*QueProcessor
}

//一个带消息队列的进程结构
type QueProcessor struct {
	id			uint64
	Name		string
	Que			chan interface{}			//
	down		chan bool
	QueMsgM
}

type CallTimeOver struct {

}

type CallReturnMsg struct {
	s interface{}
}

type CallExitReturn struct {

}

type CallSendMsg struct {
	fromID				uint64
	s interface{}
}

//退出当前进程
type QueProcessExit struct {
	reson			string		//退出原因
}

func InitQueMsgM(process *QueProcessor) *QueMsgM {
	info := make(map[string]func(interface{}))
	call := make(map[string]func(interface{}) CallReturnMsg)
	queMsgM := &QueMsgM{info,call, process}
	return queMsgM
}

func (this *QueMsgM) RegisterAll(processor *QueProcessor) {
	//默认需要注册的
	this.RegisterInfo(CallSendMsg{}, this.procPtr.CallFunc)
	this.RegisterInfo(QueProcessExit{}, this.procPtr.Exit)
	this.RegisterCall(QueProcessExit{}, this.procPtr.SynxExit)

	//测试代码
	this.RegisterInfo(Test{}, Route.Test)

}

func (this *QueMsgM) RegisterInfo(a interface{},f func(interface{})) {
	this.QueMsgInfo[Common.GetStructName(a)] = f
}

func (this *QueMsgM) RegisterCall(i interface{},f func(interface{}) CallReturnMsg) {
	this.QueMsgCall[Common.GetStructName(i)] = f
}

func (this *QueMsgM) GetInfoFunc(strucName string) func(interface{}) {
	return this.QueMsgInfo[strucName]
}

func (this *QueMsgM) GetCallFunc(strucName string) func(interface{}) CallReturnMsg {
	return this.QueMsgCall[strucName]
}

func InitQueProcessor(name string) QueProcessor {
	queProcessor := QueProcessor{}
	queProcessor.Que = make(chan interface{},1000)
	queProcessor.down = make(chan bool)
	queMsgM := InitQueMsgM(&queProcessor)
	queProcessor.QueMsgM = *queMsgM
	queProcessor.Name = name
	return queProcessor
}

func (this *QueProcessor) RunQueProcessor() {
	this.RunQueProcessorInit()
	defer DeleteFromQueProcessorM(this)			//退出必定执行在管理列表中注销
	this.Loop()
}

func (this *QueProcessor) RunQueProcessorWithFunc(defaultFunc func()) {
	this.RunQueProcessorInit()
	defer DeleteFromQueProcessorM(this)			//退出必定执行在管理列表中注销
	this.LoopWithFunc(defaultFunc)
}

func (this *QueProcessor) RunQueProcessorInit() {
	processorID := GetGoroutineID()
	this.id = processorID
	processorManager := GetQueProcessorManagerMInstance()
	processorManager.AddWithID(this)
	processorManager.AddWithName(this)
}

//把当前进程加入到管理中
func (this *QueProcessor) AddProcToM() {
	GetQueProcessorManagerMInstance().AddWithName(this)
}

//异步发送消息给指定进程
func (this *QueProcessor) SendWithName(processorName string,i interface{}) {
	instance := GetQueProcessorManagerMInstance()
	targetProc := instance.FindWithName(processorName)
	this.Send(targetProc,i)
}

func (this *QueProcessor) SendSelf(i interface{}) {
	this.SendWithID(this.id,i)
}

func (this *QueProcessor) SendWithID(id uint64,i interface{}) {
	instance := GetQueProcessorManagerMInstance()
	targetProc := instance.FindWithID(id)
	this.Send(targetProc,i)
}

func (this *QueProcessor) Send(target *QueProcessor,i interface{}) {
	if target == nil {
		fmt.Println("SendWithName error",i)
	}else{
		target.Que <- i
	}
}

//执行channel中的函数进行路由
func (this *QueProcessor) LoopWithFunc(defaultFunc func()) {
	for {
		select {
			case a := <- this.Que:
				this.Route(a)
			case <- this.down:
				fmt.Println("goroutin down")
				return
			default:
				defaultFunc()
		}
	}
}

//执行channel中的函数进行路由
func (this *QueProcessor) Loop() {
	for {
		select {
		case a := <- this.Que:
			this.Route(a)
		case <- this.down:
			fmt.Println("goroutin down")
			return
		}
	}
}

func (this *QueProcessor) Route(i interface{}) {
	f := this.GetInfoFunc(Common.GetStructName(i))
	if f == nil {
		fmt.Println("Route not find,",i,this.Name)
		return
	}
	f(i)
}

func (this *QueProcessor) CallRoute(i interface{}) CallReturnMsg {
	f := this.QueMsgM.GetCallFunc(Common.GetStructName(i))
	return f(i)
}

//同步执行
func (this *QueProcessor) CallWithID(id uint64,s interface{},timeOut int16) {
	this.SendWithID(id, CallSendMsg{this.id,s})
	this.Call2()
}

func (this *QueProcessor) CallWithName(name string,s interface{},timeOut int16) {
	this.SendWithName(name, CallSendMsg{this.id,s})
	this.Call2()
}

func (this *QueProcessor) Call2() {
	for {
		//避免对面
		timer := time.NewTimer(time.Millisecond * 60)
		defer timer.Stop()

		timeOutFunc := func() {
			<-timer.C

			//投递一个退出消息给process
			this.Que <- CallTimeOver{}
		}
		// 等待定时器完成异步任务
		go timeOutFunc()

		recive := <- this.Que
		if reflect.TypeOf(recive).Name() == "CallTimeOver" {
			//接收到超时信息
			fmt.Println("call timeout")
			break
		}else if reflect.TypeOf(recive).Name() == "CallReturnMsg" {
			callReturnMsg := recive.(CallReturnMsg)
			this.Route(callReturnMsg.s)
		}
	}
}

func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

//异步处理的消息
func (this *QueProcessor) Exit(i interface{}) {
	exit := i.(QueProcessExit)
	fmt.Println("exit go,name :%s,id :%d,reson:",this.Name,this.id,exit.reson)
	this.DoExit()
}

//同步处理的消息
func (this *QueProcessor) SynxExit(i interface{}) CallReturnMsg {
	exit := i.(QueProcessExit)
	fmt.Println("exit go,name :%s,id :%d,reson:",this.Name,this.id,exit.reson)
	this.DoExit()
	return CallReturnMsg{}
}

func (this *QueProcessor) DoExit() {
	this.down <- true
}

func (this *QueProcessor) CallFunc(i interface{}) {
	callSendMsg := i.(CallSendMsg)
	callReturnmsg := this.CallRoute(callSendMsg)
	this.SendWithID(callSendMsg.fromID, callReturnmsg)
}