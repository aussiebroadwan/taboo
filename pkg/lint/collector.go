package lint

import "fmt"

// Collector accumulates validation issues with a fluent API.
type Collector struct {
	issues Issues
}

// NewCollector creates a new validation collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Error adds an error-level issue.
func (c *Collector) Error(rule, location, message string) *Collector {
	c.issues = append(c.issues, Issue{
		Severity: Error,
		Rule:     rule,
		Location: location,
		Message:  message,
	})
	return c
}

// Errorf adds an error-level issue with formatted message.
func (c *Collector) Errorf(rule, location, format string, args ...any) *Collector {
	return c.Error(rule, location, fmt.Sprintf(format, args...))
}

// Warn adds a warning-level issue.
func (c *Collector) Warn(rule, location, message string) *Collector {
	c.issues = append(c.issues, Issue{
		Severity: Warning,
		Rule:     rule,
		Location: location,
		Message:  message,
	})
	return c
}

// Warnf adds a warning-level issue with formatted message.
func (c *Collector) Warnf(rule, location, format string, args ...any) *Collector {
	return c.Warn(rule, location, fmt.Sprintf(format, args...))
}

// Info adds an info-level issue.
func (c *Collector) Info(rule, location, message string) *Collector {
	c.issues = append(c.issues, Issue{
		Severity: Info,
		Rule:     rule,
		Location: location,
		Message:  message,
	})
	return c
}

// Infof adds an info-level issue with formatted message.
func (c *Collector) Infof(rule, location, format string, args ...any) *Collector {
	return c.Info(rule, location, fmt.Sprintf(format, args...))
}

// Add appends an existing issue.
func (c *Collector) Add(issue Issue) *Collector {
	c.issues = append(c.issues, issue)
	return c
}

// Merge appends all issues from another collection.
func (c *Collector) Merge(issues Issues) *Collector {
	c.issues = append(c.issues, issues...)
	return c
}

// Issues returns the collected issues.
func (c *Collector) Issues() Issues {
	return c.issues
}

// HasErrors returns true if any collected issues are errors.
func (c *Collector) HasErrors() bool {
	return c.issues.HasErrors()
}

// Err returns the issues as an error if there are any errors, nil otherwise.
func (c *Collector) Err() error {
	return c.issues.Err()
}
