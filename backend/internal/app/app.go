package app

import (
	"fmt"
	"net/http"

	"github.com/erpang/post-sync/internal/api"
	"github.com/erpang/post-sync/internal/config"
)

type App struct {
	config config.Config
	server *http.Server
}

func New() (*App, error) {
	cfg := config.Load()

	return &App{
		config: cfg,
		server: &http.Server{
			Addr:         cfg.ServerAddr,
			Handler:      api.NewRouter(),
			ReadTimeout:  cfg.HTTPReadTimeout,
			WriteTimeout: cfg.HTTPWriteTimeout,
		},
	}, nil
}

func (a *App) Run() error {
	fmt.Printf("post-sync backend listening on %s\n", a.config.ServerAddr)
	return a.server.ListenAndServe()
}
