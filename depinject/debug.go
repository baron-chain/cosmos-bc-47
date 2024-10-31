package depinject

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"cosmossdk.io/depinject/internal/graphviz"
)

const (
	defaultDebugFile  = "debug_container.dot"
	defaultFilePerms  = 0o644
	defaultFontSize   = "12.0"
	defaultPenWidth   = "0.5"
	defaultGraphStyle = "rounded"
)

// Debug Configuration Types

type (
	// DebugOption configures debug logging and visualization output
	DebugOption interface {
		applyConfig(*debugConfig) error
	}

	debugConfig struct {
		loggers       []func(string)
		indentStr     string
		logBuf        *[]string
		graph         *graphviz.Graph
		visualizers   []func(string)
		logVisualizer bool
		onError       DebugOption
		onSuccess     DebugOption
		cleanup       []func()
	}

	debugOption func(*debugConfig) error
)

// Debug Option Constructors

// StdoutLogger routes logging output to stdout
func StdoutLogger() DebugOption {
	return Logger(func(s string) { fmt.Fprintln(os.Stdout, s) })
}

// StderrLogger routes logging output to stderr
func StderrLogger() DebugOption {
	return Logger(func(s string) { fmt.Fprintln(os.Stderr, s) })
}

// Visualizer provides a function to receive container rendering in Graphviz DOT format
func Visualizer(visualizer func(dotGraph string)) DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.addFuncVisualizer(visualizer)
		return nil
	})
}

// LogVisualizer dumps a graphviz DOT rendering to the log
func LogVisualizer() DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.enableLogVisualizer()
		return nil
	})
}

// FileVisualizer dumps a graphviz DOT rendering to the specified file
func FileVisualizer(filename string) DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.addFileVisualizer(filename)
		return nil
	})
}

// Logger provides a logging function for container messages
func Logger(logger func(string)) DebugOption {
	return debugOption(func(c *debugConfig) error {
		logger("Initializing logger")
		c.loggers = append(c.loggers, logger)
		c.sendBufferedLogs(logger)
		return nil
	})
}

// Debug sets default debug options (stderr output and file visualization)
func Debug() DebugOption {
	return DebugOptions(
		StderrLogger(),
		FileVisualizer(defaultDebugFile),
	)
}

// AutoDebug enables Debug on error and cleans up files on success
func AutoDebug() DebugOption {
	return DebugOptions(
		OnError(Debug()),
		OnSuccess(DebugCleanup(func() { deleteIfExists(defaultDebugFile) })),
	)
}

// Conditional Debug Options

// OnError sets debug options to apply only when an error occurs
func OnError(option DebugOption) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.initLogBuf()
		config.onError = option
		return nil
	})
}

// OnSuccess sets debug options to apply only on successful resolution
func OnSuccess(option DebugOption) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.initLogBuf()
		config.onSuccess = option
		return nil
	})
}

// DebugCleanup specifies cleanup functions for debug resources
func DebugCleanup(cleanup func()) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.cleanup = append(config.cleanup, cleanup)
		return nil
	})
}

// DebugOptions bundles multiple debug options together
func DebugOptions(options ...DebugOption) DebugOption {
	return debugOption(func(c *debugConfig) error {
		for _, opt := range options {
			if err := opt.applyConfig(c); err != nil {
				return err
			}
		}
		return nil
	})
}

// Debug Config Implementation

func (c debugOption) applyConfig(ctr *debugConfig) error { return c(ctr) }

func newDebugConfig() (*debugConfig, error) {
	return &debugConfig{graph: graphviz.NewGraph()}, nil
}

func (c *debugConfig) initLogBuf() {
	if c.logBuf == nil {
		c.logBuf = &[]string{}
		c.loggers = append(c.loggers, func(s string) {
			*c.logBuf = append(*c.logBuf, s)
		})
	}
}

func (c *debugConfig) sendBufferedLogs(logger func(string)) {
	if c.logBuf != nil {
		for _, s := range *c.logBuf {
			logger(s)
		}
	}
}

// Logging Methods

func (c *debugConfig) indentLogger()  { c.indentStr += " " }
func (c *debugConfig) dedentLogger()  { c.indentStr = c.indentStr[1:] }

func (c debugConfig) logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(c.indentStr+format, args...)
	for _, logger := range c.loggers {
		logger(msg)
	}
}

// Graph Management

func (c *debugConfig) generateGraph() {
	dotStr := c.graph.String()
	if c.logVisualizer {
		c.logf("DOT Graph: %s", dotStr)
	}
	for _, v := range c.visualizers {
		v(dotStr)
	}
}

func (c *debugConfig) addFuncVisualizer(f func(string)) {
	c.visualizers = append(c.visualizers, f)
}

func (c *debugConfig) enableLogVisualizer() {
	c.logVisualizer = true
}

func (c *debugConfig) addFileVisualizer(filename string) {
	c.visualizers = append(c.visualizers, func(_ string) {
		if err := c.saveGraphToFile(filename); err != nil {
			c.logf("Error saving graphviz file %s: %+v", filename, err)
		}
	})
}

func (c *debugConfig) saveGraphToFile(filename string) error {
	if err := os.WriteFile(filename, []byte(c.graph.String()), defaultFilePerms); err != nil {
		return err
	}
	
	if path, err := filepath.Abs(filename); err == nil {
		c.logf("Saved graph of container to %s", path)
	}
	return nil
}

// Node Management

func (c *debugConfig) locationGraphNode(location Location, key *moduleKey) *graphviz.Node {
	graph := c.moduleSubGraph(key)
	name := location.Name()
	node, found := graph.FindOrCreateNode(name)
	if !found {
		node.SetShape("box")
		setUnusedStyle(node.Attributes)
	}
	return node
}

func (c *debugConfig) typeGraphNode(typ reflect.Type) *graphviz.Node {
	name := formatTypeString(typ)
	node, found := c.graph.FindOrCreateNode(name)
	if !found {
		setUnusedStyle(node.Attributes)
	}
	return node
}

func (c *debugConfig) moduleSubGraph(key *moduleKey) *graphviz.Graph {
	if key == nil {
		return c.graph
	}
	
	gname := fmt.Sprintf("cluster_%s", key.name)
	graph, found := c.graph.FindOrCreateSubGraph(gname)
	if !found {
		graph.SetLabel(fmt.Sprintf("Module: %s", key.name))
		graph.SetPenWidth(defaultPenWidth)
		graph.SetFontSize(defaultFontSize)
		graph.SetStyle(defaultGraphStyle)
	}
	return graph
}

func (c *debugConfig) addGraphEdge(from, to *graphviz.Node) {
	_ = c.graph.CreateEdge(from, to)
}

// Helper Functions

func setUnusedStyle(attr *graphviz.Attributes) {
	attr.SetColor("lightgrey")
	attr.SetPenWidth(defaultPenWidth)
	attr.SetFontColor("dimgrey")
}

func formatTypeString(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Struct, reflect.Interface:
		return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	case reflect.Pointer:
		return fmt.Sprintf("*%s", formatTypeString(typ.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", 
			formatTypeString(typ.Key()), 
			formatTypeString(typ.Elem()))
	case reflect.Slice:
		return fmt.Sprintf("[]%s", formatTypeString(typ.Elem()))
	default:
		return typ.String()
	}
}

func deleteIfExists(filename string) {
	if _, err := os.Stat(filename); err == nil {
		_ = os.Remove(filename)
	}
}
