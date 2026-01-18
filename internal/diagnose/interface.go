package diagnose

import "time"

// Process represents a running process with diagnostic information
type Process struct {
	PID        int
	Name       string
	CPUUsage   float64
	MemoryUsed uint64
	MemoryMax  uint64
	Type       string // Language/runtime type: java, go, cpp, etc.
}

// DiagnosticResult contains the results of a diagnostic check
type DiagnosticResult struct {
	Process         Process
	IsOOM           bool
	IsHighCPU       bool
	IsStuckThread   bool
	HasFrequentGC   bool
	HasFullGCIssues bool
	ThreadDump      string
	HeapDumpPath    string
	GCActivity      string
	Timestamp       time.Time
	Issues          []Issue
}

// Issue represents a specific issue found during diagnostics
type Issue struct {
	Type        string
	Description string
	Severity    string // low, medium, high, critical
	Suggestion  string
}

// Diagnoser is the interface for language-specific diagnostic implementations
type Diagnoser interface {
	// IdentifyProcess determines if a process is of the language type this diagnoser handles
	IdentifyProcess(pid int) (bool, error)
	
	// GetProcessInfo retrieves basic information about a process
	GetProcessInfo(pid int) (*Process, error)
	
	// Diagnose performs a full diagnostic on a process
	Diagnose(pid int) (*DiagnosticResult, error)
	
	// GetName returns the name of the diagnoser (e.g., "java", "go", "cpp")
	GetName() string
}