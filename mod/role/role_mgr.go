package role

import (
	"fmt"

	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"

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

type roleManager struct {
	act.Actor
	role    Role
	session sessionBinding
}

func createRoleManager() gen.ProcessBehavior { return &roleManager{} }

func (m *roleManager) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("role manager expects role snapshot")
	}
	value, ok := args[0].(Role)
	if !ok {
		return fmt.Errorf("invalid role snapshot")
	}
	m.role = value
	return nil
}

func (m *roleManager) HandleMessage(_ gen.PID, message any) error {
	bind, ok := message.(bindSession)
	if !ok {
		return nil
	}
	p := bind.packet
	m.session = sessionBinding{replyTo: p.ReplyTo, sessionID: p.SessionID, sessionToken: p.SessionToken}
	body, err := proto.Marshal(&gamepb.LoginResponse{Code: clusterproto.CodeOK, Message: "success", Role: summary(m.role)})
	if err != nil {
		return err
	}
	if err := m.SendImportant(m.session.replyTo, clusterproto.MessageGatewayResponse{SessionID: m.session.sessionID, SessionToken: m.session.sessionToken, MsgID: clusterproto.MsgLoginResponse, Seq: p.Seq, Body: body}); err != nil {
		m.Log().Warning("role login response failed for session %d: %s", m.session.sessionID, err)
	}
	return nil
}
