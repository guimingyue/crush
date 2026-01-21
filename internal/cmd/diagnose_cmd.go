package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/diagnose"
	"github.com/spf13/cobra"
)

var (
	pid         int
	all         bool
	oomOnly     bool
	cpuOnly     bool
	typeFilter  string
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose application issues including OOM, high CPU, and other online issues",
	Long: `Diagnose application issues including OutOfMemoryError, high CPU usage, 
and other online issues. This command can scan all processes or focus on a specific one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := diagnose.NewManager()
		
		// Create diagnosers for different application types
		javaDiag := diagnose.New(10 * time.Second)
		manager.Register(javaDiag)

		if pid > 0 {
			// Diagnose specific process
			result, err := manager.DiagnoseProcess(pid)
			if err != nil {
				return fmt.Errorf("failed to diagnose process %d: %w", pid, err)
			}

			printDiagnosticResult(result, oomOnly, cpuOnly)
		} else if all {
			// Diagnose all processes
			processes, err := manager.GetAllProcesses()
			if err != nil {
				return fmt.Errorf("failed to find processes: %w", err)
			}

			if len(processes) == 0 {
				fmt.Println("No diagnosable processes found.")
				return nil
			}

			for _, process := range processes {
				// Apply type filter if specified
				if typeFilter != "" && !strings.Contains(strings.ToLower(process.Type), strings.ToLower(typeFilter)) {
					continue
				}

				result, err := manager.DiagnoseProcess(process.PID)
				if err != nil {
					fmt.Printf("Failed to diagnose process %d: %v\n", process.PID, err)
					continue
				}

				printDiagnosticResult(result, oomOnly, cpuOnly)
				fmt.Println(strings.Repeat("-", 80))
			}
		} else {
			// Just list processes
			processes, err := manager.GetAllProcesses()
			if err != nil {
				return fmt.Errorf("failed to find processes: %w", err)
			}

			if len(processes) == 0 {
				fmt.Println("No diagnosable processes found.")
				return nil
			}

			fmt.Printf("%-8s %-60s %-10s %-12s %-12s %-8s\n", "PID", "NAME", "CPU%", "MEM USED", "MEM MAX", "TYPE")
			fmt.Println(strings.Repeat("-", 108))
			for _, process := range processes {
				// Apply type filter if specified
				if typeFilter != "" && !strings.Contains(strings.ToLower(process.Type), strings.ToLower(typeFilter)) {
					continue
				}

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
	diagnoseCmd.Flags().IntVarP(&pid, "pid", "p", 0, "Specific PID to diagnose (optional)")
	diagnoseCmd.Flags().BoolVarP(&all, "all", "a", false, "Diagnose all processes")
	diagnoseCmd.Flags().BoolVar(&oomOnly, "oom-only", false, "Show only OOM-related diagnostics")
	diagnoseCmd.Flags().BoolVar(&cpuOnly, "cpu-only", false, "Show only CPU-related diagnostics")
	diagnoseCmd.Flags().StringVar(&typeFilter, "type", "", "Filter by application type (java, go, etc.)")

	rootCmd.AddCommand(diagnoseCmd)
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

	if result.ThreadDump != "" && !oomOnly && !cpuOnly {
		fmt.Printf("\nThread Dump (first 1000 chars):\n%s\n", truncateString(result.ThreadDump, 1000))
	}

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
	return s[:maxLength] + "..."
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