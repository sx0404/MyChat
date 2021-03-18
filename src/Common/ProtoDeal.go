package Common

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	ChatMsg "test/src/proto"
)

type ProtoDeal struct{
	PdFInfo				map[string]func()proto.Message
}

var ProtoDealInstance *ProtoDeal

func (d *ProtoDeal) RegisterAll() {
	d.Register(&ChatMsg.CsLogin{},ChatMsg.NewCsLogin)
	d.Register(&ChatMsg.ScLogin{},ChatMsg.NewScLogin)
	d.Register(&ChatMsg.CsChat{},ChatMsg.NewCsChat)
	d.Register(&ChatMsg.ScChat{},ChatMsg.NewScChat)
	d.Register(&ChatMsg.CsChatTarget{},ChatMsg.NewCsChatTarget)
	d.Register(&ChatMsg.ScChatTarget{},ChatMsg.NewScChatTarget)
	d.Register(&ChatMsg.ScChatFrom{},ChatMsg.NewScChatFrom)
	d.Register(&ChatMsg.ScOfflineChatFrom{},ChatMsg.NewScOfflineChatFrom)
}

func (d *ProtoDeal) Register(msg proto.Message,f func()proto.Message) {
	PdName := GetStructName(msg)
	d.PdFInfo[PdName] = f
}

func GetProtoDealInstance() *ProtoDeal {
	if ProtoDealInstance == nil {
		pdFInfo := make(map[string]func()proto.Message)
		ProtoDealInstance = &ProtoDeal{
			PdFInfo: pdFInfo,
		}
		ProtoDealInstance.RegisterAll()
	}
	return ProtoDealInstance
}

func (d *ProtoDeal) PdFactory(PdName string) proto.Message {
	f := d.PdFInfo[PdName]
	if f == nil {
		fmt.Println("wrong Pd Factory")
		return nil
	}
	return f()
}

func (d *ProtoDeal) Marshal(i proto.Message) [] byte {
	b, err := proto.Marshal(i)
	if err != nil {
		fmt.Println("proto error:",i)
	}
	PdName := GetStructName(i)
	buff := append(StrToBytes(PdName), b...)
	lenPdNameB := make([]byte, 2)
	binary.BigEndian.PutUint16(lenPdNameB, uint16(len(PdName)))
	buff = append(lenPdNameB,buff...)
	lenB := make([]byte,2)
	binary.BigEndian.PutUint16(lenB, uint16(len(buff) + 2))
	buff = append(lenB,buff...)
	return buff
}