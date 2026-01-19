package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/crush/internal/diagnose"
	"github.com/spf13/cobra"
)

var javaDiagCmd = &cobra.Command{
	Use:   "java-diag",
	Short: "Diagnose Java application issues (OOM, high CPU, etc.)",
	Long: `Diagnose Java application issues including OutOfMemoryError, high CPU usage, 
and other online issues. This command can scan all Java processes or focus on a specific one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, _ := cmd.Flags().GetInt("pid")
		all, _ := cmd.Flags().GetBool("all")
		oomOnly, _ := cmd.Flags().GetBool("oom-only")
		cpuOnly, _ := cmd.Flags().GetBool("cpu-only")

		manager := diagnose.NewManager()
		javaDiag := diagnose.New(0) // Use default timeout
		manager.Register(javaDiag)

		if pid > 0 {
			// Diagnose specific process
			result, err := manager.DiagnoseProcess(pid)
			if err != nil {
				return fmt.Errorf("failed to diagnose process %d: %w", pid, err)
			}

			// Convert to the expected format for printing
			wrappedResult := &diagnose.DiagnosticResult{
				Process:         result.Process,
				IsOOM:           false, // Will be determined from issues
				IsHighCPU:       false, // Will be determined from issues
				IsStuckThread:   false, // Will be determined from issues
				HasFrequentGC:   false, // Will be determined from issues
				HasFullGCIssues: false, // Will be determined from issues
				ThreadDump:      "",    // May be populated in issues
				HeapDumpPath:    "",
				GCActivity:      "",
				Timestamp:       result.Timestamp,
				Issues:          result.Issues,
			}

			// Set boolean flags based on issues found
			for _, issue := range result.Issues {
				switch issue.Type {
				case "OOM":
					wrappedResult.IsOOM = true
				case "HighCPU":
					wrappedResult.IsHighCPU = true
				case "StuckThread":
					wrappedResult.IsStuckThread = true
				case "FrequentGC":
					wrappedResult.HasFrequentGC = true
				case "FullGCIssues":
					wrappedResult.HasFullGCIssues = true
				}
			}

			printDiagnosticResult(wrappedResult, oomOnly, cpuOnly)
		} else if all {
			// Diagnose all processes
			processes, err := manager.GetAllProcesses()
			if err != nil {
				return fmt.Errorf("failed to find processes: %w", err)
			}

			if len(processes) == 0 {
				fmt.Println("No processes found.")
				return nil
			}

			for _, process := range processes {
				result, err := manager.DiagnoseProcess(process.PID)
				if err != nil {
					slog.Error("Failed to diagnose process", "pid", process.PID, "error", err)
					continue
				}

				// Convert to the expected format for printing
				wrappedResult := &diagnose.DiagnosticResult{
					Process:         result.Process,
					IsOOM:           false, // Will be determined from issues
					IsHighCPU:       false, // Will be determined from issues
					IsStuckThread:   false, // Will be determined from issues
					HasFrequentGC:   false, // Will be determined from issues
					HasFullGCIssues: false, // Will be determined from issues
					ThreadDump:      "",    // May be populated in issues
					HeapDumpPath:    "",
					GCActivity:      "",
					Timestamp:       result.Timestamp,
					Issues:          result.Issues,
				}

				// Set boolean flags based on issues found
				for _, issue := range result.Issues {
					switch issue.Type {
					case "OOM":
						wrappedResult.IsOOM = true
					case "HighCPU":
						wrappedResult.IsHighCPU = true
					case "StuckThread":
						wrappedResult.IsStuckThread = true
					case "FrequentGC":
						wrappedResult.HasFrequentGC = true
					case "FullGCIssues":
						wrappedResult.HasFullGCIssues = true
					}
				}

				printDiagnosticResult(wrappedResult, oomOnly, cpuOnly)
				fmt.Println(strings.Repeat("-", 80))
			}
		} else {
			// Just list processes
			processes, err := manager.GetAllProcesses()
			if err != nil {
				return fmt.Errorf("failed to find processes: %w", err)
			}

			if len(processes) == 0 {
				fmt.Println("No processes found.")
				return nil
			}

			fmt.Printf("%-8s %-60s %-10s %-12s %-12s %-8s\n", "PID", "NAME", "CPU%", "MEM USED", "MEM MAX", "TYPE")
			fmt.Println(strings.Repeat("-", 108))
			for _, process := range processes {
				fmt.Printf("%-8d %-60s %-10.2f %-12s %-12s %-8s\n",
					process.PID,
					truncateString(process.Name, 60),
					process.CPUUsage,
					formatBytes(process.MemoryUsed),
					formatBytes(process.MemoryMax),
					process.Type)
			}
		}

		return nil
	},
}

func init() {
	javaDiagCmd.Flags().IntP("pid", "p", 0, "Specific PID to diagnose (optional)")
	javaDiagCmd.Flags().BoolP("all", "a", false, "Diagnose all Java processes")
	javaDiagCmd.Flags().Bool("oom-only", false, "Show only OOM-related diagnostics")
	javaDiagCmd.Flags().Bool("cpu-only", false, "Show only CPU-related diagnostics")

	rootCmd.AddCommand(javaDiagCmd)
}

func printDiagnosticResult(result *diagnose.DiagnosticResult, oomOnly, cpuOnly bool) {
	fmt.Printf("\n=== DIAGNOSTIC RESULT FOR PID %d ===\n", result.Process.PID)
	fmt.Printf("Name: %s\n", truncateString(result.Process.Name, 80))
	fmt.Printf("Type: %s\n", result.Process.Type)
	fmt.Printf("Timestamp: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("CPU Usage: %.2f%%\n", result.Process.CPUUsage)
	fmt.Printf("Memory Used: %s\n", formatBytes(result.Process.MemoryUsed))
	fmt.Printf("Memory Max: %s\n", formatBytes(result.Process.MemoryMax))

	if !cpuOnly {
		fmt.Printf("Potential OOM: %t\n", result.IsOOM)
	}

	if !oomOnly {
		fmt.Printf("High CPU Usage: %t\n", result.IsHighCPU)
		fmt.Printf("Stuck Threads Detected: %t\n", result.IsStuckThread)
	}

	if !oomOnly && !cpuOnly {
		fmt.Printf("Frequent GC Activity: %t\n", result.HasFrequentGC)
		fmt.Printf("Full GC Issues: %t\n", result.HasFullGCIssues)
	}

	if result.GCActivity != "" && !oomOnly && !cpuOnly {
		fmt.Printf("\nGC Activity:\n%s\n", result.GCActivity)
	}

	// Print issues found
	if len(result.Issues) > 0 && !oomOnly && !cpuOnly {
		fmt.Printf("\nIssues Found:\n")
		for _, issue := range result.Issues {
			fmt.Printf("  [%s] %s: %s\n", issue.Severity, issue.Type, issue.Description)
			fmt.Printf("      Suggestion: %s\n", issue.Suggestion)
		}
	}

	if result.IsOOM && !cpuOnly {
		fmt.Println("\nRecommendation: Consider generating a heap dump for further analysis:")
		fmt.Printf("  jmap -dump:format=b,file=/tmp/java_heap_%d.hprof %d\n", result.Process.PID, result.Process.PID)
	}

	if result.IsHighCPU && !oomOnly {
		fmt.Println("\nRecommendation: Consider getting a thread dump to identify hot methods:")
		fmt.Printf("  jstack %d\n", result.Process.PID)
	}

	if result.HasFrequentGC && !oomOnly {
		fmt.Println("\nRecommendation: High frequency of GC activity detected. Consider tuning GC parameters or investigating memory usage patterns.")
	}

	if result.HasFullGCIssues && !oomOnly {
		fmt.Println("\nRecommendation: Full GC issues detected. This may cause long pause times. Consider increasing heap size or tuning GC algorithm.")
	}
}


func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
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