package soop

import (
	"fmt"
)

// ErrorLevel represents the severity of an error.
type ErrorLevel int

const (
	// ErrorLevelDefault selects the default error level.
	ErrorLevelDefault ErrorLevel = iota
	// ErrorLevelDebug represents a very granular message.
	ErrorLevelDebug
	// ErrorLevelInfo represents an informational message.
	ErrorLevelInfo
	// ErrorLevelWarning represents a warning message.
	ErrorLevelWarning
	// ErrorLevelError represents an error message.
	ErrorLevelError
	// ErrorLevelCritical represents a critical error message.
	ErrorLevelCritical
)

const DefaultErrorLevel = ErrorLevelInfo

// Error wraps a standard error with additional context like error level and input item.
type Error[I any] struct {
	Node      uint64
	Level     ErrorLevel
	Message   string
	Err       error
	InputItem *I
}

// Error implements the error interface.
func (e *Error[I]) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v (input: %+v)", e.Level.String(), e.Message, e.Err, e.InputItem)
	}
	return fmt.Sprintf("[%s] %s (input: %+v)", e.Level.String(), e.Message, e.InputItem)
}

// Unwrap returns the underlying error.
func (e *Error[I]) Unwrap() error {
	return e.Err
}

// String returns the string representation of the error level.
func (level ErrorLevel) String() string {
	switch level {
	case ErrorLevelDebug:
		return "DEBUG"
	case ErrorLevelInfo:
		return "INFO"
	case ErrorLevelWarning:
		return "WARNING"
	case ErrorLevelError:
		return "ERROR"
	case ErrorLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// NewError creates a new Error.
func NewError[I any](node uint64, level ErrorLevel, message string, err error, inputItem *I) *Error[I] {
	if level == ErrorLevelDefault {
		level = DefaultErrorLevel
	}

	return &Error[I]{
		Node:      node,
		Level:     level,
		Message:   message,
		Err:       err,
		InputItem: inputItem,
	}
}
