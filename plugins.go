package engine

import (
	"fmt"

	"github.com/cmskitdev/plugins"
	"github.com/cmskitdev/redis"
)

// Event is a type of event that can be published to the event bus.
type Event string

const (
	// EventInit is the event that is published when the plugin is initialized.
	EventInit Event = "init"
	// EventShutdown is the event that is published when the plugin is shutdown.
	EventShutdown Event = "shutdown"
	// EventDispatch is the event that is published when the plugin is dispatched.
	EventDispatch Event = "dispatch"
)

// RegisterPlugin registers a plugin with the engine.
func (e *Engine[T]) RegisterPlugin(name string, fn func() (plugins.Plugin[any], error)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	plugin, err := fn()
	if err != nil {
		fmt.Printf("failed to create plugin %s: %v", name, err)
		return
	}

	e.plugins[name] = &plugin
}

func (e *Engine[T]) registerAll() {
	e.RegisterPlugin("redis", func() (plugins.Plugin[any], error) {
		plugin, err := redis.NewPlugin(redis.Config{
			Address:      "127.0.0.1:6379",
			Username:     "foo",
			Password:     "bar",
			Database:     15,
			KeyPrefix:    "notion_codes",
			KeySeparator: ":",
			Workers:      10,
			MaxRetries:   3,
		}, e.bus)
		if err != nil {
			return nil, err
		}
		e.pluginRegistry.Register(plugin)

		// plugin.Receive(plugins.EventInit)
		return plugin, nil
	})
}
