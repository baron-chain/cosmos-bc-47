package depinject

import (
	"fmt"
)

// Error definitions
var (
	ErrInvalidDebugConfig = fmt.Errorf("failed to create debug configuration")
	ErrProviderRegistration = fmt.Errorf("failed to register providers")
)

// InjectionOptions holds the configuration for injection
type InjectionOptions struct {
	location   Location
	debugOpt   DebugOption
	config     Config
	outputs    []interface{}
}

// Inject builds and runs a dependency injection container.
// It uses AutoDebug mode which provides verbose debug info on error.
//
// Parameters:
// - containerConfig: The container configuration
// - outputs: Pointer values to be filled by the container
//
// Example:
//
//	var x int
//	err := Inject(Provide(func() int { return 1 }), &x)
func Inject(containerConfig Config, outputs ...interface{}) error {
	opts := InjectionOptions{
		location: LocationFromCaller(1),
		debugOpt: AutoDebug(),
		config:   containerConfig,
		outputs:  outputs,
	}
	return runInjection(opts)
}

// InjectDebug is like Inject but with configurable debug options.
//
// Parameters:
// - debugOpt: Debug configuration for logging and visualization
// - config: The container configuration
// - outputs: Pointer values to be filled by the container
func InjectDebug(debugOpt DebugOption, config Config, outputs ...interface{}) error {
	opts := InjectionOptions{
		location: LocationFromCaller(1),
		debugOpt: debugOpt,
		config:   config,
		outputs:  outputs,
	}
	return runInjection(opts)
}

// runInjection handles the main injection process
func runInjection(opts InjectionOptions) error {
	cfg, err := setupDebugConfig()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDebugConfig, err)
	}

	// Ensure cleanup and graph generation on function exit
	defer func() {
		cfg.generateGraph()
		runCleanup(cfg)
	}()

	// Run injection process
	if err := performInjection(cfg, opts); err != nil {
		return handleInjectionError(cfg, err)
	}

	return handleInjectionSuccess(cfg)
}

// setupDebugConfig creates and validates the debug configuration
func setupDebugConfig() (*debugConfig, error) {
	cfg, err := newDebugConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// performInjection executes the main injection logic
func performInjection(cfg *debugConfig, opts InjectionOptions) error {
	if opts.debugOpt != nil {
		if err := opts.debugOpt.applyConfig(cfg); err != nil {
			return err
		}
	}

	return buildContainer(cfg, opts)
}

// buildContainer creates and configures the dependency container
func buildContainer(cfg *debugConfig, opts InjectionOptions) error {
	cfg.logf("Registering providers")
	cfg.indentLogger()
	defer cfg.dedentLogger()

	container := newContainer(cfg)
	
	if err := opts.config.apply(container); err != nil {
		cfg.logf("Failed registering providers: %+v", err)
		return fmt.Errorf("%w: %v", ErrProviderRegistration, err)
	}

	return container.build(opts.location, opts.outputs...)
}

// handleInjectionError processes errors during injection
func handleInjectionError(cfg *debugConfig, err error) error {
	cfg.logf("Error: %v", err)
	
	if cfg.onError != nil {
		if err2 := cfg.onError.applyConfig(cfg); err2 != nil {
			return fmt.Errorf("error handling failed: %v (original error: %v)", err2, err)
		}
	}
	
	return err
}

// handleInjectionSuccess handles successful injection
func handleInjectionSuccess(cfg *debugConfig) error {
	if cfg.onSuccess != nil {
		if err := cfg.onSuccess.applyConfig(cfg); err != nil {
			return fmt.Errorf("success handling failed: %v", err)
		}
	}
	return nil
}

// runCleanup executes cleanup functions
func runCleanup(cfg *debugConfig) {
	for _, cleanup := range cfg.cleanup {
		cleanup()
	}
}
