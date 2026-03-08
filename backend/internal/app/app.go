package app

import (
	"fmt"
	"net/http"

	"github.com/erpang/post-sync/internal/api"
	"github.com/erpang/post-sync/internal/config"
	"github.com/erpang/post-sync/internal/db"
)

type App struct {
	config config.Config
	server *http.Server
}

func New() (*App, error) {
	cfg := config.Load()
	database, err := db.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &App{
		config: cfg,
		server: &http.Server{
			Addr:         cfg.ServerAddr,
			Handler:      api.NewRouter(database, cfg),
			ReadTimeout:  cfg.HTTPReadTimeout,
			WriteTimeout: cfg.HTTPWriteTimeout,
		},
	}, nil
}

func (a *App) Run() error {
	fmt.Printf("post-sync backend listening on %s\n", a.config.ServerAddr)
	return a.server.ListenAndServe()
}
