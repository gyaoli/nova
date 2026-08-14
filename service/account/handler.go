package account

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const maxRequestBody = 8 << 10

const requestTimeout = 3 * time.Second

type Handler struct {
	service *Service
	limit   chan struct{}
}

func newHandler(service *Service, maxInFlight int) http.Handler {
	if maxInFlight <= 0 {
		maxInFlight = 128
	}
	h := &Handler{service: service, limit: make(chan struct{}, maxInFlight)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /account/register", h.register)
	mux.HandleFunc("POST /account/login", h.login)
	mux.HandleFunc("POST /account/verify", h.verify)
	mux.HandleFunc("GET /health/live", h.live)
	return h.admit(mux)
}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type registerRequest struct {
	AcctName   string `json:"acct_name"`
	Password   string `json:"password"`
	Platform   string `json:"platform"`
	ClientType int    `json:"client_type"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type loginRequest struct {
	AcctName string `json:"acct_name"`
	Password string `json:"password"`
}

type verifyRequest struct {
	Account string `json:"account"`
	Token   string `json:"token"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decode(w, r, &request); err != nil {
		writeError(w, ErrInvalidArgument)
		return
	}
	value, err := h.service.Register(r.Context(), RegisterCommand{
		AcctName: request.AcctName, Password: request.Password,
		Platform: request.Platform, ClientType: request.ClientType,
		DeviceID: request.DeviceID, DeviceName: request.DeviceName,
		RegIP: remoteIP(r.RemoteAddr),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response{Code: 0, Message: "success", Data: map[string]string{
		"account": value.Account, "acct_name": value.AcctName,
	}})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decode(w, r, &request); err != nil {
		writeError(w, ErrInvalidArgument)
		return
	}
	result, err := h.service.Login(r.Context(), LoginCommand{AcctName: request.AcctName, Password: request.Password})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "success", Data: map[string]any{
		"account": result.Account, "token": result.Token, "expires_in": int64(result.TTL / time.Second),
	}})
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	var request verifyRequest
	if err := decode(w, r, &request); err != nil {
		writeError(w, ErrInvalidArgument)
		return
	}
	if err := h.service.VerifyToken(r.Context(), request.Account, request.Token); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "success"})
}

func (h *Handler) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{Code: 0, Message: "ok"})
}

func (h *Handler) admit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case h.limit <- struct{}{}:
			defer func() { <-h.limit }()
			ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			writeJSON(w, http.StatusServiceUnavailable, response{Code: 1006, Message: "service busy"})
		}
	})
}

func decode(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one json value")
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status, code, message := http.StatusInternalServerError, 1005, "dependency error"
	switch {
	case errors.Is(err, ErrInvalidArgument):
		status, code, message = http.StatusBadRequest, 1001, "invalid argument"
	case errors.Is(err, ErrAccountExists):
		status, code, message = http.StatusConflict, 1002, "account already exists"
	case errors.Is(err, ErrAccountNotFound):
		status, code, message = http.StatusNotFound, 1003, "account not found"
	case errors.Is(err, ErrPasswordWrong):
		status, code, message = http.StatusUnauthorized, 1004, "password wrong"
	case errors.Is(err, ErrTokenInvalid):
		status, code, message = http.StatusUnauthorized, 1007, "token invalid"
	}
	writeJSON(w, status, response{Code: code, Message: message})
}

func writeJSON(w http.ResponseWriter, status int, value response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		if len(host) <= 16 {
			return host
		}
		return ""
	}
	remoteAddr = strings.TrimSpace(remoteAddr)
	if len(remoteAddr) <= 16 {
		return remoteAddr
	}
	return ""
}
