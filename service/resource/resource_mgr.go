package resource

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type resourceManager struct{ act.Actor }

func createResourceManager() gen.ProcessBehavior { return &resourceManager{} }

func (m *resourceManager) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("resource owner expects application argument")
	}
	a, ok := args[0].(*Application)
	if !ok {
		return fmt.Errorf("invalid resource application")
	}
	if a.options.MySQL && a.mysql == nil {
		return ErrMySQLUnavailable
	}
	if a.options.Redis && a.redis == nil {
		return ErrRedisUnavailable
	}
	return nil
}
