package ChatMsg

import "github.com/golang/protobuf/proto"

func NewCsLogin() proto.Message {
	return proto.Message(new(CsLogin))
}
func NewScLogin() proto.Message {
	return proto.Message(new(ScLogin))
}
func NewCsChatTarget() proto.Message {
	return proto.Message(new(CsChatTarget))
}
func NewScChatTarget() proto.Message {
	return proto.Message(new(ScChatTarget))
}
func NewCsChat() proto.Message {
	return proto.Message(new(CsChat))
}
func NewScChat() proto.Message {
	return proto.Message(new(ScChat))
}
func NewScChatFrom() proto.Message {
	return proto.Message(new(ScChatFrom))
}
func NewScOfflineChatFrom() proto.Message {
	return proto.Message(new(ScOfflineChatFrom))
}