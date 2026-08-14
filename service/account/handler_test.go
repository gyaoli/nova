package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type verifyTokens struct{ token string }

func (v verifyTokens) Replace(context.Context, string, string, time.Duration) error { return nil }
func (v verifyTokens) Revoke(context.Context, string, string) error                 { return nil }
func (v verifyTokens) Verify(_ context.Context, _ string, token string) error {
	if token != v.token {
		return ErrTokenInvalid
	}
	return nil
}

func TestVerifyEndpoint(t *testing.T) {
	service := NewService(nil, verifyTokens{token: "valid"}, nil, nil, time.Hour)
	handler := newHandler(service, 1)

	request := httptest.NewRequest(http.MethodPost, "/account/verify", strings.NewReader(`{"account":"account-1","token":"valid"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("valid token status = %d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/account/verify", strings.NewReader(`{"account":"account-1","token":"wrong"}`))
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d body=%s", response.Code, response.Body.String())
	}
}
