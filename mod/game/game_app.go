package game

import (
	"time"

	configenv "nova/core/env"
	"nova/mod/role"
	clusterproto "nova/protocol/cluster"
	"nova/service/resource"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const applicationName gen.Atom = "game"

type application struct {
	app.Application
	config    configenv.NodeConfig
	resources resource.Provider
	lease     *resource.Lease
	roles     *role.Service
}

func NewApplication(config configenv.NodeConfig, resources resource.Provider) gen.ApplicationBehavior {
	return &application{config: config, resources: resources}
}

func (a *application) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name: applicationName, Description: "Nova game logic modules",
		Version: gen.Version{Name: "game", Release: "1.0.0"}, Mode: gen.ApplicationModePermanent,
		Depends:     gen.ApplicationDepends{Applications: []gen.Atom{resource.ApplicationName}},
		Network:     gen.ApplicationNetwork{RegisterTypes: clusterproto.RegisterTypes()},
		Map:         map[string]gen.Atom{"role": role.RouterManagerName},
		Group:       []gen.ApplicationMemberSpec{{Factory: createSupervisor, Name: supervisorName, Args: []any{a}}},
		InitTimeout: 5 * time.Second, StopTimeout: 5 * time.Second,
	}, nil
}

func (a *application) Init(gen.Ref, gen.ApplicationMode) error {
	lease, err := a.resources.Acquire(resource.Options{MySQL: true})
	if err != nil {
		return err
	}
	client, err := lease.MySQL()
	if err != nil {
		lease.Release()
		return err
	}
	a.lease = lease
	a.roles = role.NewService(role.NewRepository(client.DB()))
	return nil
}

func (a *application) Terminate(error) {
	if a.lease != nil {
		a.lease.Release()
		a.lease = nil
	}
}
