package cluster

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/testing/check"
	"ergo.services/ergo/testing/stage"
	"testing"
)

type echo struct{ act.Actor }

func createEcho() gen.ProcessBehavior { return &echo{} }

func (e *echo) HandleCall(_ gen.PID, _ gen.Ref, request any) (any, error) { return request, nil }

func TestMessageGatewayPacketCrossNodeEDF(t *testing.T) {
	s := stage.New(t)
	n1 := s.StartNode("gate")
	n2 := s.StartNode("zone")

	if err := n1.Native().Network().RegisterTypes(RegisterTypes()); err != nil {
		t.Fatal(err)
	}
	if err := n2.Native().Network().RegisterTypes(RegisterTypes()); err != nil {
		t.Fatal(err)
	}
	echoPID := n2.Spawn(createEcho, gen.ProcessOptions{})
	packet := MessageGatewayPacket{
		ReplyTo:   gen.ProcessID{Name: "gateway_writer", Node: "gate@localhost"},
		SessionID: 9, SessionToken: "token", Account: "account-1", Platform: "dev",
		ZoneID: 1, MsgID: MsgLoginRequest, Seq: 7, Body: []byte{1, 2, 3},
	}

	response, err := n1.Call(echoPID, packet)
	check.NoError(t, err)
	got := response.(MessageGatewayPacket)
	if got.SessionID != packet.SessionID || got.ReplyTo != packet.ReplyTo || string(got.Body) != string(packet.Body) {
		t.Fatalf("unexpected EDF round trip: %+v", got)
	}
}
