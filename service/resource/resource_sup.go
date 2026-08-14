package resource

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const (
	supervisorName gen.Atom = "resource_sup"
	managerName    gen.Atom = "resource_mgr"
)

// supervisor defines the resource failure boundary.
type supervisor struct{ act.Supervisor }

func createSupervisor() gen.ProcessBehavior { return &supervisor{} }

func (s *supervisor) Init(args ...any) (act.SupervisorSpec, error) {
	if len(args) != 1 {
		return act.SupervisorSpec{}, fmt.Errorf("resource supervisor expects application argument")
	}
	a, ok := args[0].(*Application)
	if !ok {
		return act.SupervisorSpec{}, fmt.Errorf("invalid resource application")
	}
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Children: []act.SupervisorChildSpec{{
			Name: managerName, Factory: createResourceManager, Args: []any{a},
		}},
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 5,
			Period:    30,
		},
	}, nil
}
