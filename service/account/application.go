package account

import (
	"net/http"
	"time"

	"nova/service/resource"

	configenv "nova/core/env"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	applicationName gen.Atom = "account"
)

type application struct {
	app.Application
	config    configenv.NodeConfig
	resources resource.Provider
	lease     *resource.Lease
	owner     *httpOwnerDependencies
}

type httpOwnerDependencies struct {
	host    string
	port    uint16
	handler http.Handler
}

func NewApplication(config configenv.NodeConfig, resources resource.Provider) gen.ApplicationBehavior {
	return &application{config: config, resources: resources}
}

func (a *application) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name: applicationName, Description: "Nova development account service",
		Version:     gen.Version{Name: "account", Release: "1.0.0"},
		Mode:        gen.ApplicationModePermanent,
		Depends:     gen.ApplicationDepends{Applications: []gen.Atom{resource.ApplicationName}},
		Group:       []gen.ApplicationMemberSpec{{Factory: createSupervisor, Name: supervisorName, Args: []any{a}}},
		InitTimeout: 5 * time.Second, StopTimeout: 5 * time.Second,
	}, nil
}

func (a *application) Init(ref gen.Ref, mode gen.ApplicationMode) error {
	lease, err := a.resources.Acquire(resource.Options{MySQL: true, Redis: true})
	if err != nil {
		return err
	}
	mysqlClient, err := lease.MySQL()
	if err != nil {
		lease.Release()
		return err
	}
	redisClient, err := lease.Redis()
	if err != nil {
		lease.Release()
		return err
	}

	a.lease = lease
	service := NewService(newRepository(mysqlClient.DB()), newTokenStore(redisClient.Command()), RandomIDGenerator{}, RandomTokenGenerator{}, a.config.Account.TokenTTL)
	a.owner = &httpOwnerDependencies{
		host: a.config.Host, port: uint16(a.config.WebPort),
		handler: newHandler(service, a.config.Account.HTTPMaxInFlight),
	}
	return nil
}

func (a *application) Terminate(error) {
	if a.lease != nil {
		a.lease.Release()
		a.lease = nil
	}
}
