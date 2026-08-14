package account

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	supervisorName gen.Atom = "account_sup"
	httpOwnerName  gen.Atom = "account_http"
)

type supervisor struct{ act.Supervisor }

func createSupervisor() gen.ProcessBehavior { return &supervisor{} }

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	if len(args) != 1 {
		return act.SupervisorSpec{}, fmt.Errorf("account supervisor expects application argument")
	}
	a, ok := args[0].(*application)
	if !ok || a.owner == nil {
		return act.SupervisorSpec{}, fmt.Errorf("account application is not initialized")
	}
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{{
			Name: httpOwnerName, Factory: createHTTPOwner, Args: []any{a},
		}},
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    30,
		},
	}, nil
}
