// Package lint provides structured validation with severity levels.
//
// It supports three severity levels: Error, Warning, and Info.
// Use the Collector to build up validation results with a fluent API:
//
//	c := lint.NewCollector()
//	c.Error("rule-name", "location", "error message")
//	c.Warn("rule-name", "location", "warning message")
//	c.Infof("rule-name", "location", "info with %s", "formatting")
//
//	issues := c.Issues()
//	if issues.HasErrors() {
//	    return issues.Err()
//	}
package lint
