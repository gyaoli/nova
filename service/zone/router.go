package zone

import (
	"errors"
	"fmt"

	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"
	"nova/service/role"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"google.golang.org/protobuf/proto"
)

type router struct {
	act.Actor
	app    *application
	worker gen.Alias
	roles  map[string]gen.PID
}

func createRouter() gen.ProcessBehavior { return &router{} }

func (r *router) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("zone router expects application argument")
	}
	app, ok := args[0].(*application)
	if !ok || app.roles == nil {
		return fmt.Errorf("zone role service is not initialized")
	}
	r.app = app
	r.roles = make(map[string]gen.PID)
	worker, err := r.SpawnMeta(newRoleWorker(r.app.roles), gen.MetaOptions{})
	if err != nil {
		return err
	}
	r.worker = worker
	return nil
}

func (r *router) HandleMessage(_ gen.PID, message any) error {
	switch m := message.(type) {
	case clusterproto.MessageGatewayPacket:
		if m.ZoneID != uint32(r.app.config.Zone.ID) || m.Account == "" || m.Platform == "" || m.SessionToken == "" {
			r.sendOrLogError(m, clusterproto.CodeInvalidArgument, "invalid zone request")
			return nil
		}
		if m.MsgID != clusterproto.MsgLoginRequest && m.MsgID != clusterproto.MsgCreateRoleRequest {
			r.sendOrLogError(m, clusterproto.CodeInvalidArgument, "unknown zone message")
			return nil
		}
		if err := r.SendAlias(r.worker, roleJob{packet: m}); err != nil {
			r.sendOrLogError(m, clusterproto.CodeBusy, "zone busy")
		}
	case roleJobResult:
		if err := r.handleResult(m); err != nil {
			r.Log().Warning("zone response failed for session %d: %s", m.packet.SessionID, err)
		}
	}
	return nil
}

func (r *router) handleResult(result roleJobResult) error {
	if result.packet.MsgID == clusterproto.MsgLoginRequest {
		if result.err == nil && result.role.ID != 0 {
			key := roleKey(result.packet)
			pid := r.roles[key]
			if pid == (gen.PID{}) {
				var err error
				pid, err = r.Spawn(createRoleActor, gen.ProcessOptions{MailboxSize: 1024, LinkParent: true}, result.role)
				if err != nil {
					return r.respondError(result.packet, clusterproto.CodeInternal, "activate role failed")
				}
				r.roles[key] = pid
			}
			return r.SendPID(pid, bindSession{packet: result.packet})
		}
		response := &gamepb.LoginResponse{Code: clusterproto.CodeOK, Message: "success", NeedCreate: errors.Is(result.err, role.ErrNotFound)}
		if result.err != nil && !response.NeedCreate {
			response.Code, response.Message = clusterproto.CodeInternal, "role dependency error"
		}
		return r.respond(result.packet, clusterproto.MsgLoginResponse, response)
	}
	response := &gamepb.CreateRoleResponse{Code: clusterproto.CodeOK, Message: "success"}
	switch {
	case result.err == nil:
		response.Role = summary(result.role)
	case errors.Is(result.err, role.ErrNameConflict):
		response.Code, response.Message = clusterproto.CodeRoleNameTaken, "role name taken"
	case errors.Is(result.err, role.ErrAlreadyExists):
		response.Code, response.Message = clusterproto.CodeRoleExists, "role already exists"
	case errors.Is(result.err, role.ErrInvalidArgument):
		response.Code, response.Message = clusterproto.CodeInvalidArgument, "invalid role"
	default:
		response.Code, response.Message = clusterproto.CodeInternal, "role dependency error"
	}
	return r.respond(result.packet, clusterproto.MsgCreateRoleResponse, response)
}

func (r *router) respondError(packet clusterproto.MessageGatewayPacket, code int32, message string) error {
	if packet.MsgID == clusterproto.MsgLoginRequest {
		return r.respond(packet, clusterproto.MsgLoginResponse, &gamepb.LoginResponse{Code: code, Message: message})
	}
	return r.respond(packet, clusterproto.MsgCreateRoleResponse, &gamepb.CreateRoleResponse{Code: code, Message: message})
}

func (r *router) sendOrLogError(packet clusterproto.MessageGatewayPacket, code int32, message string) {
	if err := r.respondError(packet, code, message); err != nil {
		r.Log().Warning("zone error response failed for session %d: %s", packet.SessionID, err)
	}
}

func (r *router) respond(packet clusterproto.MessageGatewayPacket, msgID uint32, message proto.Message) error {
	body, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return r.SendImportant(packet.ReplyTo, clusterproto.MessageGatewayResponse{SessionID: packet.SessionID, SessionToken: packet.SessionToken, MsgID: msgID, Seq: packet.Seq, Body: body})
}

func summary(value role.Role) *gamepb.RoleSummary {
	return &gamepb.RoleSummary{RoleId: uint64(value.ID), Name: value.Name, Level: uint32(value.Level), Platform: value.Platform, ZoneId: uint32(value.ZoneID)}
}

func roleKey(packet clusterproto.MessageGatewayPacket) string {
	return fmt.Sprintf("%s\x00%s\x00%d", packet.Account, packet.Platform, packet.ZoneID)
}
