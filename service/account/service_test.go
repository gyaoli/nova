package account

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

type repositoryStub struct {
	account Account
	create  func(Account) error
}

func (r repositoryStub) Create(_ context.Context, value Account) error {
	if r.create != nil {
		return r.create(value)
	}
	return nil
}

func (r repositoryStub) FindByName(context.Context, string) (Account, error) {
	return r.account, nil
}

type sequentialGenerator struct {
	mu   sync.Mutex
	next int
}

func (g *sequentialGenerator) NewID() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return fmt.Sprintf("id-%d", g.next), nil
}

func (g *sequentialGenerator) NewToken() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return fmt.Sprintf("token-%d", g.next), nil
}

type memoryTokens struct {
	mu      sync.RWMutex
	current map[string]string
}

func newMemoryTokens() *memoryTokens { return &memoryTokens{current: make(map[string]string)} }

func (m *memoryTokens) Replace(_ context.Context, accountID, token string, _ time.Duration) error {
	m.mu.Lock()
	m.current[accountID] = token
	m.mu.Unlock()
	return nil
}

func (m *memoryTokens) Verify(_ context.Context, accountID, token string) error {
	m.mu.RLock()
	current := m.current[accountID]
	m.mu.RUnlock()
	if current != token {
		return ErrTokenInvalid
	}
	return nil
}

func (m *memoryTokens) Revoke(_ context.Context, accountID, token string) error {
	m.mu.Lock()
	if m.current[accountID] == token {
		delete(m.current, accountID)
	}
	m.mu.Unlock()
	return nil
}

func TestRegisterRetriesAccountIDConflict(t *testing.T) {
	created := 0
	repository := repositoryStub{create: func(Account) error {
		created++
		if created == 1 {
			return ErrAccountIDConflict
		}
		return nil
	}}
	generator := &sequentialGenerator{}
	service := NewService(repository, newMemoryTokens(), generator, generator, time.Hour)

	value, err := service.Register(context.Background(), RegisterCommand{
		AcctName: "tester", Password: "123456", Platform: "dev", ClientType: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created != 2 || value.Account != "id-2" {
		t.Fatalf("created=%d account=%q", created, value.Account)
	}
}

func TestConcurrentLoginLeavesOnlyOneUsableToken(t *testing.T) {
	tokens := newMemoryTokens()
	generator := &sequentialGenerator{}
	service := NewService(repositoryStub{account: Account{Account: "account-1", Password: "pw"}}, tokens, generator, generator, time.Hour)

	const count = 100
	results := make(chan LoginResult, count)
	errorsCh := make(chan error, count)
	var group sync.WaitGroup
	for i := 0; i < count; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := service.Login(context.Background(), LoginCommand{AcctName: "tester", Password: "pw"})
			if err != nil {
				errorsCh <- err
				return
			}
			results <- result
		}()
	}
	group.Wait()
	close(results)
	close(errorsCh)
	for err := range errorsCh {
		t.Fatal(err)
	}

	valid := 0
	for result := range results {
		err := service.VerifyToken(context.Background(), result.Account, result.Token)
		if err == nil {
			valid++
			continue
		}
		if !errors.Is(err, ErrTokenInvalid) {
			t.Fatal(err)
		}
	}
	if valid != 1 {
		t.Fatalf("valid tokens=%d, want 1", valid)
	}
}
