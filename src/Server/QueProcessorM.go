package Server

//一个带消息队列的进程结构
type QueProcessorM struct {
	ProcessorNameInfo	map[string]*QueProcessor
	ProcessorIDInfo		map[uint64]*QueProcessor
}

const (
	_ = iota + 1
	PROCEXIST 		//进程已经存在
	PROCNOTEXIST
	ADDPROCIGNORE	//添加进程到管理的参数不合法被跳过
	ADDPROCOK		//添加进程到管理成功
)

var QueProcessorMInstance *QueProcessorM

func GetQueProcessorManagerMInstance() *QueProcessorM {
	if QueProcessorMInstance == nil {
		processNameInfo := make(map[string]*QueProcessor)
		processIDInfo := make(map[uint64]*QueProcessor)
		QueProcessorMInstance = &QueProcessorM{processNameInfo,processIDInfo}
	}
	return QueProcessorMInstance
}

func (this *QueProcessorM) AddWithID(processor *QueProcessor) {
	this.ProcessorIDInfo[(*processor).id] = processor
}

//添加一个进程信息到管理中来
func (this *QueProcessorM) AddWithName(processor *QueProcessor) uint8 {
	if (*processor).Name == "" {
		return ADDPROCIGNORE
	}else{
		if this.ProcessorNameInfo[(*processor).Name] == nil {
			//进程还未注册
			this.ProcessorNameInfo[(*processor).Name] = processor
		}else{
			//进程已经注册返回给上层选择是否杀死
			return PROCEXIST
		}
	}
	return ADDPROCOK
}

func DeleteFromQueProcessorM(processor *QueProcessor) {
	instance := GetQueProcessorManagerMInstance()
	instance.Delete(processor)
}

func (this *QueProcessorM) Delete(processor *QueProcessor) {
	this.ProcessorIDInfo[(*processor).id] = nil
	this.ProcessorNameInfo[(*processor).Name] = nil
}

func (this *QueProcessorM) FindWithName(name string) *QueProcessor {
	return this.ProcessorNameInfo[name]
}

func (this *QueProcessorM) FindWithID(id uint64) *QueProcessor {
	return this.ProcessorIDInfo[id]
}

//异步关闭一个进程
func (this *QueProcessorM) AsyncExitWithName(name string){
	if this.ProcessorNameInfo[name] == nil {
		return
	}else{
		stopProcess := this.ProcessorNameInfo[name]
		stopProcess.SendWithName(name, QueProcessExit{})
	}
}

func (this *QueProcessorM) AsyncExitWithID(id uint64) {
	if this.ProcessorIDInfo[id] == nil {
		return
	}else{
		stopProcess := this.ProcessorIDInfo[id]
		stopProcess.SendWithID(id, QueProcessExit{})
	}
}

func (this *QueProcessorM) SyncExitWithName(name string) {
	if this.ProcessorNameInfo[name] == nil {
		return
	}else{
		stopProcess := this.ProcessorNameInfo[name]
		stopProcess.CallWithName(name, QueProcessExit{},3000)
	}
}

func (this *QueProcessorM) SyncExitWithID(id uint64) {
	if this.ProcessorIDInfo[id] == nil {
		return
	}else{
		stopProcess := this.ProcessorIDInfo[id]
		stopProcess.CallWithID(id, QueProcessExit{},3000)
	}
}