package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/diagnose"
)

// AIDiagnosticTool provides AI-driven diagnostic capabilities
type AIDiagnosticTool struct {
	appInstance *App
	diagManager *diagnose.Manager
}

// NewAIDiagnosticTool creates a new AI diagnostic tool
func NewAIDiagnosticTool(appInstance *App) *AIDiagnosticTool {
	return &AIDiagnosticTool{
		appInstance: appInstance,
		diagManager: diagnose.NewManager(),
	}
}

// DiagnoseIssue takes an issue description and performs AI-guided diagnostics
func (aidt *AIDiagnosticTool) DiagnoseIssue(ctx context.Context, issueDescription string) error {
	// First, register available diagnosers
	javaDiag := diagnose.New(10 * time.Second)
	aidt.diagManager.Register(javaDiag)

	// Prepare diagnostic context
	systemPrompt := fmt.Sprintf(`You are an expert system administrator and application diagnostician. 
The user has reported the following issue: "%s"

Based on this issue description, determine what diagnostic steps should be taken.
Look for processes that might be related to the issue, run appropriate diagnostic tools,
and provide recommendations to resolve the problem.

Available diagnostic tools:
- Find processes that might be related to the issue
- Run language-specific diagnostics (Java, Go, etc.)
- Analyze memory usage, CPU usage, thread states, garbage collection
- Generate recommendations based on findings

Return your response in a format that can be executed as diagnostic steps.`, issueDescription)

	// Call the LLM to get diagnostic plan
	response, err := aidt.appInstance.AskLLM(ctx, systemPrompt)
	if err != nil {
		return fmt.Errorf("failed to get diagnostic plan from AI: %w", err)
	}

	// Execute the diagnostic plan
	return aidt.executeDiagnosticPlan(ctx, response, issueDescription)
}

// executeDiagnosticPlan executes the diagnostic steps suggested by the AI
func (aidt *AIDiagnosticTool) executeDiagnosticPlan(ctx context.Context, aiPlan, issueDescription string) error {
	// For now, we'll implement a basic version that looks for keywords in the AI response
	// and executes appropriate diagnostic steps
	
	planLower := strings.ToLower(aiPlan)
	
	// Check if AI suggested looking for Java processes
	if strings.Contains(planLower, "java") || strings.Contains(planLower, "jvm") {
		fmt.Println("🔍 Looking for Java processes...")
		
		// Find all processes
		processes, err := aidt.diagManager.GetAllProcesses()
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not get all processes: %v\n", err)
		} else {
			if len(processes) == 0 {
				fmt.Println("No processes found.")
			} else {
				fmt.Printf("Found %d processes:\n", len(processes))
				for _, proc := range processes {
					fmt.Printf("  PID: %d, Name: %s, Type: %s, CPU: %.2f%%, Memory: %s/%s\n",
						proc.PID, proc.Name, proc.Type, proc.CPUUsage,
						formatBytes(proc.MemoryUsed), formatBytes(proc.MemoryMax))
				}
				
				// Run diagnostics on each process
				for _, proc := range processes {
					fmt.Printf("\n📋 Diagnosing process %d (%s)...\n", proc.PID, proc.Name)
					result, err := aidt.diagManager.DiagnoseProcess(proc.PID)
					if err != nil {
						fmt.Printf("❌ Failed to diagnose process %d: %v\n", proc.PID, err)
						continue
					}
					
					// Print diagnostic results
					fmt.Printf("✅ Process %d analysis complete:\n", proc.PID)
					if result.IsOOM {
						fmt.Printf("  ⚠️  Potential OOM issue detected\n")
					}
					if result.IsHighCPU {
						fmt.Printf("  ⚠️  High CPU usage detected (%.2f%%)\n", proc.CPUUsage)
					}
					if result.IsStuckThread {
						fmt.Printf("  ⚠️  Stuck threads detected\n")
					}
					if result.HasFrequentGC {
						fmt.Printf("  ⚠️  Frequent garbage collection detected\n")
					}
					if result.HasFullGCIssues {
						fmt.Printf("  ⚠️  Full GC issues detected\n")
					}
					
					// Print specific issues found
					if len(result.Issues) > 0 {
						fmt.Printf("  📋 Issues found:\n")
						for _, issue := range result.Issues {
							fmt.Printf("    • [%s] %s: %s\n", issue.Severity, issue.Type, issue.Description)
							fmt.Printf("      💡 Suggestion: %s\n", issue.Suggestion)
						}
					}
				}
			}
		}
	} else {
		// If no specific language mentioned, just list processes
		fmt.Println("🔍 Listing all processes...")
		processes, err := aidt.diagManager.GetAllProcesses()
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not get processes: %v\n", err)
		} else {
			if len(processes) == 0 {
				fmt.Println("No processes found.")
			} else {
				fmt.Printf("Found %d processes:\n", len(processes))
				for _, proc := range processes {
					fmt.Printf("  PID: %d, Name: %s, Type: %s, CPU: %.2f%%, Memory: %s/%s\n",
						proc.PID, proc.Name, proc.Type, proc.CPUUsage,
						formatBytes(proc.MemoryUsed), formatBytes(proc.MemoryMax))
				}
			}
		}
	}
	
	// Provide AI-generated recommendations based on the original issue
	summaryPrompt := fmt.Sprintf(`Based on the issue description "%s" and the system diagnostics performed, 
provide specific recommendations to resolve the issue. Include command-line tools that could be used 
to further investigate or resolve the problem.`, issueDescription)
	
	recommendations, err := aidt.appInstance.AskLLM(ctx, summaryPrompt)
	if err != nil {
		fmt.Printf("⚠️  Could not get AI recommendations: %v\n", err)
	} else {
		fmt.Printf("\n💡 AI Recommendations:\n%s\n", recommendations)
	}
	
	return nil
}

// formatBytes formats bytes into human-readable format
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