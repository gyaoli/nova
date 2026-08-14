package account

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ergo.services/ergo/gen"
)

const (
	httpReadHeaderTimeout = 2 * time.Second
	httpReadTimeout       = 5 * time.Second
	httpWriteTimeout      = 5 * time.Second
	httpIdleTimeout       = 60 * time.Second
	httpMaxHeaderBytes    = 16 << 10
)

// webServer keeps blocking HTTP accept/serve work inside an Ergo meta process.
type webServer struct {
	gen.MetaProcess
	listener net.Listener
	server   http.Server
}

func createWebServer(host string, port uint16, handler http.Handler) (*webServer, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}
	return &webServer{
		listener: listener,
		server: http.Server{
			Handler:           handler,
			ReadHeaderTimeout: httpReadHeaderTimeout,
			ReadTimeout:       httpReadTimeout,
			WriteTimeout:      httpWriteTimeout,
			IdleTimeout:       httpIdleTimeout,
			MaxHeaderBytes:    httpMaxHeaderBytes,
		},
	}, nil
}

func (w *webServer) Init(process gen.MetaProcess) error {
	w.MetaProcess = process
	w.server.ErrorLog = log.New(w, "", 0)
	w.Log().Info("account http server started on %s", w.listener.Addr())
	return nil
}

func (w *webServer) Start() error {
	err := w.server.Serve(w.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (w *webServer) HandleMessage(gen.PID, any) error { return nil }

func (w *webServer) HandleCall(gen.PID, gen.Ref, any) (any, error) { return nil, nil }

func (w *webServer) Terminate(error) { w.close() }

func (w *webServer) HandleInspect(gen.PID, ...string) map[string]string {
	return map[string]string{"listener": w.listener.Addr().String()}
}

func (w *webServer) Write(message []byte) (int, error) {
	w.Log().Error(strings.TrimSpace(string(message)))
	return len(message), nil
}

func (w *webServer) close() {
	_ = w.server.Close()
	_ = w.listener.Close()
}
