package account

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const defaultPrefix = "nova:account:token:"

var revokeScript = goredis.NewScript(`
local current = redis.call("GET", KEYS[1])
if not current or current ~= ARGV[1] then
    return 0
end
return redis.call("DEL", KEYS[1])
`)

type redisTokenStore struct {
	client goredis.UniversalClient
	prefix string
}

func newTokenStore(client goredis.UniversalClient) *redisTokenStore {
	return &redisTokenStore{client: client, prefix: defaultPrefix}
}

func (s *redisTokenStore) Replace(ctx context.Context, accountID, rawToken string, ttl time.Duration) error {
	return s.client.Set(ctx, s.key(accountID), tokenHash(rawToken), ttl).Err()
}

func (s *redisTokenStore) Verify(ctx context.Context, accountID, rawToken string) error {
	current, err := s.client.Get(ctx, s.key(accountID)).Result()
	if errors.Is(err, goredis.Nil) {
		return ErrTokenInvalid
	}
	if err != nil {
		return err
	}
	want := tokenHash(rawToken)
	if len(current) != len(want) || subtle.ConstantTimeCompare([]byte(current), []byte(want)) != 1 {
		return ErrTokenInvalid
	}
	return nil
}

func (s *redisTokenStore) Revoke(ctx context.Context, accountID, rawToken string) error {
	_, err := revokeScript.Run(ctx, s.client, []string{s.key(accountID)}, tokenHash(rawToken)).Result()
	return err
}

func (s *redisTokenStore) key(accountID string) string { return s.prefix + accountID }

func tokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
