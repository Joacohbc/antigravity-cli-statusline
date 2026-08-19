package payload

import (
	"errors"
	"fmt"
)

var (
	// ErrEmptyPayload indicates that the provided input stream or byte slice was empty or contained only whitespace.
	ErrEmptyPayload = errors.New("payload is empty")

	// ErrPayloadOversized indicates that the input payload exceeded the maximum permitted size (2MB).
	ErrPayloadOversized = errors.New("payload exceeds maximum allowed size (2MB)")

	// ErrInvalidJSON indicates that the payload could not be parsed as valid JSON.
	ErrInvalidJSON = errors.New("invalid JSON payload")
)

// ParseError represents a structured error that occurs during payload parsing.
type ParseError struct {
	Op  string // Operation that failed (e.g., "read", "limit_check", "parse", "unmarshal")
	Err error  // Underlying sentinel or root cause error
	Msg string // Contextual description
}

// Error formats the ParseError into a readable string.
func (e *ParseError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Msg != "" && e.Err != nil {
		return fmt.Sprintf("payload parse error during %s: %s: %v", e.Op, e.Msg, e.Err)
	}
	if e.Err != nil {
		return fmt.Sprintf("payload parse error during %s: %v", e.Op, e.Err)
	}
	if e.Msg != "" {
		return fmt.Sprintf("payload parse error during %s: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("payload parse error during %s", e.Op)
}

// Unwrap returns the underlying error.
func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is reports whether the ParseError matches target error for errors.Is inspection.
func (e *ParseError) Is(target error) bool {
	if target == nil {
		return false
	}
	if target == e {
		return true
	}
	return errors.Is(e.Err, target)
}
