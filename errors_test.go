package soop

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorLevelString(t *testing.T) {
	tests := []struct {
		level    ErrorLevel
		expected string
	}{
		{ErrorLevelDefault, "UNKNOWN"},
		{ErrorLevelDebug, "DEBUG"},
		{ErrorLevelInfo, "INFO"},
		{ErrorLevelWarning, "WARNING"},
		{ErrorLevelError, "ERROR"},
		{ErrorLevelCritical, "CRITICAL"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.level.String(), "Expected string representation of ErrorLevel to match")
	}
}

func TestError_DefaultLevel(t *testing.T) {
	inputItem := 123
	e := NewError(1, ErrorLevelDefault, "default level error", nil, &inputItem)
	assert.NotNil(t, e, "Expected NewError to return a non-nil error")
	assert.Equal(t, DefaultErrorLevel, e.Level, "Expected error level to default to DefaultErrorLevel")
}

func TestError_CustomLevel(t *testing.T) {
	inputItem := "test input"
	e := NewError(2, ErrorLevelWarning, "custom level error", nil, &inputItem)
	assert.NotNil(t, e, "Expected NewError to return a non-nil error")
	assert.Equal(t, ErrorLevelWarning, e.Level, "Expected error level to match provided level")
}

func TestError_ErrorMethod(t *testing.T) {
	inputItem := "test input"
	err := errors.New("underlying error")
	e := NewError(3, ErrorLevelError, "error with underlying error", err, &inputItem)
	expectedMessage := "[ERROR] error with underlying error: underlying error (input: test input)"
	assert.Equal(t, expectedMessage, e.Error(), "Expected error message to match")

	eNoUnderlying := NewError(4, ErrorLevelInfo, "error without underlying error", nil, &inputItem)
	expectedMessageNoUnderlying := "[INFO] error without underlying error (input: test input)"
	assert.Equal(t, expectedMessageNoUnderlying, eNoUnderlying.Error(), "Expected error message to match")
}

func TestError_UnwrapMethod(t *testing.T) {
	err := errors.New("underlying error")
	e := NewError(5, ErrorLevelCritical, "critical error", err, &TestInputType{})
	assert.Equal(t, err, e.Unwrap(), "Expected Unwrap to return the underlying error")
}

func TestError_NilError(t *testing.T) {
	var e *Error[int]
	assert.Equal(t, "<nil>", e.Error(), "Expected <nil> string for nil error")
}

func TestError_WithoutInputItem(t *testing.T) {
	e := NewError[*string](6, ErrorLevelDebug, "debug error", nil, nil)
	expectedMessage := "[DEBUG] debug error (input: <nil>)"
	assert.Equal(t, expectedMessage, e.Error(), "Expected error message to handle nil input item")
}
