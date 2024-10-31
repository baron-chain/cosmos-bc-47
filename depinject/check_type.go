package depinject

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// ErrTypeNotExported indicates that a type is not exported (doesn't start with a capital letter)
var ErrTypeNotExported = errors.New("type must be exported")

// ErrInternalPackage indicates that a type comes from an internal package
var ErrInternalPackage = errors.New("type must not come from an internal package")

// isExportedType verifies that a type is exported and not from an internal package.
// For composite types, it recursively checks their element types.
// Note: Generic type parameters are not checked due to reflection limitations.
func isExportedType(typ reflect.Type) error {
	// Check named types first
	if isNamedType(typ) {
		return checkNamedType(typ)
	}

	// Handle composite types
	return checkCompositeType(typ)
}

// isNamedType returns true if the type has both a name and package path
func isNamedType(typ reflect.Type) bool {
	return typ.Name() != "" && typ.PkgPath() != ""
}

// checkNamedType validates that a named type is exported and not from an internal package
func checkNamedType(typ reflect.Type) error {
	if !isExported(typ.Name()) {
		return errors.Wrapf(ErrTypeNotExported, "%s", typ)
	}

	if isInternalPackage(typ.PkgPath()) {
		return errors.Wrapf(ErrInternalPackage, "%s", typ)
	}

	return nil
}

// checkCompositeType handles validation for composite types like slices, maps, etc.
func checkCompositeType(typ reflect.Type) error {
	switch typ.Kind() {
	case reflect.Array, reflect.Slice, reflect.Chan, reflect.Pointer:
		return isExportedType(typ.Elem())

	case reflect.Map:
		if err := isExportedType(typ.Key()); err != nil {
			return err
		}
		return isExportedType(typ.Elem())

	case reflect.Func:
		return checkFuncType(typ)

	default:
		// Built-in, non-composite types (integers, etc.) are always valid
		return nil
	}
}

// checkFuncType validates all input and output types of a function
func checkFuncType(typ reflect.Type) error {
	// Check input parameters
	for i := 0; i < typ.NumIn(); i++ {
		if err := isExportedType(typ.In(i)); err != nil {
			return err
		}
	}

	// Check output parameters
	for i := 0; i < typ.NumOut(); i++ {
		if err := isExportedType(typ.Out(i)); err != nil {
			return err
		}
	}

	return nil
}

// isExported checks if the first character of a name is uppercase
func isExported(name string) bool {
	return !unicode.IsLower([]rune(name)[0])
}

// isInternalPackage checks if the package path contains an "internal" directory
func isInternalPackage(pkgPath string) bool {
	return slices.Contains(strings.Split(pkgPath, "/"), "internal")
}
