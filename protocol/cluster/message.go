package cluster

import "ergo.services/ergo/gen"

const (
	MsgLoginRequest       uint32 = 1001
	MsgLoginResponse      uint32 = 1002
	MsgCreateRoleRequest  uint32 = 1003
	MsgCreateRoleResponse uint32 = 1004
	MsgHeartbeatRequest   uint32 = 1005
	MsgHeartbeatResponse  uint32 = 1006
)

const (
	CodeOK              int32 = 0
	CodeInvalidArgument int32 = 1001
	CodeUnauthorized    int32 = 1002
	CodeRoleNameTaken   int32 = 1003
	CodeRoleExists      int32 = 1004
	CodeBusy            int32 = 1005
	CodeInternal        int32 = 1006
)

type MessageGatewayPacket struct {
	ReplyTo      gen.ProcessID
	SessionID    uint64
	SessionToken string
	Account      string
	Platform     string
	ZoneID       uint32
	RegIP        string
	MsgID        uint32
	Seq          uint32
	Body         []byte
}

type MessageGatewayResponse struct {
	SessionID    uint64
	SessionToken string
	MsgID        uint32
	Seq          uint32
	Body         []byte
}

func RegisterTypes() []any {
	return []any{MessageGatewayPacket{}, MessageGatewayResponse{}}
}
