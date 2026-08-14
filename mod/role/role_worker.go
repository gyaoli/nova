package role

import (
	"context"
	"strings"
	"time"

	clusterproto "nova/protocol/cluster"
	gamepb "nova/protocol/game"

	"ergo.services/ergo/gen"
	"google.golang.org/protobuf/proto"
)

type roleJob struct {
	packet clusterproto.MessageGatewayPacket
}

type roleJobResult struct {
	packet clusterproto.MessageGatewayPacket
	role   Role
	err    error
}

type roleWorker struct {
	gen.MetaProcess
	service *Service
	jobs    chan roleJob
	done    chan struct{}
}

func newRoleWorker(service *Service) *roleWorker {
	return &roleWorker{service: service, jobs: make(chan roleJob, 256), done: make(chan struct{})}
}

func (w *roleWorker) Init(process gen.MetaProcess) error {
	w.MetaProcess = process
	return nil
}

func (w *roleWorker) Start() error {
	for {
		select {
		case job := <-w.jobs:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			result := roleJobResult{packet: job.packet}
			if job.packet.MsgID == clusterproto.MsgLoginRequest {
				result.role, result.err = w.service.Login(ctx, job.packet.Account, job.packet.Platform, int32(job.packet.ZoneID))
			} else {
				var request gamepb.CreateRoleRequest
				if err := proto.Unmarshal(job.packet.Body, &request); err != nil || strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.OperationId) == "" {
					result.err = ErrInvalidArgument
				} else {
					initial, err := proto.Marshal(&gamepb.RoleData{SchemaVersion: 1})
					if err != nil {
						result.err = err
					} else {
						result.role, result.err = w.service.Create(ctx, CreateCommand{Account: job.packet.Account, Platform: job.packet.Platform, ZoneID: int32(job.packet.ZoneID), Name: request.Name, RegIP: job.packet.RegIP, InitialData: initial})
					}
				}
			}
			cancel()
			_ = w.Send(w.Parent(), result)
		case <-w.done:
			return nil
		}
	}
}

func (w *roleWorker) HandleMessage(_ gen.PID, message any) error {
	job, ok := message.(roleJob)
	if !ok {
		return nil
	}
	select {
	case w.jobs <- job:
		return nil
	default:
		return gen.ErrProcessMailboxFull
	}
}

func (w *roleWorker) HandleCall(gen.PID, gen.Ref, any) (any, error) {
	return gen.ErrUnsupported, nil
}

func (w *roleWorker) Terminate(error) {
	close(w.done)
}

func (w *roleWorker) HandleInspect(gen.PID, ...string) map[string]string {
	return map[string]string{"state": "running"}
}
