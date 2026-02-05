package app

import (
	"context"
	"fmt"

	"owlx/internal/config"
	"owlx/internal/tmux"
)

type State struct {
	Cfg  config.Config
	Tmux tmux.Manager
}

type appKey struct{}

func Initialize() (*State, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	tm, err := tmux.New(cfg)
	if err != nil {
		return nil, err
	}
	if err := tmux.EnsureInitialized(cfg, tm); err != nil {
		return nil, err
	}
	return &State{Cfg: cfg, Tmux: tm}, nil
}

func WithContext(ctx context.Context, state *State) context.Context {
	return context.WithValue(ctx, appKey{}, state)
}

func FromContext(ctx context.Context) (*State, error) {
	val := ctx.Value(appKey{})
	state, ok := val.(*State)
	if !ok || state == nil {
		return nil, fmt.Errorf("missing app context")
	}
	return state, nil
}
