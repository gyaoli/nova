package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestReplaceVerifyAndConditionalRevoke(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := newTokenStore(client)
	ctx := context.Background()

	if err := store.Replace(ctx, "account-1", "old-token", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := store.Replace(ctx, "account-1", "new-token", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := store.Verify(ctx, "account-1", "old-token"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("old token verify error=%v", err)
	}
	if err := store.Verify(ctx, "account-1", "new-token"); err != nil {
		t.Fatal(err)
	}

	if err := store.Revoke(ctx, "account-1", "old-token"); err != nil {
		t.Fatal(err)
	}
	if err := store.Verify(ctx, "account-1", "new-token"); err != nil {
		t.Fatalf("old revoke removed new token: %v", err)
	}
	if err := store.Revoke(ctx, "account-1", "new-token"); err != nil {
		t.Fatal(err)
	}
	if err := store.Verify(ctx, "account-1", "new-token"); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("revoked token verify error=%v", err)
	}
}
