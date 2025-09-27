package generator

import (
	"github.com/lefeck/gonginx/config"
)

// ConfigBuilder provides a fluent interface for building nginx configurations
type ConfigBuilder struct {
	config *config.Config
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &config.Config{
			Block: &config.Block{
				Directives: []config.IDirective{},
			},
		},
	}
}

// Build returns the built configuration
func (cb *ConfigBuilder) Build() *config.Config {
	return cb.config
}

// AddDirective adds a custom directive to the configuration
func (cb *ConfigBuilder) AddDirective(name string, parameters ...string) *ConfigBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	cb.config.Block.AddDirective(directive)
	return cb
}

// AddComment adds a comment to the configuration
func (cb *ConfigBuilder) AddComment(comment string) *ConfigBuilder {
	// For now, we'll add comments as special directives
	// In a real implementation, you might want to handle this differently
	directive := &config.Directive{
		Name:    "#",
		Comment: []string{comment},
	}

	cb.config.Block.AddDirective(directive)
	return cb
}

// HTTP creates and returns an HTTP block builder
func (cb *ConfigBuilder) HTTP() *HTTPBuilder {
	httpDirective := &config.Directive{
		Name:  "http",
		Block: &config.Block{Directives: []config.IDirective{}},
	}

	httpBlock, _ := config.NewHTTP(httpDirective)
	cb.config.Block.AddDirective(httpBlock)

	return &HTTPBuilder{
		config:    cb.config,
		httpBlock: httpBlock,
	}
}

// Stream creates and returns a Stream block builder
func (cb *ConfigBuilder) Stream() *StreamBuilder {
	streamDirective := &config.Directive{
		Name:  "stream",
		Block: &config.Block{Directives: []config.IDirective{}},
	}

	streamBlock, _ := config.NewStream(streamDirective)
	cb.config.Block.AddDirective(streamBlock)

	return &StreamBuilder{
		config:      cb.config,
		streamBlock: streamBlock,
	}
}

// Events creates an events block with common settings
func (cb *ConfigBuilder) Events() *EventsBuilder {
	eventsDirective := &config.Directive{
		Name:  "events",
		Block: &config.Block{Directives: []config.IDirective{}},
	}

	cb.config.Block.AddDirective(eventsDirective)

	return &EventsBuilder{
		config:      cb.config,
		eventsBlock: eventsDirective,
	}
}

// WorkerProcesses sets the worker_processes directive
func (cb *ConfigBuilder) WorkerProcesses(value string) *ConfigBuilder {
	return cb.AddDirective("worker_processes", value)
}

// WorkerConnections is a shortcut for events { worker_connections N; }
func (cb *ConfigBuilder) WorkerConnections(value string) *ConfigBuilder {
	return cb.Events().WorkerConnections(value).End()
}

// ErrorLog sets the error_log directive
func (cb *ConfigBuilder) ErrorLog(path string, level ...string) *ConfigBuilder {
	if len(level) > 0 {
		return cb.AddDirective("error_log", path, level[0])
	}
	return cb.AddDirective("error_log", path)
}

// PidFile sets the pid directive
func (cb *ConfigBuilder) PidFile(path string) *ConfigBuilder {
	return cb.AddDirective("pid", path)
}

// EventsBuilder provides methods for building events block
type EventsBuilder struct {
	config      *config.Config
	eventsBlock *config.Directive
}

// WorkerConnections sets worker_connections in events block
func (eb *EventsBuilder) WorkerConnections(value string) *EventsBuilder {
	directive := &config.Directive{
		Name:       "worker_connections",
		Parameters: []config.Parameter{config.NewParameter(value)},
	}
	eb.eventsBlock.Block.(*config.Block).AddDirective(directive)
	return eb
}

// UseEpoll enables epoll
func (eb *EventsBuilder) UseEpoll() *EventsBuilder {
	directive := &config.Directive{
		Name:       "use",
		Parameters: []config.Parameter{config.NewParameter("epoll")},
	}
	eb.eventsBlock.Block.(*config.Block).AddDirective(directive)
	return eb
}

// MultiAccept enables multi_accept
func (eb *EventsBuilder) MultiAccept(enabled bool) *EventsBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	directive := &config.Directive{
		Name:       "multi_accept",
		Parameters: []config.Parameter{config.NewParameter(value)},
	}
	eb.eventsBlock.Block.(*config.Block).AddDirective(directive)
	return eb
}

// End returns to the main config builder
func (eb *EventsBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: eb.config}
}
