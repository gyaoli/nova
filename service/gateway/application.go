package gateway

import (
	"errors"
	"time"

	configenv "nova/core/env"
	clusterproto "nova/protocol/cluster"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	applicationName gen.Atom = "gateway"
)

type application struct {
	app.Application
	config configenv.NodeConfig
}

func NewApplication(config configenv.NodeConfig) gen.ApplicationBehavior {
	return &application{config: config}
}

func (a *application) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name: applicationName, Description: "Nova client gateway",
		Version: gen.Version{Name: "gateway", Release: "1.0.0"},
		Mode:    gen.ApplicationModePermanent,
		Network: gen.ApplicationNetwork{RegisterTypes: clusterproto.RegisterTypes()},
		Map:     map[string]gen.Atom{"writer": ownerName},
		Group: []gen.ApplicationMemberSpec{{
			Factory: createSupervisor, Name: "gateway_sup", Args: []any{a},
		}},
		InitTimeout: 5 * time.Second, StopTimeout: 5 * time.Second,
	}, nil
}

func (a *application) Init(gen.Ref, gen.ApplicationMode) error {
	err := a.Node().Network().AddRoute(a.config.Gateway.ZoneNode, gen.NetworkRoute{
		Route: gen.Route{Host: a.config.Gateway.ZoneHost, Port: uint16(a.config.Gateway.ZonePort)},
	}, 100)
	if errors.Is(err, gen.ErrTaken) {
		return nil
	}
	return err
}

func (a *application) Terminate(error) {
	_ = a.Node().Network().RemoveRoute(a.config.Gateway.ZoneNode)
}
