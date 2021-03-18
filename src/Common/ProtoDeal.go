package Common

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	ChatMsg "test/src/proto"
)

type ProtoDeal struct{
	PdFInfo				map[string]func()interface{}
}

var ProtoDealInstance *ProtoDeal

func (this *ProtoDeal) Unmarshal(data []byte) (interface{}, string) {

	// id
	var pdNameLen uint16
	pdNameLen = binary.BigEndian.Uint16(data)		//前两位作为proto名称的长度
	var pdName string
	pdName = string(data[2:2+pdNameLen])

	msg := this.PdFactory(pdName)
	if msg == nil {
		fmt.Println("factory proto wrong,msg:",pdName)
		return msg , ""
	}
	proto.UnmarshalMerge(data[2+pdNameLen:], msg.(proto.Message))
	return msg, pdName
}

func (this *ProtoDeal) RegisterAll() {
	this.Register(&ChatMsg.CsLogin{},ChatMsg.NewCsLogin)
	this.Register(&ChatMsg.ScLogin{},ChatMsg.NewScLogin)
	this.Register(&ChatMsg.CsChat{},ChatMsg.NewCsChat)
	this.Register(&ChatMsg.ScChat{},ChatMsg.NewScChat)
	this.Register(&ChatMsg.CsChatTarget{},ChatMsg.NewCsChatTarget)
	this.Register(&ChatMsg.ScChatTarget{},ChatMsg.NewScChatTarget)
	this.Register(&ChatMsg.ScChatFrom{},ChatMsg.NewScChatFrom)
	this.Register(&ChatMsg.ScOfflineChatFrom{},ChatMsg.NewScOfflineChatFrom)
}

func (this *ProtoDeal) Register(msg proto.Message,f func()interface{}) {
	PdName := GetStructName(msg)
	this.PdFInfo[PdName] = f
}

func GetProtoDealInstance() *ProtoDeal {
	if ProtoDealInstance == nil {
		pdFInfo := make(map[string]func()interface{})
		ProtoDealInstance = &ProtoDeal{
			PdFInfo: pdFInfo,
		}

	}
	return ProtoDealInstance
}

func (this *ProtoDeal) PdFactory(PdName string) interface{} {
	f := this.PdFInfo[PdName]
	if f == nil {
		fmt.Println("wrong Pd Factory")
		return nil
	}
	return f()
}

func (this *ProtoDeal) Marshal(i interface{}) [] byte {
	b, err := proto.Marshal(i.(proto.Message))
	if err != nil {
		fmt.Println("proto error:",i)
	}
	PdName := GetStructName(i)
	r := append(StrToBytes(PdName), b...)
	lenPdNameB := make([]byte, 2)
	binary.BigEndian.PutUint16(lenPdNameB, uint16(len(PdName)))
	r = append(lenPdNameB,r...)
	lenB := make([]byte,2)
	binary.BigEndian.PutUint16(lenB, uint16(len(r) + 2))
	r = append(lenB,r...)
	return r
}