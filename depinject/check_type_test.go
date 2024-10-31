package depinject

import (
	"os"
	"reflect"
	"testing"

	"cosmossdk.io/depinject/internal/graphviz"
	"gotest.tools/v3/assert"
)

func TestCheckIsExportedType(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		shouldError bool
		errMsg      string
	}{
		// Primitive types
		{name: "bool", value: false},
		{name: "string", value: ""},
		{name: "uintptr", value: uintptr(0)},
		
		// Unsigned integers
		{name: "uint", value: uint(0)},
		{name: "uint8", value: uint8(0)},
		{name: "uint16", value: uint16(0)},
		{name: "uint32", value: uint32(0)},
		{name: "uint64", value: uint64(0)},
		
		// Signed integers
		{name: "int", value: int(0)},
		{name: "int8", value: int8(0)},
		{name: "int16", value: int16(0)},
		{name: "int32", value: int32(0)},
		{name: "int64", value: int64(0)},
		
		// Floating point and complex
		{name: "float32", value: float32(0)},
		{name: "float64", value: float64(0)},
		{name: "complex64", value: complex64(0)},
		{name: "complex128", value: complex128(0)},
		
		// System types
		{name: "FileMode", value: os.FileMode(0)},
		
		// Arrays and slices
		{name: "array", value: [1]int{0}},
		{name: "slice", value: []int{}},
		
		// Channels
		{name: "bidirectional channel", value: make(chan int)},
		{name: "receive channel", value: make(<-chan int)},
		{name: "send channel", value: make(chan<- int)},
		
		// Functions
		{name: "simple function", value: func(int, string) (bool, error) { return false, nil }},
		{name: "variadic function", value: func(int, ...string) (bool, error) { return false, nil }},
		
		// Maps and pointers
		{name: "map", value: map[string]In{}},
		{name: "pointer", value: &In{}},
		{name: "nil pointer", value: (*Location)(nil)},
		{name: "struct", value: In{}},
		
		// Invalid types
		{
			name:        "unexported struct",
			value:       container{},
			shouldError: true,
			errMsg:      "must be exported",
		},
		{
			name:        "unexported struct pointer",
			value:       &container{},
			shouldError: true,
			errMsg:      "must be exported",
		},
		{
			name:        "internal package struct",
			value:       graphviz.Attributes{},
			shouldError: true,
			errMsg:      "internal",
		},
		{
			name:        "map with internal package value",
			value:       map[string]graphviz.Attributes{},
			shouldError: true,
			errMsg:      "internal",
		},
		{
			name:        "slice with internal package type",
			value:       []graphviz.Attributes{},
			shouldError: true,
			errMsg:      "internal",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := isExportedType(reflect.TypeOf(tc.value))
			if tc.shouldError {
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

// Helper functions for backwards compatibility
func expectValidType(t *testing.T, v interface{}) {
	t.Helper()
	assert.NilError(t, isExportedType(reflect.TypeOf(v)))
}

func expectInvalidType(t *testing.T, v interface{}, errContains string) {
	t.Helper()
	assert.ErrorContains(t, isExportedType(reflect.TypeOf(v)), errContains)
}
