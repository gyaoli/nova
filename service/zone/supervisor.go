package zone

import (
	"fmt"

	clusterproto "nova/protocol/cluster"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const supervisorName gen.Atom = "zone_sup"

type supervisor struct{ act.Supervisor }

func createSupervisor() gen.ProcessBehavior { return &supervisor{} }

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	if len(args) != 1 {
		return act.SupervisorSpec{}, fmt.Errorf("zone supervisor expects application argument")
	}
	a, ok := args[0].(*application)
	if !ok || a.roles == nil {
		return act.SupervisorSpec{}, fmt.Errorf("zone application is not initialized")
	}
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{{
			Name: clusterproto.ZoneRouterName, Factory: createRouter, Args: []any{a},
			Options: gen.ProcessOptions{MailboxSize: 8192},
		}},
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    30,
		},
	}, nil
}
