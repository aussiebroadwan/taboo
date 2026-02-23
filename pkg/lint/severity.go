package lint

// Severity represents the severity level of a validation issue.
type Severity int

const (
	Error Severity = iota
	Warning
	Info
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "ERROR"
	case Warning:
		return "WARN"
	case Info:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}
