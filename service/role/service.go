package role

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound        = errors.New("role not found")
	ErrAlreadyExists   = errors.New("role already exists")
	ErrNameConflict    = errors.New("role name conflict")
	ErrInvalidArgument = errors.New("invalid argument")
)

// Role mirrors the persistent columns of the role table. Data is an opaque,
// serialized snapshot owned by the role domain.
type Role struct {
	ID             int64
	Platform       string
	ZoneID         int32
	Account        string
	Name           string
	Level          int8
	RegIP          string
	RegTime        int64
	Data           []byte
	LastUpdateTime time.Time
	LastLoginTime  int64
	LastLogoutTime int64
}

type CreateCommand struct {
	Platform    string
	ZoneID      int32
	Account     string
	Name        string
	RegIP       string
	InitialData []byte
}

type Repository interface {
	Create(ctx context.Context, value Role) (Role, error)
	FindByAccount(ctx context.Context, account, platform string, zoneID int32) (Role, error)
	UpdateLastLogin(ctx context.Context, id int64, platform string, zoneID int32, loginTime int64) error
}

// Service implements the role domain operations.
type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

// Create is idempotent for the database identity (account, platform, zone_id).
// The repository's unique constraint is the final arbiter for concurrent calls.
func (s *Service) Create(ctx context.Context, command CreateCommand) (Role, error) {
	command.Account = strings.TrimSpace(command.Account)
	command.Platform = strings.TrimSpace(command.Platform)
	command.Name = strings.TrimSpace(command.Name)
	if command.Account == "" || command.Platform == "" || command.Name == "" || command.ZoneID < 0 || len(command.InitialData) == 0 {
		return Role{}, ErrInvalidArgument
	}

	value := Role{
		Platform: command.Platform,
		ZoneID:   command.ZoneID,
		Account:  command.Account,
		Name:     command.Name,
		RegIP:    command.RegIP,
		RegTime:  s.now().Unix(),
		Data:     append([]byte(nil), command.InitialData...),
	}
	created, err := s.repository.Create(ctx, value)
	if err == nil {
		return created, nil
	}
	if !errors.Is(err, ErrAlreadyExists) {
		return Role{}, err
	}

	// A concurrent request may have committed the same role. Returning that
	// row makes retrying the create request safe without hiding name conflicts.
	return s.repository.FindByAccount(ctx, command.Account, command.Platform, command.ZoneID)
}

func (s *Service) FindByAccount(ctx context.Context, account, platform string, zoneID int32) (Role, error) {
	account = strings.TrimSpace(account)
	platform = strings.TrimSpace(platform)
	if account == "" || platform == "" || zoneID < 0 {
		return Role{}, ErrInvalidArgument
	}
	return s.repository.FindByAccount(ctx, account, platform, zoneID)
}

func (s *Service) Login(ctx context.Context, account, platform string, zoneID int32) (Role, error) {
	value, err := s.FindByAccount(ctx, account, platform, zoneID)
	if err != nil {
		return Role{}, err
	}
	value.LastLoginTime = s.now().Unix()
	if err := s.repository.UpdateLastLogin(ctx, value.ID, value.Platform, value.ZoneID, value.LastLoginTime); err != nil {
		return Role{}, err
	}
	return value, nil
}
