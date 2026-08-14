package role

import (
	"context"
	"errors"
	"testing"
	"time"
)

// repositoryStub isolates role service tests from MySQL.
type repositoryStub struct {
	createValue Role
	createRole  Role
	createErr   error
	findRole    Role
	findErr     error
	findCalls   int
}

func (r *repositoryStub) UpdateLastLogin(context.Context, int64, string, int32, int64) error {
	return nil
}

func (r *repositoryStub) Create(_ context.Context, value Role) (Role, error) {
	r.createValue = value
	return r.createRole, r.createErr
}

func (r *repositoryStub) FindByAccount(_ context.Context, _, _ string, _ int32) (Role, error) {
	r.findCalls++
	return r.findRole, r.findErr
}

func TestServiceCreateBuildsInitialRole(t *testing.T) {
	repository := &repositoryStub{createRole: Role{ID: 7, Account: "account-1"}}
	service := NewService(repository)
	service.now = func() time.Time { return time.Unix(1234, 0) }
	data := []byte{1, 2, 3}

	created, err := service.Create(context.Background(), CreateCommand{
		Account: " account-1 ", Platform: " test ", ZoneID: 2,
		Name: " role-name ", RegIP: "127.0.0.1", InitialData: data,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != 7 {
		t.Fatalf("Create() ID = %d, want 7", created.ID)
	}
	got := repository.createValue
	if got.Account != "account-1" || got.Platform != "test" || got.Name != "role-name" || got.ZoneID != 2 {
		t.Fatalf("Create() persisted identity = %#v", got)
	}
	if got.Level != 0 || got.RegTime != 1234 || len(got.Data) == 0 {
		t.Fatalf("Create() persisted initial state = %#v", got)
	}
	data[0] = 9
	if got.Data[0] != 1 {
		t.Fatal("Create() retained caller-owned mutable data")
	}
}

func TestServiceCreateIsIdempotentOnAccountConstraint(t *testing.T) {
	want := Role{ID: 8, Account: "account-1", Platform: "test", ZoneID: 2, Name: "existing"}
	repository := &repositoryStub{createErr: ErrAlreadyExists, findRole: want}
	service := NewService(repository)

	got, err := service.Create(context.Background(), CreateCommand{
		Account: "account-1", Platform: "test", ZoneID: 2,
		Name: "retry-name", InitialData: []byte{1},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ID != want.ID || repository.findCalls != 1 {
		t.Fatalf("Create() = %#v, find calls = %d", got, repository.findCalls)
	}
}

func TestServiceCreatePreservesNameConflict(t *testing.T) {
	repository := &repositoryStub{createErr: ErrNameConflict}
	service := NewService(repository)

	_, err := service.Create(context.Background(), CreateCommand{
		Account: "account-1", Platform: "test", ZoneID: 2,
		Name: "taken", InitialData: []byte{1},
	})
	if !errors.Is(err, ErrNameConflict) {
		t.Fatalf("Create() error = %v, want ErrNameConflict", err)
	}
	if repository.findCalls != 0 {
		t.Fatalf("Create() find calls = %d, want 0", repository.findCalls)
	}
}

func TestServiceRejectsEmptyInitialData(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.Create(context.Background(), CreateCommand{
		Account: "account-1", Platform: "test", Name: "role-name",
	})
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("Create() error = %v, want ErrInvalidArgument", err)
	}
}
