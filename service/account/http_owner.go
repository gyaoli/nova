package account

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type httpOwner struct{ act.Actor }

func createHTTPOwner() gen.ProcessBehavior { return &httpOwner{} }

func (h *httpOwner) Init(args ...any) error {
	if len(args) != 1 {
		return fmt.Errorf("account http owner expects application argument")
	}
	a, ok := args[0].(*application)
	if !ok || a.owner == nil {
		return fmt.Errorf("account application is not initialized")
	}
	server, err := createWebServer(a.owner.host, a.owner.port, a.owner.handler)
	if err != nil {
		return fmt.Errorf("create account web server: %w", err)
	}
	webAlias, err := h.SpawnMeta(server, gen.MetaOptions{})
	if err != nil {
		server.close()
		return fmt.Errorf("spawn account web server: %w", err)
	}
	if err := h.Link(webAlias); err != nil {
		return fmt.Errorf("link account web server: %w", err)
	}
	return nil
}
