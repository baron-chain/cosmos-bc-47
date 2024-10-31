package depinject

import (
	"reflect"

	"github.com/pkg/errors"
)

// Common errors
var (
	ErrEmptyModuleName = errors.New("expected non-empty module name")
)

// Config defines a functional configuration of a container
type Config interface {
	apply(*container) error
}

type containerConfig func(*container) error

func (c containerConfig) apply(ctr *container) error {
	return c(ctr)
}

// Provide registers dependency injection providers in global scope.
// Providers:
// - Are called at most once (except module-scoped providers)
// - Must be exported functions from non-internal packages
// - Must have exported input/output types from non-internal packages
// - Should have exported generic type parameters (not checked)
func Provide(providers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		return provide(ctr, nil, providers)
	})
}

// ProvideInModule registers dependency injection providers in a specific module scope.
// See Provide for provider requirements.
func ProvideInModule(moduleName string, providers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		if moduleName == "" {
			return ErrEmptyModuleName
		}
		return provide(ctr, ctr.moduleKeyContext.createOrGetModuleKey(moduleName), providers)
	})
}

// Invoke registers invoker functions to run in global scope after dependency graph configuration.
// Invokers:
// - Are called in registration order
// - May only return error as output
// - Have all inputs marked as optional
// - Should nil-check all inputs
// - Must be exported functions from non-internal packages
// - Must have exported input types from non-internal packages
// - Should have exported generic type parameters (not checked)
func Invoke(invokers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		return invoke(ctr, nil, invokers)
	})
}

// InvokeInModule registers invoker functions to run in a specific module scope.
// See Invoke for invoker requirements.
func InvokeInModule(moduleName string, invokers ...interface{}) Config {
	return containerConfig(func(ctr *container) error {
		if moduleName == "" {
			return ErrEmptyModuleName
		}
		return invoke(ctr, ctr.moduleKeyContext.createOrGetModuleKey(moduleName), invokers)
	})
}

// BindInterface defines a global scope interface binding.
// Example:
//
//	BindInterface(
//	    "pkg/path.Duck",    // interface
//	    "pkg/path.DuckImpl" // implementation
//	)
func BindInterface(inTypeName, outTypeName string) Config {
	return containerConfig(func(ctr *container) error {
		return bindInterface(ctr, inTypeName, outTypeName, "")
	})
}

// BindInterfaceInModule defines a module-scoped interface binding.
// Example:
//
//	BindInterfaceInModule(
//	    "myModule",         // module name
//	    "pkg/path.Duck",    // interface
//	    "pkg/path.DuckImpl" // implementation
//	)
func BindInterfaceInModule(moduleName, inTypeName, outTypeName string) Config {
	return containerConfig(func(ctr *container) error {
		return bindInterface(ctr, inTypeName, outTypeName, moduleName)
	})
}

// Supply registers concrete values directly into the container
func Supply(values ...interface{}) Config {
	loc := LocationFromCaller(1)
	return containerConfig(func(ctr *container) error {
		for _, v := range values {
			if err := ctr.supply(reflect.ValueOf(v), loc); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

// Error registers an error that will cause container initialization to fail
func Error(err error) Config {
	return containerConfig(func(*container) error {
		return errors.WithStack(err)
	})
}

// Configs bundles multiple Config definitions into a single Config
func Configs(configs ...Config) Config {
	return containerConfig(func(ctr *container) error {
		for _, cfg := range configs {
			if err := cfg.apply(ctr); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

// Helper functions

func provide(ctr *container, key *moduleKey, providers []interface{}) error {
	for _, provider := range providers {
		desc, err := extractProviderDescriptor(provider)
		if err != nil {
			return errors.WithStack(err)
		}
		if _, err = ctr.addNode(&desc, key); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func invoke(ctr *container, key *moduleKey, invokers []interface{}) error {
	for _, invoker := range invokers {
		desc, err := extractInvokerDescriptor(invoker)
		if err != nil {
			return errors.WithStack(err)
		}
		if err = ctr.addInvoker(&desc, key); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func bindInterface(ctr *container, inTypeName, outTypeName, moduleName string) error {
	var mk *moduleKey
	if moduleName != "" {
		mk = &moduleKey{name: moduleName}
	}
	
	ctr.addBinding(interfaceBinding{
		interfaceName: inTypeName,
		implTypeName:  outTypeName,
		moduleKey:     mk,
	})
	
	return nil
}
