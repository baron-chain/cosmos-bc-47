package depinject_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/depinject"
)

// Test Types

type (
	KVStoreKey struct {
		name string
	}

	MsgClientA struct {
		key string
	}

	KeeperA struct {
		key  KVStoreKey
		name string
	}

	KeeperB struct {
		key        KVStoreKey
		msgClientA MsgClientA
	}

	KeeperC struct {
		key        KVStoreKey
		msgClientA MsgClientA
	}

	KeeperD struct {
		key KVStoreKey
	}
)

// Handler types
type Handler struct {
	Handle func()
}

func (Handler) IsOnePerModuleType() {}

type Command struct {
	Run func()
}

func (Command) IsManyPerContainerType() {}

// Provider Functions

func ProvideKVStoreKey(moduleKey depinject.ModuleKey) KVStoreKey {
	return KVStoreKey{name: moduleKey.Name()}
}

func ProvideMsgClientA(key depinject.ModuleKey) MsgClientA {
	return MsgClientA{key.Name()}
}

// Module Definitions

type (
	ModuleA struct{}
	ModuleB struct{}
	ModuleD struct{}

	BDependencies struct {
		depinject.In
		Key KVStoreKey
		A   MsgClientA
	}

	BProvides struct {
		depinject.Out
		KeeperB  KeeperB
		Commands []Command
	}

	DDependencies struct {
		depinject.In
		Key     KVStoreKey
		KeeperC KeeperC
	}

	DProvides struct {
		depinject.Out
		KeeperD  KeeperD
		Commands []Command
	}
)

func (ModuleA) Provide(key KVStoreKey, moduleKey depinject.OwnModuleKey) (KeeperA, Handler, Command) {
	return KeeperA{
		key:  key,
		name: depinject.ModuleKey(moduleKey).Name(),
	}, Handler{}, Command{}
}

func (ModuleB) Provide(deps BDependencies) (BProvides, Handler, error) {
	return BProvides{
		KeeperB: KeeperB{
			key:        deps.Key,
			msgClientA: deps.A,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}

func (ModuleD) Provide(deps DDependencies) (DProvides, Handler, error) {
	return DProvides{
		KeeperD: KeeperD{
			key: deps.Key,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}

// Common Test Configurations

var baseConfig = depinject.Configs(
	depinject.Provide(ProvideMsgClientA),
	depinject.ProvideInModule("runtime", ProvideKVStoreKey),
)

var scenarioConfig = depinject.Configs(
	baseConfig,
	depinject.ProvideInModule("a", ModuleA.Provide),
	depinject.ProvideInModule("b", ModuleB.Provide),
	depinject.Supply(ModuleA{}, ModuleB{}),
)

// Test Cases

func TestScenario(t *testing.T) {
	var (
		handlers map[string]Handler
		commands []Command
		a        KeeperA
		b        KeeperB
	)

	require.NoError(t, depinject.Inject(scenarioConfig, &handlers, &commands, &a, &b))

	// Verify handlers
	require.Len(t, handlers, 2)
	require.Equal(t, Handler{}, handlers["a"])
	require.Equal(t, Handler{}, handlers["b"])

	// Verify commands
	require.Len(t, commands, 3)

	// Verify KeeperA
	require.Equal(t, KeeperA{
		key:  KVStoreKey{name: "a"},
		name: "a",
	}, a)

	// Verify KeeperB
	require.Equal(t, KeeperB{
		key:        KVStoreKey{name: "b"},
		msgClientA: MsgClientA{key: "b"},
	}, b)
}

func TestResolutionErrors(t *testing.T) {
	tests := []struct {
		name   string
		config depinject.Config
		err    string
	}{
		{
			name: "dependency resolution error",
			config: depinject.Provide(
				func(x float64) string { return fmt.Sprintf("%f", x) },
				func(x int) float64 { return float64(x) },
				func(x float32) int { return int(x) },
			),
			err: "error",
		},
		{
			name: "cyclic dependency",
			config: depinject.Provide(
				func(x int) float64 { return float64(x) },
				func(x float64) (int, string) { return int(x), "hi" },
			),
			err: "error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var x string
			require.Error(t, depinject.Inject(tc.config, &x))
		})
	}
}

func TestDebugOptions(t *testing.T) {
	t.Run("logging and visualization", func(t *testing.T) {
		var logOut, dotGraph string

		// Setup temporary files
		outfile := setupTempFile(t, "out")
		graphfile := setupTempFile(t, "graph")
		defer cleanup(t, outfile, graphfile)

		// Redirect stdout
		stdout := os.Stdout
		os.Stdout = outfile
		defer func() { os.Stdout = stdout }()

		// Test debug options
		debugOpts := depinject.DebugOptions(
			depinject.Logger(func(s string) { logOut += s }),
			depinject.Visualizer(func(g string) { dotGraph = g }),
			depinject.LogVisualizer(),
			depinject.FileVisualizer(graphfile.Name()),
			depinject.StdoutLogger(),
		)

		require.NoError(t, depinject.InjectDebug(debugOpts, depinject.Configs()))
		verifyDebugOutput(t, logOut, dotGraph, outfile, graphfile)
	})
}

// Helper functions

func setupTempFile(t *testing.T, prefix string) *os.File {
	file, err := os.CreateTemp("", prefix)
	require.NoError(t, err)
	return file
}

func cleanup(t *testing.T, files ...*os.File) {
	for _, file := range files {
		require.NoError(t, os.Remove(file.Name()))
	}
}

func verifyDebugOutput(t *testing.T, logOut, dotGraph string, outfile, graphfile *os.File) {
	require.Contains(t, logOut, "digraph")
	require.Contains(t, dotGraph, "digraph")

	outfileContents, err := os.ReadFile(outfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(outfileContents), "digraph")

	graphfileContents, err := os.ReadFile(graphfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(graphfileContents), "digraph")
}
