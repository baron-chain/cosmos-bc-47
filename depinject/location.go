// Copyright (c) 2019 Uber Technologies, Inc.
// Copyright information preserved as in original file...

package depinject

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
)

const (
	vendorPrefix = "/vendor/"
	defaultSkip  = 1
)

// Location provides information about a code location
type Location interface {
	isLocation()
	Name() string
	fmt.Stringer
	fmt.Formatter
}

// location implements Location with file, function and package information
type location struct {
	name string // function name
	pkg  string // package path
	file string // file path
	line int    // line number
}

// LocationFromPC creates a Location from a program counter
func LocationFromPC(pc uintptr) Location {
	funcInfo := runtime.FuncForPC(pc)
	if funcInfo == nil {
		return newEmptyLocation()
	}

	pkgName, funcName := splitFuncName(funcInfo.Name())
	fileName, lineNum := funcInfo.FileLine(pc)

	return &location{
		name: funcName,
		pkg:  pkgName,
		file: fileName,
		line: lineNum,
	}
}

// LocationFromCaller creates a Location from the caller's stack frame
func LocationFromCaller(skip int) Location {
	pc, _, _, ok := runtime.Caller(skip + defaultSkip)
	if !ok {
		return newEmptyLocation()
	}
	return LocationFromPC(pc)
}

// Implementation of Location interface

func (f *location) isLocation() {}

func (f *location) String() string {
	return fmt.Sprint(f)
}

func (f *location) Name() string {
	return formatFullName(f.pkg, f.name)
}

func (f *location) Format(w fmt.State, c rune) {
	if w.Flag('+') && c == 'v' {
		f.formatMultiLine(w)
	} else {
		f.formatSingleLine(w)
	}
}

// Helper methods for formatting

func (f *location) formatMultiLine(w fmt.State) {
	// "path/to/package".MyFunction
	//     path/to/file.go:42
	_, _ = fmt.Fprintf(w, "%s.%s", f.pkg, f.name)
	_, _ = fmt.Fprintf(w, "\n\t%s:%d", f.file, f.line)
}

func (f *location) formatSingleLine(w fmt.State) {
	// "path/to/package".MyFunction (path/to/file.go:42)
	_, _ = fmt.Fprintf(w, "%s.%s (%s:%d)", f.pkg, f.name, f.file, f.line)
}

// Helper functions

// splitFuncName splits a function's full name into package and function parts
func splitFuncName(function string) (packageName, functionName string) {
	if function == "" {
		return "", ""
	}

	lastSlash := strings.LastIndex(function, "/")
	if lastSlash < 0 {
		lastSlash = 0
	}

	firstDot := strings.Index(function[lastSlash:], ".")
	if firstDot < 0 {
		return function, ""
	}

	idx := lastSlash + firstDot
	packageName, functionName = function[:idx], function[idx+1:]

	// Handle vendored packages
	if vendorIdx := strings.Index(packageName, vendorPrefix); vendorIdx > 0 {
		packageName = packageName[vendorIdx+len(vendorPrefix):]
	}

	// URL-decode package name to handle special cases (e.g., .git in package name)
	if unescaped, err := url.QueryUnescape(packageName); err == nil {
		packageName = unescaped
	}

	return packageName, functionName
}

// formatFullName formats the package and function name into a fully qualified name
func formatFullName(pkg, name string) string {
	return fmt.Sprintf("%v.%v", pkg, name)
}

// newEmptyLocation creates an empty location for error cases
func newEmptyLocation() Location {
	return &location{
		name: "unknown",
		pkg:  "unknown",
		file: "unknown",
		line: 0,
	}
}
