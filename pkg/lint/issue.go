package lint

import (
	"fmt"
	"strings"
)

// Issue represents a single validation issue.
type Issue struct {
	Severity Severity
	Rule     string // e.g., "env-invalid", "port-privileged"
	Message  string
	Location string // e.g., "server.port", "database.dsn"
}

func (i Issue) String() string {
	return fmt.Sprintf("[%s] %s: %s (location: %s)", i.Severity, i.Rule, i.Message, i.Location)
}

// Issues is a collection of validation issues with helper methods.
type Issues []Issue

// HasErrors returns true if any issues are errors.
func (issues Issues) HasErrors() bool {
	for _, issue := range issues {
		if issue.Severity == Error {
			return true
		}
	}
	return false
}

// Errors returns only the error-level issues.
func (issues Issues) Errors() Issues {
	return issues.Filter(Error)
}

// Warnings returns only the warning-level issues.
func (issues Issues) Warnings() Issues {
	return issues.Filter(Warning)
}

// Infos returns only the info-level issues.
func (issues Issues) Infos() Issues {
	return issues.Filter(Info)
}

// Filter returns issues matching the given severity.
func (issues Issues) Filter(severity Severity) Issues {
	var filtered Issues
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// Count returns counts of errors, warnings, and infos.
func (issues Issues) Count() (errors, warnings, infos int) {
	for _, issue := range issues {
		switch issue.Severity {
		case Error:
			errors++
		case Warning:
			warnings++
		case Info:
			infos++
		}
	}
	return
}

// Error implements the error interface, returning error messages only.
func (issues Issues) Error() string {
	var msgs []string
	for _, issue := range issues {
		if issue.Severity == Error {
			msgs = append(msgs, issue.Message)
		}
	}
	if len(msgs) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

// Err returns the issues as an error if there are any errors, nil otherwise.
func (issues Issues) Err() error {
	if issues.HasErrors() {
		return issues
	}
	return nil
}
