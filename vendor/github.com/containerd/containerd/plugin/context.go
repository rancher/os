package plugin

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/log"
)

func NewContext(ctx context.Context, plugins map[PluginType]map[string]interface{}, root, state, id string) *InitContext {
	return &InitContext{
		plugins: plugins,
		Root:    filepath.Join(root, id),
		State:   filepath.Join(state, id),
		Context: log.WithModule(ctx, id),
	}
}

type InitContext struct {
	Root    string
	State   string
	Address string
	Context context.Context
	Config  interface{}
	Events  *events.Exchange

	plugins map[PluginType]map[string]interface{}
}

func (i *InitContext) Get(t PluginType) (interface{}, error) {
	for _, v := range i.plugins[t] {
		return v, nil
	}
	return nil, fmt.Errorf("no plugins registered for %s", t)
}

func (i *InitContext) GetAll(t PluginType) (map[string]interface{}, error) {
	p, ok := i.plugins[t]
	if !ok {
		return nil, fmt.Errorf("no plugins registered for %s", t)
	}
	return p, nil
}
