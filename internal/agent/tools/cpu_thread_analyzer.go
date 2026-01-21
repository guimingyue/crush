package tools

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"charm.land/fantasy"
)

// CPUThreadAnalyzerParams are parameters for the CPU thread analyzer tool
type CPUThreadAnalyzerParams struct {
	PID int `json:"pid" description:"Process ID of the Java application to analyze"`
}

// CPUThreadAnalyzerToolName is the name of the CPU thread analyzer tool
const CPUThreadAnalyzerToolName = "cpu_thread_analyzer"

// ThreadInfo holds information about a specific thread
type ThreadInfo struct {
	ID          int
	Name        string
	NativeID    int
	CPUUsage    float64
	State       string
	StackTrace  []string
	Address     string // Hex address from jstack
}

// NewCPUThreadAnalyzerTool creates a new CPU thread analyzer tool
func NewCPUThreadAnalyzerTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		CPUThreadAnalyzerToolName,
		`Analyze a Java process to identify threads consuming excessive CPU resources.
This tool uses top/htop to identify high-CPU threads and matches them with jstack output to show exactly which Java threads are consuming CPU.
Use this when investigating high CPU usage in Java applications.`,
		func(ctx context.Context, params CPUThreadAnalyzerParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return analyzeCPUThreads(ctx, params)
		})
}


// analyzeCPUThreads identifies threads consuming excessive CPU resources in a Java process
func analyzeCPUThreads(ctx context.Context, params CPUThreadAnalyzerParams) (fantasy.ToolResponse, error) {
	if params.PID <= 0 {
		return fantasy.NewTextErrorResponse("PID must be greater than 0"), nil
	}

	// Get top threads by CPU usage
	topThreads, err := getTopThreadsByCPU(ctx, params.PID)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get top threads: %v", err)), nil
	}

	// Get thread dump from jstack
	threadDump, err := getThreadDump(ctx, params.PID)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get thread dump: %v", err)), nil
	}

	// Parse thread dump to get thread information
	threadMap, err := parseThreadDump(threadDump)
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to parse thread dump: %v", err)), nil
	}

	// Match top CPU threads with thread dump information
	var highCPUThreads []ThreadInfo
	for _, topThread := range topThreads {
		// Look for the thread in the thread dump
		threadInfo, found := findThreadInDump(topThread.NativeID, threadMap)
		if found {
			threadInfo.CPUUsage = topThread.CPUUsage
			highCPUThreads = append(highCPUThreads, threadInfo)
		} else {
			// If not found in thread dump, create a basic entry
			highCPUThreads = append(highCPUThreads, ThreadInfo{
				ID:         topThread.NativeID,
				Name:       fmt.Sprintf("Thread-%d", topThread.NativeID),
				NativeID:   topThread.NativeID,
				CPUUsage:   topThread.CPUUsage,
				State:      "UNKNOWN",
				StackTrace: []string{},
			})
		}
	}

	// Sort by CPU usage in descending order
	sort.Slice(highCPUThreads, func(i, j int) bool {
		return highCPUThreads[i].CPUUsage > highCPUThreads[j].CPUUsage
	})

	// Format results
	var result strings.Builder
	result.WriteString(fmt.Sprintf("=== HIGH CPU THREAD ANALYSIS FOR PID %d ===\n", params.PID))
	
	if len(highCPUThreads) == 0 {
		result.WriteString("No high CPU threads found or unable to analyze.\n")
		return fantasy.NewTextResponse(result.String()), nil
	}

	result.WriteString(fmt.Sprintf("Found %d high CPU threads:\n\n", len(highCPUThreads)))

	for i, thread := range highCPUThreads {
		result.WriteString(fmt.Sprintf("Thread #%d:\n", i+1))
		result.WriteString(fmt.Sprintf("  Name: %s\n", thread.Name))
		result.WriteString(fmt.Sprintf("  Native ID: %d (0x%x)\n", thread.NativeID, thread.NativeID))
		result.WriteString(fmt.Sprintf("  CPU Usage: %.2f%%\n", thread.CPUUsage))
		result.WriteString(fmt.Sprintf("  State: %s\n", thread.State))
		
		if len(thread.StackTrace) > 0 {
			result.WriteString("  Stack Trace (top 5 frames):\n")
			for j, frame := range thread.StackTrace {
				if j >= 5 { // Limit to top 5 frames to avoid too much output
					result.WriteString("    ... (truncated)\n")
					break
				}
				result.WriteString(fmt.Sprintf("    %s\n", frame))
			}
		}
		
		// Provide analysis based on thread state and stack trace
		analysis := analyzeThreadBehavior(thread)
		if analysis != "" {
			result.WriteString(fmt.Sprintf("  Analysis: %s\n", analysis))
		}
		
		result.WriteString("\n")
	}

	// Provide recommendations
	result.WriteString("=== RECOMMENDATIONS ===\n")
	result.WriteString(getRecommendations(highCPUThreads))

	return fantasy.NewTextResponse(result.String()), nil
}

// getTopThreadsByCPU gets the top threads by CPU usage for a process
func getTopThreadsByCPU(ctx context.Context, pid int) ([]ThreadInfo, error) {
	// Use ps to get thread information with CPU usage
	// On Linux: ps -T -p <PID> -o tid,pcpu,comm
	// On macOS: ps -M -p <PID> -o tid,pcpu,comm
	var cmd *exec.Cmd

	if isMacOS() {
		cmd = exec.CommandContext(ctx, "ps", "-M", "-p", strconv.Itoa(pid), "-o", "tid,pcpu,comm", "--no-headers")
	} else {
		cmd = exec.CommandContext(ctx, "ps", "-T", "-p", strconv.Itoa(pid), "-o", "tid,pcpu,comm", "--no-headers")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to using top command
		return getTopThreadsUsingTop(ctx, pid)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var threads []ThreadInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		nativeID, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		cpuUsage, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		// Only include threads with significant CPU usage (> 1%)
		if cpuUsage > 1.0 {
			threads = append(threads, ThreadInfo{
				NativeID: nativeID,
				CPUUsage: cpuUsage,
			})
		}
	}

	// Sort by CPU usage in descending order
	sort.Slice(threads, func(i, j int) bool {
		return threads[i].CPUUsage > threads[j].CPUUsage
	})

	// Return top 10 threads
	if len(threads) > 10 {
		threads = threads[:10]
	}

	return threads, nil
}

// getTopThreadsUsingTop gets top threads using the top command as fallback
func getTopThreadsUsingTop(ctx context.Context, pid int) ([]ThreadInfo, error) {
	// Use top command to get thread information
	cmd := exec.CommandContext(ctx, "top", "-H", "-p", strconv.Itoa(pid), "-n", "1", "-b")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run top command: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var threads []ThreadInfo

	// Look for lines that contain thread information
	// Format varies by system but typically includes PID, TID, %CPU, etc.
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// Try to parse the thread ID and CPU usage
		// The exact position may vary depending on the system
		var tidStr, cpuStr string
		
		// Look for the thread ID (usually in the second column after PID)
		// and CPU usage (usually in a column with %)
		for i, field := range fields {
			if i > 0 && isNumeric(field) {
				// Check if this looks like a thread ID (similar to PID but different)
				if i > 0 && i < len(fields)-1 {
					nextField := fields[i+1]
					if isNumeric(nextField) {
						// Check if next field looks like CPU percentage
						cpuVal, err := strconv.ParseFloat(nextField, 64)
						if err == nil && cpuVal >= 0 {
							tidStr = field
							cpuStr = nextField
							break
						}
					}
				}
			}
		}

		if tidStr != "" && cpuStr != "" {
			nativeID, err := strconv.Atoi(tidStr)
			if err != nil {
				continue
			}

			cpuUsage, err := strconv.ParseFloat(cpuStr, 64)
			if err != nil {
				continue
			}

			// Only include threads with significant CPU usage
			if cpuUsage > 1.0 {
				threads = append(threads, ThreadInfo{
					NativeID: nativeID,
					CPUUsage: cpuUsage,
				})
			}
		}
	}

	// Sort by CPU usage in descending order
	sort.Slice(threads, func(i, j int) bool {
		return threads[i].CPUUsage > threads[j].CPUUsage
	})

	// Return top 10 threads
	if len(threads) > 10 {
		threads = threads[:10]
	}

	return threads, nil
}

// getThreadDump gets the thread dump for a Java process
func getThreadDump(ctx context.Context, pid int) (string, error) {
	cmd := exec.CommandContext(ctx, "jstack", strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get thread dump: %w", err)
	}

	return string(output), nil
}

// parseThreadDump parses the thread dump to extract thread information
func parseThreadDump(threadDump string) (map[int]*ThreadInfo, error) {
	threadMap := make(map[int]*ThreadInfo)
	
	lines := strings.Split(threadDump, "\n")
	var currentThread *ThreadInfo
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Look for thread start markers
		if strings.HasPrefix(line, "\"") && strings.Contains(line, "nid=") {
			// Parse thread information
			threadInfo, err := parseThreadHeader(line)
			if err != nil {
				continue // Skip malformed thread headers
			}
			
			currentThread = &threadInfo
			threadMap[currentThread.NativeID] = currentThread
		} else if currentThread != nil && strings.HasPrefix(line, "\tat ") {
			// Add stack frame to current thread
			frame := strings.TrimPrefix(line, "\tat ")
			currentThread.StackTrace = append(currentThread.StackTrace, frame)
		} else if currentThread != nil && strings.HasPrefix(line, "\t- ") {
			// Add lock information
			lockInfo := strings.TrimSpace(line)
			currentThread.StackTrace = append(currentThread.StackTrace, lockInfo)
		} else if currentThread != nil && line == "" && i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "\"") {
			// End of current thread's stack trace
			currentThread = nil
		}
	}
	
	return threadMap, nil
}

// parseThreadHeader parses a thread header line from jstack output
func parseThreadHeader(header string) (ThreadInfo, error) {
	threadInfo := ThreadInfo{}
	
	// Extract thread name (everything between quotes)
	nameRegex := regexp.MustCompile(`^"([^"]+)"`)
	nameMatches := nameRegex.FindStringSubmatch(header)
	if len(nameMatches) > 1 {
		threadInfo.Name = nameMatches[1]
	}
	
	// Extract native thread ID (nid)
	nidRegex := regexp.MustCompile(`nid=0x([0-9a-fA-F]+)|nid=(\d+)`)
	nidMatches := nidRegex.FindStringSubmatch(header)
	if len(nidMatches) > 0 {
		if nidMatches[1] != "" {
			// Hex format
			nid, err := strconv.ParseInt(nidMatches[1], 16, 64)
			if err == nil {
				threadInfo.NativeID = int(nid)
			}
		} else if nidMatches[2] != "" {
			// Decimal format
			nid, err := strconv.Atoi(nidMatches[2])
			if err == nil {
				threadInfo.NativeID = nid
			}
		}
	}
	
	// Extract thread state
	stateRegex := regexp.MustCompile(`java\.lang\.Thread\.State: ([A-Z_]+)`)
	stateMatches := stateRegex.FindStringSubmatch(header)
	if len(stateMatches) > 1 {
		threadInfo.State = stateMatches[1]
	}
	
	return threadInfo, nil
}

// findThreadInDump finds a thread in the parsed thread dump by native ID
func findThreadInDump(nativeID int, threadMap map[int]*ThreadInfo) (ThreadInfo, bool) {
	thread, exists := threadMap[nativeID]
	if !exists {
		return ThreadInfo{}, false
	}
	
	return *thread, true
}

// analyzeThreadBehavior analyzes thread behavior based on state and stack trace
func analyzeThreadBehavior(thread ThreadInfo) string {
	switch thread.State {
	case "RUNNABLE":
		// Check stack trace for common CPU-intensive operations
		for _, frame := range thread.StackTrace {
			if strings.Contains(frame, "sun.nio.ch") || strings.Contains(frame, "java.nio") {
				return "Thread is performing intensive I/O operations"
			}
			if strings.Contains(frame, "java.util.zip") || strings.Contains(frame, "java.util.jar") {
				return "Thread is performing compression/decompression operations"
			}
			if strings.Contains(frame, "java.security") || strings.Contains(frame, "javax.crypto") {
				return "Thread is performing cryptographic operations"
			}
			if strings.Contains(frame, "java.math.BigInteger") {
				return "Thread is performing heavy mathematical computations"
			}
			if strings.Contains(frame, "java.util.regex") {
				return "Thread is performing intensive regex operations"
			}
		}
		return "Thread is actively running and consuming CPU"
	case "BLOCKED":
		return "Thread is blocked waiting for a monitor lock - may indicate contention"
	case "WAITING", "TIMED_WAITING":
		return "Thread is waiting - high CPU usage may be misleading if recently woke up"
	default:
		return "Thread is in an unusual state"
	}
}

// getRecommendations provides recommendations based on high CPU threads
func getRecommendations(threads []ThreadInfo) string {
	var recommendations strings.Builder
	
	highCPUCount := 0
	for _, thread := range threads {
		if thread.CPUUsage > 50.0 {
			highCPUCount++
		}
	}
	
	if highCPUCount > 0 {
		recommendations.WriteString(fmt.Sprintf("- Found %d thread(s) with very high CPU usage (>50%%)\n", highCPUCount))
		recommendations.WriteString("  Consider profiling these threads to identify hotspots\n")
		recommendations.WriteString("  Use: jstack <PID> to get detailed thread information\n")
	}
	
	// Check for specific patterns in stack traces
	for _, thread := range threads {
		for _, frame := range thread.StackTrace {
			if strings.Contains(frame, "sun.nio.ch") || strings.Contains(frame, "java.nio") {
				recommendations.WriteString("- Intensive I/O operations detected: Consider optimizing I/O patterns or using async I/O\n")
				break
			}
			if strings.Contains(frame, "java.util.zip") || strings.Contains(frame, "java.util.jar") {
				recommendations.WriteString("- Heavy compression/decompression detected: Consider caching or optimizing algorithms\n")
				break
			}
			if strings.Contains(frame, "java.security") || strings.Contains(frame, "javax.crypto") {
				recommendations.WriteString("- Cryptographic operations consuming CPU: Consider algorithm optimization or caching\n")
				break
			}
			if strings.Contains(frame, "java.math.BigInteger") {
				recommendations.WriteString("- Heavy mathematical computations detected: Consider algorithm optimization\n")
				break
			}
			if strings.Contains(frame, "java.util.regex") {
				recommendations.WriteString("- Intensive regex operations detected: Consider precompiling patterns or optimizing expressions\n")
				break
			}
		}
	}
	
	if recommendations.Len() == 0 {
		recommendations.WriteString("- No specific issues detected in thread analysis\n")
		recommendations.WriteString("- Consider using a profiler like VisualVM or JProfiler for deeper analysis\n")
		recommendations.WriteString("- Monitor GC activity with: jstat -gc <PID>\n")
	}
	
	return recommendations.String()
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isMacOS checks if the current OS is macOS
func isMacOS() bool {
	// This is a simple check - in a real implementation, you'd use runtime.GOOS
	cmd := exec.Command("uname")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.TrimSpace(string(output)), "Darwin")
}