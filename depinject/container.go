package depinject

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"cosmossdk.io/depinject/internal/graphviz"
)

var (
	stringType = reflect.TypeOf("")

	ErrCyclicDependency     = errors.New("cyclic dependency detected")
	ErrModuleScopeRequired  = errors.New("module scope required for this operation")
	ErrInvalidOutputType    = errors.New("output type must be a pointer")
	ErrDuplicateModuleScope = errors.New("duplicate module-scoped dependencies")
	ErrInvalidInvoker       = errors.New("invoker function should not return any outputs")
)

// container manages dependency injection and resolution
type container struct {
	*debugConfig
	resolvers         map[string]resolver
	interfaceBindings map[string]interfaceBinding
	invokers          []invoker
	moduleKeyContext  *ModuleKeyContext
	resolveStack      []resolveFrame
	callerStack      []Location
	callerMap        map[Location]bool
}

type (
	invoker struct {
		fn     *providerDescriptor
		modKey *moduleKey
	}

	resolveFrame struct {
		loc Location
		typ reflect.Type
	}

	interfaceBinding struct {
		interfaceName string
		implTypeName  string
		moduleKey     *moduleKey
		resolver      resolver
	}
)

// newContainer creates a new dependency injection container
func newContainer(cfg *debugConfig) *container {
	return &container{
		debugConfig:       cfg,
		resolvers:         make(map[string]resolver),
		moduleKeyContext:  &ModuleKeyContext{},
		interfaceBindings: make(map[string]interfaceBinding),
		callerMap:        make(map[Location]bool),
	}
}

// Provider Resolution

func (c *container) call(provider *providerDescriptor, moduleKey *moduleKey) ([]reflect.Value, error) {
	loc := provider.Location
	graphNode := c.locationGraphNode(loc, moduleKey)
	markGraphNodeAsFailed(graphNode)

	if err := c.checkCyclicDependency(loc); err != nil {
		return nil, err
	}

	c.pushCaller(loc)
	defer c.popCaller(loc)

	c.logf("Resolving dependencies for %s", loc)
	c.indentLogger()
	inVals, err := c.resolveInputs(provider.Inputs, moduleKey, loc)
	c.dedentLogger()
	if err != nil {
		return nil, err
	}

	c.logf("Calling %s", loc)
	out, err := provider.Fn(inVals)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling provider %s", loc)
	}

	markGraphNodeAsUsed(graphNode)
	return out, nil
}

func (c *container) resolveInputs(inputs []providerInput, moduleKey *moduleKey, loc Location) ([]reflect.Value, error) {
	inVals := make([]reflect.Value, len(inputs))
	for i, in := range inputs {
		val, err := c.resolve(in, moduleKey, loc)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
	return inVals, nil
}

func (c *container) checkCyclicDependency(loc Location) error {
	if c.callerMap[loc] {
		return errors.Wrapf(ErrCyclicDependency, "%s -> %s", loc.Name(), loc.Name())
	}
	return nil
}

func (c *container) pushCaller(loc Location) {
	c.callerMap[loc] = true
	c.callerStack = append(c.callerStack, loc)
}

func (c *container) popCaller(loc Location) {
	delete(c.callerMap, loc)
	c.callerStack = c.callerStack[:len(c.callerStack)-1]
}

// Resolver Management

func (c *container) getResolver(typ reflect.Type, key *moduleKey) (resolver, error) {
	if r, err := c.getExplicitResolver(typ, key); err != nil || r != nil {
		return r, err
	}

	if r, ok := c.resolverByType(typ); ok {
		return r, nil
	}

	elemType := c.getElementType(typ)
	if elemType == typ {
		return c.resolveInterfaceType(typ)
	}

	return c.createContainerResolver(elemType, typ)
}

func (c *container) getElementType(typ reflect.Type) reflect.Type {
	if isManyPerContainerSliceType(typ) || isOnePerModuleMapType(typ) {
		return typ.Elem()
	}
	return typ
}

func (c *container) resolveInterfaceType(typ reflect.Type) (resolver, error) {
	if typ.Kind() != reflect.Interface {
		return nil, nil
	}

	matches := c.findImplementingTypes(typ)
	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		for resolverType := range matches {
			res, _ := c.resolverByType(resolverType)
			c.logf("Implicitly registering resolver %v for interface type %v", resolverType, typ)
			c.addResolver(typ, res)
			return res, nil
		}
	default:
		return nil, newErrMultipleImplicitInterfaceBindings(typ, matches)
	}

	return nil, nil // unreachable but required by compiler
}

func (c *container) findImplementingTypes(interfaceType reflect.Type) map[reflect.Type]reflect.Type {
	matches := make(map[reflect.Type]reflect.Type)
	for _, r := range c.resolvers {
		resolverType := r.getType()
		if resolverType.Kind() != reflect.Interface && resolverType.Implements(interfaceType) {
			matches[resolverType] = resolverType
		}
	}
	return matches
}

// Node Management

func (c *container) addNode(provider *providerDescriptor, key *moduleKey) (interface{}, error) {
	providerGraphNode := c.locationGraphNode(provider.Location, key)

	if err := c.validateProviderInputs(provider, key); err != nil {
		return nil, err
	}

	hasModuleKey := c.hasModuleKeyParam(provider)
	if !hasModuleKey {
		return c.addSimpleNode(provider, key, providerGraphNode)
	}

	return c.addModuleScopedNode(provider, providerGraphNode)
}

func (c *container) validateProviderInputs(provider *providerDescriptor, key *moduleKey) error {
	for _, in := range provider.Inputs {
		if err := c.validateInput(in.Type, key); err != nil {
			return err
		}

		if err := c.addInputTypeToGraph(in.Type, provider, key); err != nil {
			return err
		}
	}
	return nil
}

func (c *container) addInputTypeToGraph(typ reflect.Type, provider *providerDescriptor, key *moduleKey) error {
	vr, err := c.getResolver(typ, key)
	if err != nil {
		return err
	}

	var typeGraphNode *graphviz.Node
	if vr != nil {
		typeGraphNode = vr.typeGraphNode()
	} else {
		typeGraphNode = c.typeGraphNode(typ)
	}

	c.addGraphEdge(typeGraphNode, c.locationGraphNode(provider.Location, key))
	return nil
}

// Helper Functions

func markGraphNodeAsUsed(node *graphviz.Node) {
	node.SetColor("black")
	node.SetPenWidth("1.5")
	node.SetFontColor("black")
}

func markGraphNodeAsFailed(node *graphviz.Node) {
	node.SetColor("red")
	node.SetFontColor("red")
}

func fullyQualifiedTypeName(typ reflect.Type) string {
	pkgType := typ
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || 
	   typ.Kind() == reflect.Map || typ.Kind() == reflect.Array {
		pkgType = typ.Elem()
	}
	pkgPath := pkgType.PkgPath()
	if pkgPath == "" {
		return fmt.Sprintf("%v", typ)
	}
	return fmt.Sprintf("%s/%v", pkgPath, typ)
}

// Binding Management

func bindingKeyFromTypeName(typeName string, key *moduleKey) string {
	if key == nil {
		return fmt.Sprintf("%s;", typeName)
	}
	return fmt.Sprintf("%s;%s", typeName, key.name)
}

func bindingKeyFromType(typ reflect.Type, key *moduleKey) string {
	return bindingKeyFromTypeName(fullyQualifiedTypeName(typ), key)
}

func (c *container) addBinding(p interfaceBinding) {
	c.interfaceBindings[bindingKeyFromTypeName(p.interfaceName, p.moduleKey)] = p
}

func (c *container) addResolver(typ reflect.Type, r resolver) {
	c.resolvers[fullyQualifiedTypeName(typ)] = r
}

func (c *container) resolverByType(typ reflect.Type) (resolver, bool) {
	return c.resolverByTypeName(fullyQualifiedTypeName(typ))
}

func (c *container) resolverByTypeName(typeName string) (resolver, bool) {
	res, found := c.resolvers[typeName]
	return res, found
}
