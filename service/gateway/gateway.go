package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/meta"
	"google.golang.org/protobuf/proto"
	configenv "nova/core/env"
	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"
	"nova/protocol/wire"
)

type session struct {
	id       uint64
	token    string
	remoteIP string
	account  string
	platform string
	zoneID   uint32
}

type gateway struct {
	act.Actor
	config      configenv.NodeConfig
	server      gen.Alias
	auth        gen.Alias
	nextSession atomic.Uint64
	connections map[gen.Alias]*session
	byID        map[uint64]gen.Alias
}

func createGateway() gen.ProcessBehavior { return &gateway{} }

func (g *gateway) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("gateway expects configuration")
	}
	g.config = args[0].(configenv.NodeConfig)
	g.connections = make(map[gen.Alias]*session)
	g.byID = make(map[uint64]gen.Alias)
	authID, err := g.SpawnMeta(newAuthWorker(g.config.Gateway.AuthURL), gen.MetaOptions{})
	if err != nil {
		return err
	}
	g.auth = authID
	server, err := meta.CreateTCPServer(meta.TCPServerOptions{
		Host: g.config.Gateway.ListenHost, Port: uint16(g.config.Gateway.ListenPort),
		ReadChunk: meta.ChunkOptions{
			Enable: true, HeaderSize: wire.HeaderSize, HeaderLengthPosition: 0,
			HeaderLengthSize: 4, HeaderLengthIncludesHeader: true, MaxLength: wire.MaxFrameSize,
		},
		Advanced: meta.TCPAdvancedOptions{KeepAlivePeriod: 30 * time.Second},
	})
	if err != nil {
		return err
	}
	g.server, err = g.SpawnMeta(server, gen.MetaOptions{})
	return err
}

func (g *gateway) HandleMessage(_ gen.PID, message any) error {
	switch m := message.(type) {
	case meta.MessageTCPConnect:
		id := g.nextSession.Add(1)
		g.connections[m.ID] = &session{id: id, token: randomToken(), remoteIP: remoteIP(m.RemoteAddr)}
		g.byID[id] = m.ID
	case meta.MessageTCPDisconnect:
		if s := g.connections[m.ID]; s != nil {
			delete(g.byID, s.id)
		}
		delete(g.connections, m.ID)
	case meta.MessageTCP:
		if err := g.handleFrame(m.ID, m.Data); err != nil {
			g.Log().Warning("gateway frame failed for session %s: %s", m.ID, err)
			_ = g.SendExitMeta(m.ID, err)
		}
	case authResult:
		if err := g.handleAuthResult(m); err != nil {
			g.Log().Warning("gateway authentication result failed: %s", err)
		}
	case clusterproto.MessageGatewayResponse:
		if err := g.handleZoneResponse(m); err != nil {
			g.Log().Warning("gateway response write failed for session %d: %s", m.SessionID, err)
		}
	}
	return nil
}

func (g *gateway) handleFrame(connection gen.Alias, data []byte) error {
	s := g.connections[connection]
	if s == nil {
		return nil
	}
	frame, err := wire.Decode(data)
	if err != nil {
		return g.SendExitMeta(connection, err)
	}
	switch frame.Header.MsgID {
	case clusterproto.MsgLoginRequest:
		var request gamepb.LoginRequest
		if err := proto.Unmarshal(frame.Body, &request); err != nil || strings.TrimSpace(request.Account) == "" || strings.TrimSpace(request.Token) == "" || request.ZoneId == 0 {
			return g.sendError(connection, clusterproto.MsgLoginResponse, frame.Header.Seq, clusterproto.CodeInvalidArgument, "invalid login request")
		}
		job := authJob{connection: connection, account: request.Account, token: request.Token, frameBody: frame.Body, seq: frame.Header.Seq}
		if err := g.SendAlias(g.auth, job); err != nil {
			return g.sendError(connection, clusterproto.MsgLoginResponse, frame.Header.Seq, clusterproto.CodeBusy, "gateway busy")
		}
	case clusterproto.MsgHeartbeatRequest:
		var request gamepb.HeartbeatRequest
		if err := proto.Unmarshal(frame.Body, &request); err != nil {
			return g.sendError(connection, clusterproto.MsgHeartbeatResponse, frame.Header.Seq, clusterproto.CodeInvalidArgument, "invalid heartbeat")
		}
		body, _ := proto.Marshal(&gamepb.HeartbeatResponse{ClientTime: request.ClientTime, ServerTime: time.Now().UnixMilli()})
		return g.write(connection, clusterproto.MsgHeartbeatResponse, frame.Header.Seq, body)
	case clusterproto.MsgCreateRoleRequest:
		if s.account == "" {
			return g.sendError(connection, clusterproto.MsgCreateRoleResponse, frame.Header.Seq, clusterproto.CodeUnauthorized, "login required")
		}
		return g.forward(s, clusterproto.MsgCreateRoleRequest, frame.Header.Seq, frame.Body)
	default:
		return fmt.Errorf("unknown message id %d", frame.Header.MsgID)
	}
	return nil
}

func (g *gateway) handleAuthResult(result authResult) error {
	s := g.connections[result.job.connection]
	if s == nil {
		return nil
	}
	if !result.valid {
		return g.sendError(result.job.connection, clusterproto.MsgLoginResponse, result.job.seq, clusterproto.CodeUnauthorized, "token invalid")
	}
	var request gamepb.LoginRequest
	if err := proto.Unmarshal(result.job.frameBody, &request); err != nil {
		return err
	}
	// Every accepted login creates a new fencing token, even when the client
	// reuses the same TCP connection. Delayed responses from the old role
	// binding are therefore rejected by handleZoneResponse.
	s.token = randomToken()
	s.account, s.platform, s.zoneID = request.Account, request.Platform, request.ZoneId
	// The token is consumed at the Gateway trust boundary and is not forwarded
	// to the Zone as part of the opaque client payload.
	return g.forward(s, clusterproto.MsgLoginRequest, result.job.seq, nil)
}

func (g *gateway) forward(s *session, msgID, seq uint32, body []byte) error {
	return g.SendImportant(gen.ProcessID{Name: clusterproto.ZoneRouterName, Node: gen.Atom(g.config.Gateway.ZoneNode)}, clusterproto.MessageGatewayPacket{
		ReplyTo: gen.ProcessID{Name: ownerName, Node: g.Node().Name()}, SessionID: s.id,
		SessionToken: s.token, Account: s.account, Platform: s.platform, ZoneID: s.zoneID,
		RegIP: s.remoteIP, MsgID: msgID, Seq: seq, Body: append([]byte(nil), body...),
	})
}

func (g *gateway) handleZoneResponse(response clusterproto.MessageGatewayResponse) error {
	connection, ok := g.byID[response.SessionID]
	if !ok {
		return nil
	}
	s := g.connections[connection]
	if s == nil || s.token != response.SessionToken {
		return nil
	}
	return g.write(connection, response.MsgID, response.Seq, response.Body)
}

func (g *gateway) sendError(connection gen.Alias, msgID, seq uint32, code int32, message string) error {
	var value proto.Message
	if msgID == clusterproto.MsgLoginResponse {
		value = &gamepb.LoginResponse{Code: code, Message: message}
	} else {
		value = &gamepb.CreateRoleResponse{Code: code, Message: message}
	}
	body, _ := proto.Marshal(value)
	return g.write(connection, msgID, seq, body)
}

func (g *gateway) write(connection gen.Alias, msgID, seq uint32, body []byte) error {
	data, err := wire.Encode(msgID, seq, body)
	if err != nil {
		return err
	}
	return g.SendAlias(connection, meta.MessageTCP{Data: data})
}

func randomToken() string {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(data)
}

func remoteIP(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil || len(host) > 16 {
		return ""
	}
	return host
}
