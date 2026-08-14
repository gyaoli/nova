package zone

import (
	"fmt"

	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"
	"nova/service/role"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"google.golang.org/protobuf/proto"
)

type sessionBinding struct {
	replyTo      gen.ProcessID
	sessionID    uint64
	sessionToken string
}

type bindSession struct {
	packet clusterproto.MessageGatewayPacket
}

type roleActor struct {
	act.Actor
	role    role.Role
	session sessionBinding
}

func createRoleActor() gen.ProcessBehavior { return &roleActor{} }

func (a *roleActor) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("role actor expects role snapshot")
	}
	value, ok := args[0].(role.Role)
	if !ok {
		return fmt.Errorf("invalid role snapshot")
	}
	a.role = value
	return nil
}

func (a *roleActor) HandleMessage(_ gen.PID, message any) error {
	bind, ok := message.(bindSession)
	if !ok {
		return nil
	}
	p := bind.packet
	a.session = sessionBinding{replyTo: p.ReplyTo, sessionID: p.SessionID, sessionToken: p.SessionToken}
	body, err := proto.Marshal(&gamepb.LoginResponse{Code: clusterproto.CodeOK, Message: "success", Role: summary(a.role)})
	if err != nil {
		return err
	}
	if err := a.SendImportant(a.session.replyTo, clusterproto.MessageGatewayResponse{SessionID: a.session.sessionID, SessionToken: a.session.sessionToken, MsgID: clusterproto.MsgLoginResponse, Seq: p.Seq, Body: body}); err != nil {
		a.Log().Warning("role login response failed for session %d: %s", a.session.sessionID, err)
	}
	return nil
}
