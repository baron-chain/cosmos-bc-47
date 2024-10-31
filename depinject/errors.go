package depinject

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Error types for dependency injection
type (
	// ErrMultipleImplicitInterfaceBindings occurs when multiple implementations
	// are found for an interface with implicit binding
	ErrMultipleImplicitInterfaceBindings struct {
		error
		Interface reflect.Type   // The interface type
		Matches   []reflect.Type // Found implementations
	}

	// ErrNoTypeForExplicitBindingFound occurs when no provider is found
	// for an explicitly bound implementation
	ErrNoTypeForExplicitBindingFound struct {
		error
		Implementation string // The implementation type name
		Interface      string // The interface type name
		ModuleName     string // Optional module name
	}

	// ErrDuplicateDefinition occurs when the same type is provided multiple times
	ErrDuplicateDefinition struct {
		error
		Type         reflect.Type // The duplicated type
		NewLocation  Location     // Location of the duplicate definition
		OldLocation  string       // Location of the existing definition
	}
)

// Error constructors

// newErrMultipleImplicitInterfaceBindings creates an error for multiple interface implementations
func newErrMultipleImplicitInterfaceBindings(iface reflect.Type, matches map[reflect.Type]reflect.Type) ErrMultipleImplicitInterfaceBindings {
	implementations := make([]reflect.Type, 0, len(matches))
	for impl := range matches {
		implementations = append(implementations, impl)
	}
	return ErrMultipleImplicitInterfaceBindings{
		Interface: iface,
		Matches:   implementations,
	}
}

// newErrNoTypeForExplicitBindingFound creates an error for missing implementation
func newErrNoTypeForExplicitBindingFound(binding interfaceBinding) ErrNoTypeForExplicitBindingFound {
	var moduleName string
	if binding.moduleKey != nil {
		moduleName = binding.moduleKey.name
	}
	
	return ErrNoTypeForExplicitBindingFound{
		Implementation: binding.implTypeName,
		Interface:      binding.interfaceName,
		ModuleName:     moduleName,
	}
}

// newErrDuplicateDefinition creates an error for duplicate type definitions
func newErrDuplicateDefinition(typ reflect.Type, newLoc Location, oldLoc string) ErrDuplicateDefinition {
	return ErrDuplicateDefinition{
		Type:         typ,
		NewLocation:  newLoc,
		OldLocation:  oldLoc,
	}
}

// Error method implementations

func (e ErrMultipleImplicitInterfaceBindings) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Multiple implementations found for interface %v:", e.Interface))
	
	for _, match := range e.Matches {
		b.WriteString(fmt.Sprintf("\n  %s", fullyQualifiedTypeName(match)))
	}
	
	return b.String()
}

func (e ErrNoTypeForExplicitBindingFound) Error() string {
	if e.ModuleName != "" {
		return fmt.Sprintf(
			"No implementation found for explicit binding in module %q:\n"+
			"  Interface: %s\n"+
			"  Expected Implementation: %s",
			e.ModuleName, e.Interface, e.Implementation,
		)
	}
	
	return fmt.Sprintf(
		"No implementation found for explicit binding:\n"+
		"  Interface: %s\n"+
		"  Expected Implementation: %s",
		e.Interface, e.Implementation,
	)
}

func (e ErrDuplicateDefinition) Error() string {
	return fmt.Sprintf(
		"Duplicate provision of type %v:\n"+
		"  New definition at: %s\n"+
		"  Existing definition at: %s",
		e.Type, e.NewLocation, e.OldLocation,
	)
}

// Helper functions

// duplicateDefinitionError wraps the creation of ErrDuplicateDefinition
func duplicateDefinitionError(typ reflect.Type, newLoc Location, oldLoc string) error {
	err := newErrDuplicateDefinition(typ, newLoc, oldLoc)
	return errors.WithStack(err)
}

// Custom error checking

// IsMultipleImplicitBindingsError checks if an error is ErrMultipleImplicitInterfaceBindings
func IsMultipleImplicitBindingsError(err error) bool {
	_, ok := err.(ErrMultipleImplicitInterfaceBindings)
	return ok
}

// IsNoTypeForBindingError checks if an error is ErrNoTypeForExplicitBindingFound
func IsNoTypeForBindingError(err error) bool {
	_, ok := err.(ErrNoTypeForExplicitBindingFound)
	return ok
}

// IsDuplicateDefinitionError checks if an error is ErrDuplicateDefinition
func IsDuplicateDefinitionError(err error) bool {
	_, ok := err.(ErrDuplicateDefinition)
	return ok
}
