package diagnose

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// JavaProcess represents a Java process with its diagnostic information
type JavaProcess struct {
	Process
}


// Diagnostics is the main struct for performing Java diagnostics
type Diagnostics struct {
	timeout time.Duration
}

var _ Diagnoser = (*Diagnostics)(nil) // Ensure Diagnostics implements Diagnoser interface

// New creates a new Diagnostics instance
func New(timeout time.Duration) *Diagnostics {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Diagnostics{
		timeout: timeout,
	}
}

// FindJavaProcesses finds all running Java processes
func (d *Diagnostics) FindJavaProcesses() ([]JavaProcess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	// Use ps command to find Java processes, compatible with both Linux and macOS
	cmd := exec.CommandContext(ctx, "ps", "axo", "pid,pcpu,pmem,vsz,rss,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var processes []JavaProcess

	// Skip header line
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		pidStr := fields[0]
		command := fields[5] // comm field (command name)

		// Check if this is a Java process
		if strings.Contains(command, "java") || strings.Contains(strings.ToLower(command), "java") {
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			// Double-check that it's really a Java process by getting its full command line
			name, err := d.getJavaProcessName(pid)
			if err != nil || !strings.Contains(strings.ToLower(name), "java") {
				continue
			}

			cpu, memUsed, memMax, err := d.getProcessStats(pid)
			if err != nil {
				continue
			}

			processes = append(processes, JavaProcess{
				Process: Process{
					PID:        pid,
					Name:       name,
					CPUUsage:   cpu,
					MemoryUsed: memUsed,
					MemoryMax:  memMax,
					Type:       "java",
				},
			})
		}
	}

	return processes, nil
}

// getJavaProcessName gets the name/command line of a Java process
func (d *Diagnostics) getJavaProcessName(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	// Use different ps options for macOS compatibility (no --no-headers on macOS)
	cmd := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "args")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get process name: %w", err)
	}

	// Split output and skip the header line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("insufficient output from ps command")
	}

	name := strings.TrimSpace(lines[1]) // Skip header line
	if len(name) > 100 {
		name = name[:100] + "..."
	}
	return name, nil
}

// getProcessStats gets CPU and memory stats for a process
func (d *Diagnostics) getProcessStats(pid int) (cpuPercent float64, memUsed, memMax uint64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "%cpu,vsz,rss")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get process stats: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return 0, 0, 0, fmt.Errorf("unexpected ps output format: insufficient lines")
	}

	// Skip header line and process data
	dataLine := lines[1]
	fields := strings.Fields(dataLine)
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("unexpected ps output format: not enough fields")
	}

	cpuPercent, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse CPU percentage: %w", err)
	}

	vsz, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse virtual memory size: %w", err)
	}

	rss, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse resident memory size: %w", err)
	}

	// Convert KB to bytes
	memUsed = rss * 1024
	memMax = vsz * 1024

	return cpuPercent, memUsed, memMax, nil
}

// DiagnoseProcess performs a full diagnostic on a specific Java process
func (d *Diagnostics) DiagnoseProcess(pid int) (*DiagnosticResult, error) {
	// First, verify that the process exists and is a Java process
	name, err := d.getJavaProcessName(pid)
	if err != nil {
		return nil, fmt.Errorf("process with PID %d not found or not accessible: %w", pid, err)
	}

	if !strings.Contains(strings.ToLower(name), "java") {
		return nil, fmt.Errorf("process with PID %d is not a Java process", pid)
	}

	// Get process stats
	cpu, memUsed, memMax, err := d.getProcessStats(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get process stats for PID %d: %w", pid, err)
	}

	baseProcess := Process{
		PID:        pid,
		Name:       name,
		CPUUsage:   cpu,
		MemoryUsed: memUsed,
		MemoryMax:  memMax,
		Type:       "java",
	}

	result := &DiagnosticResult{
		Process:   baseProcess,
		Timestamp: time.Now(),
		Issues:    []Issue{},
	}

	// Check for OOM conditions
	result.IsOOM = d.checkForOOM(baseProcess)

	// Check for high CPU usage (threshold: 80%)
	result.IsHighCPU = baseProcess.CPUUsage > 80.0

	// Get thread dump
	threadDump, err := d.getThreadDump(pid)
	if err != nil {
		// Log the error but don't fail the entire diagnostic
		fmt.Printf("Warning: failed to get thread dump: %v\n", err)
	} else {
		result.ThreadDump = threadDump
		result.IsStuckThread = d.analyzeThreadDump(threadDump)
	}

	// Check GC activity
	gcActivity, err := d.getGCActivity(pid)
	if err != nil {
		fmt.Printf("Warning: failed to get GC activity: %v\n", err)
	} else {
		result.GCActivity = gcActivity
		freqGC, fullGCIssues := d.analyzeGCActivity(gcActivity)
		result.HasFrequentGC = freqGC
		result.HasFullGCIssues = fullGCIssues
	}

	// Add issues based on findings
	if result.IsOOM {
		result.Issues = append(result.Issues, Issue{
			Type:        "OOM",
			Description: "Process is approaching OutOfMemoryError conditions",
			Severity:    "high",
			Suggestion:  "Consider increasing heap size or investigating memory leaks",
		})
	}

	if result.IsHighCPU {
		result.Issues = append(result.Issues, Issue{
			Type:        "HighCPU",
			Description: "Process is consuming excessive CPU resources",
			Severity:    "medium",
			Suggestion:  "Consider profiling the application to identify hotspots",
		})
	}

	if result.IsStuckThread {
		result.Issues = append(result.Issues, Issue{
			Type:        "StuckThread",
			Description: "Process has stuck or blocked threads",
			Severity:    "high",
			Suggestion:  "Check thread dump for deadlock or infinite loop conditions",
		})
	}

	if result.HasFrequentGC {
		result.Issues = append(result.Issues, Issue{
			Type:        "FrequentGC",
			Description: "Process is experiencing frequent garbage collection",
			Severity:    "medium",
			Suggestion:  "Consider tuning GC parameters or investigating memory usage patterns",
		})
	}

	if result.HasFullGCIssues {
		result.Issues = append(result.Issues, Issue{
			Type:        "FullGCIssues",
			Description: "Process is experiencing problematic full garbage collection cycles",
			Severity:    "high",
			Suggestion:  "Consider increasing heap size or tuning GC algorithm",
		})
	}

	return result, nil
}

// checkForOOM determines if a process is experiencing OOM conditions
func (d *Diagnostics) checkForOOM(process Process) bool {
	// Heuristic: if memory usage is > 90% of max memory, flag as potential OOM
	if process.MemoryMax > 0 {
		usagePercent := float64(process.MemoryUsed) / float64(process.MemoryMax) * 100
		if usagePercent > 90.0 {
			return true
		}
	}

	// Alternative heuristic: if used memory is very high (> 8GB)
	if process.MemoryUsed > 8*1024*1024*1024 { // 8GB
		return true
	}

	// Check if the process is close to its memory limits based on common JVM defaults
	// If memory used is > 95% of a common default heap size (e.g. 4GB), flag as potential OOM
	if process.MemoryUsed > 0 && process.MemoryMax == 0 {
		// If we can't determine max memory, check against common thresholds
		return process.MemoryUsed > 4*1024*1024*1024 // 4GB
	}

	return false
}

// IdentifyProcess determines if a process is a Java process
func (d *Diagnostics) IdentifyProcess(pid int) (bool, error) {
	name, err := d.getJavaProcessName(pid)
	if err != nil {
		return false, err
	}

	return strings.Contains(strings.ToLower(name), "java"), nil
}

// GetProcessInfo retrieves basic information about a Java process
func (d *Diagnostics) GetProcessInfo(pid int) (*Process, error) {
	name, err := d.getJavaProcessName(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get process name: %w", err)
	}

	if !strings.Contains(strings.ToLower(name), "java") {
		return nil, fmt.Errorf("process with PID %d is not a Java process", pid)
	}

	cpu, memUsed, memMax, err := d.getProcessStats(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get process stats: %w", err)
	}

	process := &Process{
		PID:        pid,
		Name:       name,
		CPUUsage:   cpu,
		MemoryUsed: memUsed,
		MemoryMax:  memMax,
		Type:       "java",
	}

	return process, nil
}

// Diagnose performs a full diagnostic on a Java process
func (d *Diagnostics) Diagnose(pid int) (*DiagnosticResult, error) {
	result, err := d.DiagnoseProcess(pid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetName returns the name of the diagnoser
func (d *Diagnostics) GetName() string {
	return "java"
}

// getThreadDump generates a thread dump for the specified Java process
func (d *Diagnostics) getThreadDump(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "jstack", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get thread dump: %w", err)
	}

	return string(output), nil
}

// analyzeThreadDump looks for stuck threads in the thread dump
func (d *Diagnostics) analyzeThreadDump(threadDump string) bool {
	// Look for common signs of stuck threads
	lines := strings.Split(threadDump, "\n")

	// Count BLOCKED and WAITING threads
	blockedCount := 0
	waitingCount := 0
	deadlockCount := 0

	for i, line := range lines {
		if strings.Contains(line, "BLOCKED") {
			blockedCount++
		} else if strings.Contains(line, "WAITING") {
			waitingCount++
		} else if strings.Contains(line, "deadlock") || strings.Contains(line, "Deadlock") {
			deadlockCount++
		}

		// Check for threads in the same wait state for extended periods
		// This is a simplified check - in practice, you'd need historical data
		if strings.Contains(line, "waiting for monitor entry") {
			// Check if the next few lines show the same lock being waited on
			for j := i + 1; j < len(lines) && j < i + 5; j++ {
				if strings.Contains(lines[j], "locked") && strings.Contains(lines[j], "<") {
					// Found a potential lock contention
					return true
				}
			}
		}
	}

	// Heuristic: if there are many blocked threads, deadlocks, or excessive waiting, flag as stuck
	return blockedCount > 10 || waitingCount > 50 || deadlockCount > 0
}

// getGCActivity gets garbage collection activity for the process
func (d *Diagnostics) getGCActivity(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "jstat", "-gc", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GC statistics: %w", err)
	}

	return string(output), nil
}

// analyzeGCActivity analyzes GC statistics for potential issues
func (d *Diagnostics) analyzeGCActivity(gcStats string) (frequentGC bool, fullGCIssues bool) {
	lines := strings.Split(strings.TrimSpace(gcStats), "\n")
	if len(lines) < 2 {
		return false, false
	}

	// Skip header line and process data
	dataLine := lines[1]
	fields := strings.Fields(dataLine)

	if len(fields) < 8 {
		// Expected fields: S0C, S1C, S0U, S1U, EC, EU, OC, OU, MC, MU, CCSC, CCSU, YGC, YGCT, FGC, FGCT, GCT
		return false, false
	}

	// Parse key GC metrics
	youngGCCount, _ := parseFloatOrZero(fields[12]) // YGC - Young GC count
	_, _ = parseFloatOrZero(fields[13]) // YGCT - Young GC time
	fullGCCount, _ := parseFloatOrZero(fields[14]) // FGC - Full GC count
	_, _ = parseFloatOrZero(fields[15]) // FGCT - Full GC time

	// Check for frequent GC activity (potential memory pressure)
	// If young GCs are happening very frequently, it might indicate memory pressure
	frequentGC = youngGCCount > 100 // More than 100 young GCs might indicate issues

	// Check for full GC issues (long pauses, too frequent)
	fullGCIssues = fullGCCount > 5 // More than 5 full GCs might indicate problems

	return frequentGC, fullGCIssues
}

// parseFloatOrZero parses a float from string, returning 0 if parsing fails
func parseFloatOrZero(s string) (float64, error) {
	if s == "-" || s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

// GenerateHeapDump generates a heap dump for the specified Java process
func (d *Diagnostics) GenerateHeapDump(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Longer timeout for heap dump
	defer cancel()

	heapDumpPath := fmt.Sprintf("/tmp/java_heap_%d.hprof", pid)
	cmd := exec.CommandContext(ctx, "jmap", "-dump:format=b,file="+heapDumpPath, strconv.Itoa(pid))
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to generate heap dump: %w", err)
	}

	return heapDumpPath, nil
}

// GetJVMInfo gets basic JVM information for the process
func (d *Diagnostics) GetJVMInfo(pid int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "jinfo", "-flags", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get JVM info: %w", err)
	}

	return string(output), nil
}