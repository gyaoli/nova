package resource

import (
	"errors"
	"testing"

	configenv "nova/core/env"

	"ergo.services/ergo/gen"
)

func TestDisabledResourcesAreUnavailable(t *testing.T) {
	application := NewApplication(configenv.NodeConfig{}, Options{})
	if _, err := application.Acquire(Options{MySQL: true}); !errors.Is(err, ErrMySQLUnavailable) {
		t.Fatalf("Acquire(MySQL) error = %v", err)
	}
	if _, err := application.Acquire(Options{Redis: true}); !errors.Is(err, ErrRedisUnavailable) {
		t.Fatalf("Acquire(Redis) error = %v", err)
	}
}

func TestTerminateWaitsForLeases(t *testing.T) {
	application := NewApplication(configenv.NodeConfig{}, Options{})
	lease, err := application.Acquire(Options{})
	if err != nil {
		t.Fatal(err)
	}
	application.Terminate(nil)
	if application.leases != 1 || !application.closing {
		t.Fatalf("state after Terminate = leases:%d closing:%v", application.leases, application.closing)
	}
	lease.Release()
	if application.leases != 0 {
		t.Fatalf("leases after Release = %d", application.leases)
	}
	if _, err := application.Acquire(Options{}); !errors.Is(err, gen.ErrApplicationStopping) {
		t.Fatalf("Acquire after Terminate error = %v", err)
	}
}
