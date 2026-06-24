package promptjuggler

// RunOption sets an optional parameter on RunPrompt or RunWorkflow.
type RunOption func(*runConfig)

type runConfig struct {
	priority    *string
	thread      *string
	environment *string
	envVars     map[string]string
	metadata    map[string]any
	channel     *string
}

func applyRunOptions(opts []RunOption) runConfig {
	var c runConfig
	for _, o := range opts {
		o(&c)
	}
	return c
}

// WithPriority sets the processing priority: "onsite", "normal", or "low".
func WithPriority(priority string) RunOption {
	return func(c *runConfig) { c.priority = &priority }
}

// WithThread continues an existing conversation thread (a thread UUID).
func WithThread(thread string) RunOption {
	return func(c *runConfig) { c.thread = &thread }
}

// WithEnvironment overrides the environment the run executes in.
func WithEnvironment(environment string) RunOption {
	return func(c *runConfig) { c.environment = &environment }
}

// WithEnvVars supplies per-run environment variables (e.g. provider API keys).
func WithEnvVars(envVars map[string]string) RunOption {
	return func(c *runConfig) { c.envVars = envVars }
}

// WithMetadata attaches arbitrary metadata to the run. Values are string or []string.
func WithMetadata(metadata map[string]any) RunOption {
	return func(c *runConfig) { c.metadata = metadata }
}

// WithChannel sets the memory channel (prompt runs only; ignored by RunWorkflow).
func WithChannel(channel string) RunOption {
	return func(c *runConfig) { c.channel = &channel }
}
