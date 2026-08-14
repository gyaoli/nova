package game

import (
	"fmt"

	"nova/mod/role"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const supervisorName gen.Atom = "game_sup"

type supervisor struct{ act.Supervisor }

func createSupervisor() gen.ProcessBehavior { return &supervisor{} }

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	if len(args) != 1 {
		return act.SupervisorSpec{}, fmt.Errorf("game supervisor expects application argument")
	}
	a, ok := args[0].(*application)
	if !ok || a.roles == nil {
		return act.SupervisorSpec{}, fmt.Errorf("game application is not initialized")
	}
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{
			role.RouterManagerSpec(uint32(a.config.Zone.ID), a.roles),
		},
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    30,
		},
	}, nil
}
