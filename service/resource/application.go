package resource

import (
	"context"
	"errors"
	"sync"
	"time"

	configenv "nova/core/env"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const ApplicationName gen.Atom = "resource"

var (
	ErrMySQLUnavailable = errors.New("mysql resource is unavailable")
	ErrRedisUnavailable = errors.New("redis resource is unavailable")
)

type Options struct {
	MySQL bool
	Redis bool
}

type Provider interface {
	Acquire(Options) (*Lease, error)
}

type Application struct {
	app.Application
	config  configenv.NodeConfig
	options Options
	mu      sync.Mutex
	mysql   *MySQLClient
	redis   *RedisClient
	leases  int
	closing bool
}

type Lease struct {
	owner *Application
	mysql *MySQLClient
	redis *RedisClient
	once  sync.Once
}

func NewApplication(config configenv.NodeConfig, options Options) *Application {
	return &Application{config: config, options: options}
}

func (a *Application) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name: ApplicationName, Description: "Nova node resource pools",
		Version:     gen.Version{Name: "resource", Release: "1.0.0"},
		Mode:        gen.ApplicationModePermanent,
		Group:       []gen.ApplicationMemberSpec{{Factory: createSupervisor, Name: supervisorName, Args: []any{a}}},
		InitTimeout: 5 * time.Second, StopTimeout: 5 * time.Second,
	}, nil
}

func (a *Application) Init(gen.Ref, gen.ApplicationMode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	a.mu.Lock()
	a.closing = false
	a.mu.Unlock()

	if a.options.MySQL {
		client, err := openMySQL(ctx, MySQLConfig{
			Host: a.config.Mysql.Host, Port: a.config.Mysql.Port, User: a.config.Mysql.User,
			Password: a.config.Mysql.Password, Database: a.config.Mysql.DB,
			MaxOpenConns: a.config.Mysql.MaxOpenConns, MaxIdleConns: a.config.Mysql.MaxIdleConns,
			ConnMaxLifetime: a.config.Mysql.ConnMaxLifetime, ConnMaxIdleTime: a.config.Mysql.ConnMaxIdleTime,
		})
		if err != nil {
			return err
		}
		a.mysql = client
	}

	if a.options.Redis {
		client, err := openRedis(ctx, RedisConfig{
			Address: a.config.Redis.Address, Username: a.config.Redis.Username,
			Password: a.config.Redis.Password, DB: a.config.Redis.DB, PoolSize: a.config.Redis.PoolSize,
			DialTimeout: a.config.Redis.DialTimeout, ReadTimeout: a.config.Redis.ReadTimeout,
			WriteTimeout: a.config.Redis.WriteTimeout,
		})
		if err != nil {
			if a.mysql != nil {
				_ = a.mysql.Close()
				a.mysql = nil
			}
			return err
		}
		a.redis = client
	}
	return nil
}

func (a *Application) Terminate(error) {
	a.mu.Lock()
	a.closing = true
	if a.leases == 0 {
		a.closeLocked()
	}
	a.mu.Unlock()
}

func (a *Application) Acquire(options Options) (*Lease, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closing {
		return nil, gen.ErrApplicationStopping
	}
	if options.MySQL && a.mysql == nil {
		return nil, ErrMySQLUnavailable
	}
	if options.Redis && a.redis == nil {
		return nil, ErrRedisUnavailable
	}
	a.leases++
	return &Lease{owner: a, mysql: a.mysql, redis: a.redis}, nil
}

func (l *Lease) MySQL() (*MySQLClient, error) {
	if l == nil || l.mysql == nil {
		return nil, ErrMySQLUnavailable
	}
	return l.mysql, nil
}

func (l *Lease) Redis() (*RedisClient, error) {
	if l == nil || l.redis == nil {
		return nil, ErrRedisUnavailable
	}
	return l.redis, nil
}

func (l *Lease) Release() {
	if l == nil || l.owner == nil {
		return
	}
	l.once.Do(func() {
		l.owner.release()
	})
}

func (a *Application) release() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.leases > 0 {
		a.leases--
	}
	if a.closing && a.leases == 0 {
		a.closeLocked()
	}
}

func (a *Application) closeLocked() {
	if a.redis != nil {
		_ = a.redis.Close()
		a.redis = nil
	}
	if a.mysql != nil {
		_ = a.mysql.Close()
		a.mysql = nil
	}
}
