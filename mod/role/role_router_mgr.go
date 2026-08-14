package role

import (
	"errors"
	"fmt"

	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"google.golang.org/protobuf/proto"
)

const RouterManagerName gen.Atom = "role_router_mgr"

type routerManagerOptions struct {
	zoneID  uint32
	service *Service
}

type routerManager struct {
	act.Actor
	zoneID uint32
	worker gen.Alias
	roles  map[string]gen.PID
}

func createRouterManager() gen.ProcessBehavior { return &routerManager{} }

func RouterManagerSpec(zoneID uint32, service *Service) act.SupervisorChildSpec {
	return act.SupervisorChildSpec{
		Name: RouterManagerName, Factory: createRouterManager,
		Args:    []any{routerManagerOptions{zoneID: zoneID, service: service}},
		Options: gen.ProcessOptions{MailboxSize: 8192},
	}
}

func (r *routerManager) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("role router manager expects options")
	}
	options, ok := args[0].(routerManagerOptions)
	if !ok || options.service == nil || options.zoneID == 0 {
		return fmt.Errorf("zone role service is not initialized")
	}
	r.zoneID = options.zoneID
	r.roles = make(map[string]gen.PID)
	worker, err := r.SpawnMeta(newRoleWorker(options.service), gen.MetaOptions{})
	if err != nil {
		return err
	}
	r.worker = worker
	return nil
}

func (r *routerManager) HandleMessage(_ gen.PID, message any) error {
	switch m := message.(type) {
	case clusterproto.MessageGatewayPacket:
		if m.ZoneID != r.zoneID || m.Account == "" || m.Platform == "" || m.SessionToken == "" {
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

func (r *routerManager) handleResult(result roleJobResult) error {
	if result.packet.MsgID == clusterproto.MsgLoginRequest {
		if result.err == nil && result.role.ID != 0 {
			key := roleKey(result.packet)
			pid := r.roles[key]
			if pid == (gen.PID{}) {
				var err error
				pid, err = r.Spawn(createRoleManager, gen.ProcessOptions{MailboxSize: 1024, LinkParent: true}, result.role)
				if err != nil {
					return r.respondError(result.packet, clusterproto.CodeInternal, "activate role failed")
				}
				r.roles[key] = pid
			}
			return r.SendPID(pid, bindSession{packet: result.packet})
		}
		response := &gamepb.LoginResponse{Code: clusterproto.CodeOK, Message: "success", NeedCreate: errors.Is(result.err, ErrNotFound)}
		if result.err != nil && !response.NeedCreate {
			response.Code, response.Message = clusterproto.CodeInternal, "role dependency error"
		}
		return r.respond(result.packet, clusterproto.MsgLoginResponse, response)
	}
	response := &gamepb.CreateRoleResponse{Code: clusterproto.CodeOK, Message: "success"}
	switch {
	case result.err == nil:
		response.Role = summary(result.role)
	case errors.Is(result.err, ErrNameConflict):
		response.Code, response.Message = clusterproto.CodeRoleNameTaken, "role name taken"
	case errors.Is(result.err, ErrAlreadyExists):
		response.Code, response.Message = clusterproto.CodeRoleExists, "role already exists"
	case errors.Is(result.err, ErrInvalidArgument):
		response.Code, response.Message = clusterproto.CodeInvalidArgument, "invalid role"
	default:
		response.Code, response.Message = clusterproto.CodeInternal, "role dependency error"
	}
	return r.respond(result.packet, clusterproto.MsgCreateRoleResponse, response)
}

func (r *routerManager) respondError(packet clusterproto.MessageGatewayPacket, code int32, message string) error {
	if packet.MsgID == clusterproto.MsgLoginRequest {
		return r.respond(packet, clusterproto.MsgLoginResponse, &gamepb.LoginResponse{Code: code, Message: message})
	}
	return r.respond(packet, clusterproto.MsgCreateRoleResponse, &gamepb.CreateRoleResponse{Code: code, Message: message})
}

func (r *routerManager) sendOrLogError(packet clusterproto.MessageGatewayPacket, code int32, message string) {
	if err := r.respondError(packet, code, message); err != nil {
		r.Log().Warning("zone error response failed for session %d: %s", packet.SessionID, err)
	}
}

func (r *routerManager) respond(packet clusterproto.MessageGatewayPacket, msgID uint32, message proto.Message) error {
	body, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return r.SendImportant(packet.ReplyTo, clusterproto.MessageGatewayResponse{SessionID: packet.SessionID, SessionToken: packet.SessionToken, MsgID: msgID, Seq: packet.Seq, Body: body})
}

func summary(value Role) *gamepb.RoleSummary {
	return &gamepb.RoleSummary{RoleId: uint64(value.ID), Name: value.Name, Level: uint32(value.Level), Platform: value.Platform, ZoneId: uint32(value.ZoneID)}
}

func roleKey(packet clusterproto.MessageGatewayPacket) string {
	return fmt.Sprintf("%s\x00%s\x00%d", packet.Account, packet.Platform, packet.ZoneID)
}
