package Server

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"test/src/Common"
	"test/src/Route"
	"time"
)

type QueMsgM struct {
	QueMsgInfo map[string]func(interface{})
	QueMsgCall map[string]func(interface{}) CallReturnMsg
	procPtr    *QueProcessor
}

//一个带消息队列的进程结构
type QueProcessor struct {
	id       uint64
	Name     string
	Que      chan interface{} //
	down     chan bool
	timeTick *time.Ticker
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
	fromID uint64
	s      interface{}
}

//退出当前进程
type QueProcessExit struct {
	reson string //退出原因
}

func InitQueMsgM(process *QueProcessor) *QueMsgM {
	info := make(map[string]func(interface{}))
	call := make(map[string]func(interface{}) CallReturnMsg)
	queMsgM := &QueMsgM{info, call, process}
	return queMsgM
}

func (this *QueMsgM) RegisterAll(processor *QueProcessor) {
	//默认需要注册的
	this.RegisterInfo(CallSendMsg{}, this.procPtr.CallFunc)
	this.RegisterInfo(QueProcessExit{}, this.procPtr.Exit)
	this.RegisterCall(QueProcessExit{}, this.procPtr.SyncExit)

	//测试代码
	this.RegisterInfo(Test{}, Route.Test)

}

func (this *QueMsgM) RegisterInfo(a interface{}, f func(interface{})) {
	this.QueMsgInfo[Common.GetStructName(a)] = f
}

func (this *QueMsgM) RegisterCall(i interface{}, f func(interface{}) CallReturnMsg) {
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
	queProcessor.Que = make(chan interface{}, 1000)
	queProcessor.down = make(chan bool)
	queMsgM := InitQueMsgM(&queProcessor)
	queProcessor.QueMsgM = *queMsgM
	queProcessor.Name = name
	return queProcessor
}

func (p *QueProcessor) RunQueProcessor() {
	p.RunQueProcessorInit()
	defer DeleteFromQueProcessorM(p) //退出必定执行在管理列表中注销
	p.Loop()
}

func (p *QueProcessor) RunQueProcessorWithFunc(defaultFunc func()) {
	p.RunQueProcessorInit()
	defer DeleteFromQueProcessorM(p) //退出必定执行在管理列表中注销
	p.LoopWithFunc(defaultFunc)
}

func (p *QueProcessor) RunQueProcessorInit() {
	processorID := GetGoroutineID()
	p.id = processorID
	processorManager := GetQueProcessorManagerMInstance()
	processorManager.AddWithID(p)
	processorManager.AddWithName(p)
}

//把当前进程加入到管理中
func (p *QueProcessor) AddProcToM() {
	GetQueProcessorManagerMInstance().AddWithName(p)
}

//异步发送消息给指定进程
func (p *QueProcessor) SendWithName(processorName string, i interface{}) {
	instance := GetQueProcessorManagerMInstance()
	targetProc := instance.FindWithName(processorName)
	p.Send(targetProc, i)
}

func (p *QueProcessor) SendSelf(i interface{}) {
	p.SendWithID(p.id, i)
}

func (p *QueProcessor) SendWithID(id uint64, i interface{}) {
	instance := GetQueProcessorManagerMInstance()
	targetProc := instance.FindWithID(id)
	p.Send(targetProc, i)
}

func (p *QueProcessor) Send(target *QueProcessor, i interface{}) {
	if target == nil {
		fmt.Println("SendWithName error", i)
	} else {
		target.Que <- i
	}
}

//执行channel中的函数进行路由
func (p *QueProcessor) LoopWithFunc(defaultFunc func()) {
	for {
		select {
		case a := <-p.Que:
			p.Route(a)
		case <-p.down:
			fmt.Println("goroutine down")
			return
		default:
			defaultFunc()
		}
	}
}

//执行channel中的函数进行路由
func (p *QueProcessor) Loop() {
	p.timeTick = time.NewTicker(time.Second)
	for {
		select {
		case a := <-p.Que:
			p.Route(a)
		case <-p.down:
			fmt.Println("goroutine down")
			return
		case <-p.timeTick.C:
			p.timeTick = time.NewTicker(time.Second)
			//TODO 后面可以🏠事件注册来实现定时任务
			//fmt.Println("11111")
		}
	}
}

func (p *QueProcessor) Route(i interface{}) {
	f := p.GetInfoFunc(Common.GetStructName(i))
	if f == nil {
		fmt.Println("Route not find,", i, p.Name)
		return
	}
	f(i)
}

func (p *QueProcessor) CallRoute(i interface{}) CallReturnMsg {
	f := p.QueMsgM.GetCallFunc(Common.GetStructName(i))
	return f(i)
}

//同步执行
func (p *QueProcessor) CallWithID(id uint64, s interface{}, timeOut int64) error {
	p.SendWithID(id, CallSendMsg{p.id, s})
	return p.Call(timeOut)
}

func (p *QueProcessor) CallWithName(name string, s interface{}, timeOut int64) error {
	p.SendWithName(name, CallSendMsg{p.id, s})
	return p.Call(timeOut)
}

func (p *QueProcessor) Call(timeOut int64) error {
	for {
		var timer = time.NewTimer(time.Millisecond * time.Duration(timeOut))
		sliceI := make([]interface{}, 100)
		select {
		case receive := <-p.Que:
			if reflect.TypeOf(receive).Name() == "CallReturnMsg" {
				callReturnMsg := receive.(CallReturnMsg)
				p.Route(callReturnMsg.s)
				for i := range sliceI {
					p.SendSelf(i) //把之前遗留的消息返回
				}
				return nil
			} else {
				sliceI = append(sliceI, receive)
			}
		case <-timer.C:
			return errors.New("time out")
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
func (p *QueProcessor) Exit(i interface{}) {
	exit := i.(QueProcessExit)
	fmt.Println("goroutine exist ", p.Name, p.id, exit.reson)
	p.DoExit()
}

//同步处理的消息
func (p *QueProcessor) SyncExit(i interface{}) CallReturnMsg {
	exit := i.(QueProcessExit)
	fmt.Println("SyncExit:", p.Name, p.id, exit.reson)
	p.DoExit()
	return CallReturnMsg{}
}

func (p *QueProcessor) DoExit() {
	p.down <- true
}

func (p *QueProcessor) CallFunc(i interface{}) {
	callSendMsg := i.(CallSendMsg)
	callReturns := p.CallRoute(callSendMsg)
	p.SendWithID(callSendMsg.fromID, callReturns)
}
