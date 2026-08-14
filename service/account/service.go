package account

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrAccountExists     = errors.New("account already exists")
	ErrAccountIDConflict = errors.New("account id conflict")
	ErrAccountNotFound   = errors.New("account not found")
	ErrPasswordWrong     = errors.New("password wrong")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrTokenInvalid      = errors.New("token invalid")
)

type Account struct {
	AcctName   string
	Password   string
	Account    string
	Platform   string
	ClientType int
	DeviceID   string
	DeviceName string
	RegIP      string
	RegTime    int64
}

type RegisterCommand struct {
	AcctName   string
	Password   string
	Platform   string
	ClientType int
	DeviceID   string
	DeviceName string
	RegIP      string
}

type LoginCommand struct {
	AcctName string
	Password string
}

type Repository interface {
	Create(ctx context.Context, value Account) error
	FindByName(ctx context.Context, accountName string) (Account, error)
}

type TokenStore interface {
	Replace(ctx context.Context, account, token string, ttl time.Duration) error
	Verify(ctx context.Context, account, token string) error
	Revoke(ctx context.Context, account, token string) error
}

type IDGenerator interface {
	NewID() (string, error)
}

type TokenGenerator interface {
	NewToken() (string, error)
}

type LoginResult struct {
	Account string
	Token   string
	TTL     time.Duration
}

// Service implements the account domain operations.
type Service struct {
	repository     Repository
	tokens         TokenStore
	ids            IDGenerator
	tokenGenerator TokenGenerator
	tokenTTL       time.Duration
}

func NewService(repository Repository, tokens TokenStore, ids IDGenerator, tokenGenerator TokenGenerator, tokenTTL time.Duration) *Service {
	return &Service{repository: repository, tokens: tokens, ids: ids, tokenGenerator: tokenGenerator, tokenTTL: tokenTTL}
}

func (s *Service) Register(ctx context.Context, command RegisterCommand) (Account, error) {
	command.AcctName = strings.TrimSpace(command.AcctName)
	if command.AcctName == "" || command.Password == "" || command.Platform == "" || command.ClientType < 0 {
		return Account{}, ErrInvalidArgument
	}

	for attempt := 0; attempt < 3; attempt++ {
		id, err := s.ids.NewID()
		if err != nil {
			return Account{}, err
		}
		value := Account{
			AcctName: command.AcctName, Password: command.Password, Account: id,
			Platform: command.Platform, ClientType: command.ClientType,
			DeviceID: command.DeviceID, DeviceName: command.DeviceName,
			RegIP: command.RegIP, RegTime: time.Now().Unix(),
		}
		err = s.repository.Create(ctx, value)
		switch {
		case err == nil:
			return value, nil
		case errors.Is(err, ErrAccountIDConflict):
			continue
		default:
			return Account{}, err
		}
	}
	return Account{}, ErrAccountIDConflict
}

func (s *Service) Login(ctx context.Context, command LoginCommand) (LoginResult, error) {
	command.AcctName = strings.TrimSpace(command.AcctName)
	if command.AcctName == "" || command.Password == "" {
		return LoginResult{}, ErrInvalidArgument
	}
	value, err := s.repository.FindByName(ctx, command.AcctName)
	if err != nil {
		return LoginResult{}, err
	}
	if value.Password != command.Password {
		return LoginResult{}, ErrPasswordWrong
	}
	token, err := s.tokenGenerator.NewToken()
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.tokens.Replace(ctx, value.Account, token, s.tokenTTL); err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Account: value.Account, Token: token, TTL: s.tokenTTL}, nil
}

func (s *Service) VerifyToken(ctx context.Context, accountID, token string) error {
	if accountID == "" || token == "" {
		return ErrTokenInvalid
	}
	return s.tokens.Verify(ctx, accountID, token)
}
