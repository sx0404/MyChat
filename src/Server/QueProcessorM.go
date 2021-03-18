package Server

import "sync"

//一个带消息队列的进程结构
type QueProcessorM struct {
	ProcessorNameInfo	sync.Map
	ProcessorIDInfo		sync.Map
}

const (
	_ = iota + 1
	PROC_EXIST 		//进程已经存在
	PROC_NOTEXIST
	ADD_PROC_IGNORE	//添加进程到管理的参数不合法被跳过
	ADD_PROCOK		//添加进程到管理成功
)

var QueProcessorMInstance *QueProcessorM

func GetQueProcessorManagerMInstance() *QueProcessorM {
	if QueProcessorMInstance == nil {
		QueProcessorMInstance = &QueProcessorM{
			ProcessorNameInfo: sync.Map{},
			ProcessorIDInfo:   sync.Map{},
		}
	}
	return QueProcessorMInstance
}

func (m *QueProcessorM) AddWithID(processor *QueProcessor) {
	m.ProcessorIDInfo.Load((*processor).id)
}

//添加一个进程信息到管理中来
func (m *QueProcessorM) AddWithName(processor *QueProcessor) uint8 {
	if (*processor).Name == "" {
		return ADD_PROC_IGNORE
	}else{
		_,ok := m.ProcessorNameInfo.Load((*processor).Name)
		if !ok {
			m.ProcessorNameInfo.Store((*processor).Name,processor)
		}else{
			//进程已经注册返回给上层选择是否杀死
			return PROC_EXIST
		}
	}
	return ADD_PROCOK
}

func DeleteFromQueProcessorM(processor *QueProcessor) {
	instance := GetQueProcessorManagerMInstance()
	instance.Delete(processor)
}

func (m *QueProcessorM) Delete(processor *QueProcessor) {
	m.ProcessorIDInfo.Delete((*processor).id)
	m.ProcessorNameInfo.Delete((*processor).Name)
}

func (m *QueProcessorM) FindWithName(name string) *QueProcessor {
	v,ok := m.ProcessorNameInfo.Load(name)
	if !ok {
		return nil
	}
	return v.(*QueProcessor)
}

func (m *QueProcessorM) FindWithID(id uint64) *QueProcessor {
	v,ok := m.ProcessorIDInfo.Load(id)
	if !ok {
		return nil
	}
	return v.(*QueProcessor)
}

//异步关闭一个进程
func (m *QueProcessorM) AsyncExitWithName(name string){
	proc := m.FindWithName(name)
	if proc == nil {
		return
	}else{
		proc.SendWithName(name, QueProcessExit{})
	}
}

func (m *QueProcessorM) AsyncExitWithID(id uint64) {
	proc := m.FindWithID(id)
	if proc == nil {
		return
	}else{
		proc.SendWithID(id, QueProcessExit{})
	}
}

func (m *QueProcessorM) SyncExitWithName(name string) {
	proc := m.FindWithName(name)
	if proc == nil {
		return
	}else{
		proc.CallWithName(name, QueProcessExit{},3000)
	}
}

func (m *QueProcessorM) SyncExitWithID(id uint64) {
	proc := m.FindWithID(id)
	if proc == nil {
		return
	}else{
		proc.CallWithID(id, QueProcessExit{},3000)
	}
}