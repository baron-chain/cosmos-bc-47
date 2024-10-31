package depinject_test

import (
	"fmt"
	"reflect"
	"testing"

	"cosmossdk.io/depinject"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pkgPath = "cosmossdk.io/depinject_test/depinject_test"
)

// Test Types

type Duck interface {
	quack()
}

type (
	Mallard    struct{}
	Canvasback struct{}
	Marbled    struct{}
)

func (Mallard) quack()    {}
func (Canvasback) quack() {}
func (Marbled) quack()    {}

type DuckWrapper struct {
	Module string
	Duck   Duck
}

func (DuckWrapper) IsManyPerContainerType() {}

type Pond struct {
	Ducks []DuckWrapper
}

// Test Suite

type bindingSuite struct {
	gocuke.TestingT
	configs []depinject.Config
	pond    *Pond
	err     error
}

func TestBindInterface(t *testing.T) {
	gocuke.NewRunner(t, &bindingSuite{}).
		Path("features/bindings.feature").
		Step(`we try to resolve a "Duck" in global scope`, (*bindingSuite).WeTryToResolveADuckInGlobalScope).
		Step(`module "(\w+)" wants a "Duck"`, (*bindingSuite).ModuleWantsADuck).
		Run()
}

// Provider Functions

func ProvideMallard() Mallard       { return Mallard{} }
func ProvideCanvasback() Canvasback { return Canvasback{} }
func ProvideMarbled() Marbled       { return Marbled{} }

func ProvideDuckWrapper(duck Duck) DuckWrapper {
	return DuckWrapper{Module: "", Duck: duck}
}

func ProvideModuleDuck(duck Duck, key depinject.OwnModuleKey) DuckWrapper {
	return DuckWrapper{Module: depinject.ModuleKey(key).Name(), Duck: duck}
}

func ResolvePond(ducks []DuckWrapper) Pond {
	return Pond{Ducks: ducks}
}

// Suite Methods

func (s *bindingSuite) resolvePond() *Pond {
	if s.pond != nil {
		return s.pond
	}

	s.addConfig(depinject.Provide(ResolvePond))
	var pond Pond
	s.err = depinject.Inject(depinject.Configs(s.configs...), &pond)
	s.pond = &pond
	return s.pond
}

func (s *bindingSuite) addConfig(config depinject.Config) {
	s.configs = append(s.configs, config)
}

func (s *bindingSuite) WeTryToResolveADuckInGlobalScope() {
	s.addConfig(depinject.Provide(ProvideDuckWrapper))
}

// Duck Type Providers

func (s *bindingSuite) IsProvided(duckType string) {
	providers := map[string]func(){
		"Mallard":    func() { s.addConfig(depinject.Provide(ProvideMallard)) },
		"Canvasback": func() { s.addConfig(depinject.Provide(ProvideCanvasback)) },
		"Marbled":    func() { s.addConfig(depinject.Provide(ProvideMarbled)) },
	}

	if provider, ok := providers[duckType]; ok {
		provider()
	} else {
		s.Fatalf("unexpected duck type %s", duckType)
	}
}

// Binding Methods

func (s *bindingSuite) ThereIsAGlobalBindingForA(preferredType, interfaceType string) {
	s.addConfig(depinject.BindInterface(
		getFullTypeName(interfaceType),
		getFullTypeName(preferredType),
	))
}

func (s *bindingSuite) ThereIsABindingForAInModule(preferredType, interfaceType, moduleName string) {
	s.addConfig(depinject.BindInterfaceInModule(
		moduleName,
		getFullTypeName(interfaceType),
		getFullTypeName(preferredType),
	))
}

func (s *bindingSuite) ModuleWantsADuck(module string) {
	s.addConfig(depinject.ProvideInModule(module, ProvideModuleDuck))
}

// Verification Methods

func (s *bindingSuite) IsResolvedInGlobalScope(typeName string) {
	pond := s.resolvePond()
	found := false
	for _, dw := range pond.Ducks {
		if dw.Module == "" {
			require.Contains(s, reflect.TypeOf(dw.Duck).Name(), typeName)
			found = true
		}
	}
	assert.True(s, found)
}

func (s *bindingSuite) ModuleResolvesA(module, duckType string) {
	pond := s.resolvePond()
	moduleFound := false
	for _, dw := range pond.Ducks {
		if dw.Module == module {
			assert.Contains(s, reflect.TypeOf(dw.Duck).Name(), duckType)
			moduleFound = true
		}
	}
	assert.True(s, moduleFound)
}

// Error Handling Methods

func (s *bindingSuite) ThereIsAError(expectedErrorMsg string) {
	s.resolvePond()
	assert.ErrorContains(s, s.err, expectedErrorMsg)
}

func (s *bindingSuite) ThereIsNoError() {
	s.resolvePond()
	assert.NoError(s, s.err)
}

// No-op Methods (defined at type level)

func (s bindingSuite) AnInterfaceDuck()                      {}
func (s bindingSuite) TwoImplementationsMallardAndCanvasback() {}

// Helper Functions

func getFullTypeName(typeName string) string {
	return fmt.Sprintf("%s.%s", pkgPath, typeName)
}
