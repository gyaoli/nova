package resource

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type owner struct{ act.Actor }

func createOwner() gen.ProcessBehavior { return &owner{} }

func (o *owner) Init(args ...any) error {
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
