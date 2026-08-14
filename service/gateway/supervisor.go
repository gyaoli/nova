package gateway

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	supervisorName gen.Atom = "gateway_sup"
	ownerName      gen.Atom = "gateway_writer"
)

type supervisor struct{ act.Supervisor }

func createSupervisor() gen.ProcessBehavior { return &supervisor{} }

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	if len(args) != 1 {
		return act.SupervisorSpec{}, fmt.Errorf("gateway supervisor expects application argument")
	}
	a, ok := args[0].(*application)
	if !ok {
		return act.SupervisorSpec{}, fmt.Errorf("invalid gateway application")
	}
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{{
			Name: ownerName, Factory: createGateway, Args: []any{a.config},
			Options: gen.ProcessOptions{MailboxSize: 8192},
		}},
		Restart: act.SupervisorRestart{Strategy: act.SupervisorStrategyPermanent, Intensity: 5, Period: 30},
	}, nil
}
