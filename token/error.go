package token

import (
	"fmt"
	"strings"
)

type ErrorKind int

const (
	ErrorKindParse ErrorKind = iota
	ErrorKindCompile
	ErrorKindRuntime
)

func (k ErrorKind) String() string {
	switch k {
	case ErrorKindParse:
		return "Parse error"
	case ErrorKindCompile:
		return "Compile error"
	case ErrorKindRuntime:
		return "Runtime error"
	default:
		return "Error"
	}
}

type ScriptError struct {
	Kind       ErrorKind
	Message    string
	Line       int
	Column     int
	File       string
	Function   string
	StackTrace []string
}

func (e *ScriptError) Error() string {
	var sb strings.Builder

	// Header: "Runtime error at script.ice:87 in Wants"
	sb.WriteString(e.Kind.String())
	if e.File != "" || e.Line > 0 {
		sb.WriteString(" at ")
		if e.File != "" {
			sb.WriteString(e.File)
		} else {
			sb.WriteString("script")
		}
		if e.Line > 0 {
			sb.WriteString(fmt.Sprintf(":%d", e.Line))
		}
	}
	if e.Function != "" {
		sb.WriteString(" in ")
		sb.WriteString(e.Function)
	}
	sb.WriteString("\n")

	// Message indented
	sb.WriteString("  ")
	sb.WriteString(e.Message)

	// Stack trace if present (for panic parity, though panic puts it after)
	if len(e.StackTrace) > 0 {
		sb.WriteString("\nStack trace:\n")
		for _, frame := range e.StackTrace {
			sb.WriteString("  ")
			sb.WriteString(frame)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
