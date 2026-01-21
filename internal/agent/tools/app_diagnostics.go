package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/diagnose"
)

type AppDiagnosticsParams struct {
	ProcessID int    `json:"process_id,omitempty" description:"The process ID to diagnose (leave empty for all processes)"`
	Type      string `json:"type,omitempty" description:"Type of application to diagnose (java, go, etc.)"`
}

const AppDiagnosticsToolName = "app_diagnostics"

//go:embed app_diagnostics.md
var appDiagnosticsDescription []byte

func NewAppDiagnosticsTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		AppDiagnosticsToolName,
		string(appDiagnosticsDescription),
		func(ctx context.Context, params AppDiagnosticsParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return runAppDiagnostics(ctx, params)
		})
}

func runAppDiagnostics(ctx context.Context, params AppDiagnosticsParams) (fantasy.ToolResponse, error) {
	// Create a diagnostic manager
	manager := diagnose.NewManager()

	// Create Java diagnoser with default configuration
	javaDiag := diagnose.New(10*time.Second)
	manager.Register(javaDiag)

	var result strings.Builder

	if params.ProcessID > 0 {
		// Diagnose specific process
		diagResult, err := manager.DiagnoseProcess(params.ProcessID)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to diagnose process %d: %v", params.ProcessID, err)), nil
		}

		result.WriteString(formatDiagnosticResult(diagResult))

		// Add additional system-level diagnostics
		result.WriteString("\nAdditional System Information:\n")

		// Get system process info
		sysInfo, err := javaDiag.GetSystemProcessInfo(params.ProcessID)
		if err == nil && sysInfo != "" {
			result.WriteString(fmt.Sprintf("Process Details:\n%s\n", sysInfo))
		}

		// Get open files (if available)
		openFiles, err := javaDiag.GetOpenFiles(params.ProcessID)
		if err == nil && openFiles != "" && !strings.Contains(openFiles, "lsof not available") {
			result.WriteString(fmt.Sprintf("Open Files:\n%s\n", openFiles))
		}

		// Get network connections (if available)
		netConns, err := javaDiag.GetNetworkConnections(params.ProcessID)
		if err == nil && netConns != "" && !strings.Contains(netConns, "lsof not available") {
			result.WriteString(fmt.Sprintf("Network Connections:\n%s\n", netConns))
		}
	} else {
		// Diagnose all processes
		processes, err := manager.GetAllProcesses()
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get processes: %v", err)), nil
		}

		if len(processes) == 0 {
			result.WriteString("No diagnosable processes found.\n")
		} else {
			result.WriteString(fmt.Sprintf("Found %d processes:\n", len(processes)))
			for _, process := range processes {
				result.WriteString(fmt.Sprintf("- PID: %d, Name: %s, Type: %s, CPU: %.2f%%, Memory: %s/%s\n",
					process.PID, process.Name, process.Type, process.CPUUsage,
					formatBytes(process.MemoryUsed), formatBytes(process.MemoryMax)))

				// Also run diagnostics on each process
				diagResult, err := manager.DiagnoseProcess(process.PID)
				if err != nil {
					result.WriteString(fmt.Sprintf("  Error diagnosing: %v\n", err))
					continue
				}

				result.WriteString("  Diagnostics:\n")
				result.WriteString(indentText(formatDiagnosticResult(diagResult), "    "))

				// Add additional system-level diagnostics for each process
				sysInfo, err := javaDiag.GetSystemProcessInfo(process.PID)
				if err == nil && sysInfo != "" {
					result.WriteString(indentText(fmt.Sprintf("  Process Details:\n%s", sysInfo), "    "))
				}
			}
		}
	}

	return fantasy.NewTextResponse(result.String()), nil
}

func formatDiagnosticResult(result *diagnose.DiagnosticResult) string {
	var output strings.Builder

	fmt.Fprintf(&output, "Process: %s (PID: %d)\n", result.Process.Name, result.Process.PID)
	fmt.Fprintf(&output, "Type: %s\n", result.Process.Type)
	fmt.Fprintf(&output, "CPU Usage: %.2f%%\n", result.Process.CPUUsage)
	fmt.Fprintf(&output, "Memory Used: %s\n", formatBytes(result.Process.MemoryUsed))
	fmt.Fprintf(&output, "Memory Max: %s\n", formatBytes(result.Process.MemoryMax))
	fmt.Fprintf(&output, "Potential OOM: %t\n", result.IsOOM)
	fmt.Fprintf(&output, "High CPU Usage: %t\n", result.IsHighCPU)
	fmt.Fprintf(&output, "Stuck Threads: %t\n", result.IsStuckThread)
	fmt.Fprintf(&output, "Frequent GC: %t\n", result.HasFrequentGC)
	fmt.Fprintf(&output, "Full GC Issues: %t\n", result.HasFullGCIssues)

	if result.GCActivity != "" {
		output.WriteString("GC Activity:\n")
		output.WriteString(result.GCActivity)
		output.WriteString("\n")
	}

	if len(result.Issues) > 0 {
		output.WriteString("Issues Found:\n")
		for _, issue := range result.Issues {
			fmt.Fprintf(&output, "  - [%s] %s: %s\n    Suggestion: %s\n",
				issue.Severity, issue.Type, issue.Description, issue.Suggestion)
		}
	}

	return output.String()
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func indentText(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}