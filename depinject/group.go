package depinject

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"cosmossdk.io/depinject/internal/graphviz"
)

// Error definitions
var (
	ErrInvalidManyPerContainerType = errors.New("invalid many-per-container type usage")
)

// ManyPerContainerType marks a type which automatically gets grouped together.
// For a ManyPerContainerType T:
// - T and []T can be declared as output parameters multiple times
// - All provided values can be retrieved using []T input parameter
type ManyPerContainerType interface {
	IsManyPerContainerType() // Marker function
}

// Type validation functions

var manyPerContainerTypeType = reflect.TypeOf((*ManyPerContainerType)(nil)).Elem()

func isManyPerContainerType(t reflect.Type) bool {
	return t.Implements(manyPerContainerTypeType)
}

func isManyPerContainerSliceType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice && isManyPerContainerType(typ.Elem())
}

// Resolver types

type (
	// groupResolver handles resolution for many-per-container types
	groupResolver struct {
		typ          reflect.Type      // Base type
		sliceType    reflect.Type      // Slice type for collection
		idxsInValues []int            // Indices in provider values
		providers    []*simpleProvider // Providers for this type
		resolved     bool             // Resolution status
		values       reflect.Value     // Resolved values
		graphNode    *graphviz.Node   // Graph visualization node
	}

	// sliceGroupResolver wraps groupResolver for slice operations
	sliceGroupResolver struct {
		*groupResolver
	}
)

// Resolver interface implementations

func (g *groupResolver) getType() reflect.Type {
	return g.sliceType
}

func (g *groupResolver) describeLocation() string {
	return fmt.Sprintf("many-per-container type %v", g.typ)
}

func (g *groupResolver) typeGraphNode() *graphviz.Node {
	return g.graphNode
}

// Resolution methods

func (g *sliceGroupResolver) resolve(c *container, _ *moduleKey, caller Location) (reflect.Value, error) {
	g.logResolution(c, caller)
	
	if !g.resolved {
		if err := g.resolveValues(c); err != nil {
			return reflect.Value{}, err
		}
	}
	
	return g.values, nil
}

func (g *groupResolver) resolve(_ *container, _ *moduleKey, _ Location) (reflect.Value, error) {
	return reflect.Value{}, errors.Wrapf(ErrInvalidManyPerContainerType,
		"%v is a many-per-container type and cannot be used as an input value, use %v instead",
		g.typ, g.sliceType)
}

// Resolution helper methods

func (g *sliceGroupResolver) logResolution(c *container, caller Location) {
	c.logf("Providing many-per-container type slice %v to %s from:", g.sliceType, caller.Name())
	c.indentLogger()
	for _, node := range g.providers {
		c.logf(node.provider.Location.String())
	}
	c.dedentLogger()
}

func (g *groupResolver) resolveValues(c *container) error {
	result := reflect.MakeSlice(g.sliceType, 0, len(g.providers))

	for i, provider := range g.providers {
		values, err := provider.resolveValues(c)
		if err != nil {
			return err
		}

		value := values[g.idxsInValues[i]]
		result = g.appendValue(result, value)
	}

	g.values = result
	g.resolved = true
	return nil
}

func (g *groupResolver) appendValue(slice, value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Slice {
		return g.appendSlice(slice, value)
	}
	return reflect.Append(slice, value)
}

func (g *groupResolver) appendSlice(slice, values reflect.Value) reflect.Value {
	n := values.Len()
	for i := 0; i < n; i++ {
		slice = reflect.Append(slice, values.Index(i))
	}
	return slice
}

// Provider management

func (g *groupResolver) addNode(provider *simpleProvider, idx int) error {
	g.providers = append(g.providers, provider)
	g.idxsInValues = append(g.idxsInValues, idx)
	return nil
}

// Helper functions for creating resolvers

func newGroupResolver(typ reflect.Type) *groupResolver {
	return &groupResolver{
		typ:       typ,
		sliceType: reflect.SliceOf(typ),
	}
}

func newSliceGroupResolver(base *groupResolver) *sliceGroupResolver {
	return &sliceGroupResolver{
		groupResolver: base,
	}
}
