package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/state"
	"owlx/internal/zellij"
)

type App struct {
	Config config.Config
	Store  *state.Store
	Zellij *zellij.Client
}

var app *App

func loadApp(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	store := state.NewStore(cfg.StateDir)
	if err := store.Ensure(); err != nil {
		return fmt.Errorf("init state dir: %w", err)
	}
	zellijBin := zellij.ResolveBin(cfg.ZellijBin)
	app = &App{
		Config: cfg,
		Store:  store,
		Zellij: zellij.New(zellijBin),
	}
	return nil
}
