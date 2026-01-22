package tools

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/fantasy"
)

// NetworkDiagnoseParams are parameters for the network diagnostic tool
type NetworkDiagnoseParams struct {
	PID       int    `json:"pid,omitempty" description:"Process ID to diagnose (optional - if not provided, analyzes all network activity)"`
	Host      string `json:"host,omitempty" description:"Host to diagnose network connectivity to"`
	Port      int    `json:"port,omitempty" description:"Port to diagnose connectivity to"`
	Interface string `json:"interface,omitempty" description:"Network interface to diagnose (optional)"`
}

// NetworkDiagnoseToolName is the name of the network diagnostic tool
const NetworkDiagnoseToolName = "network_diagnose"

// NewNetworkDiagnoseTool creates a new network diagnostic tool
func NewNetworkDiagnoseTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		NetworkDiagnoseToolName,
		`Diagnose network issues including connectivity problems, high network usage, connection leaks, and network bottlenecks.
This tool can analyze network activity for a specific process or general network connectivity issues.
Use this when investigating slow network performance, connection timeouts, or network-related application issues.`,
		func(ctx context.Context, params NetworkDiagnoseParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return runNetworkDiagnosis(ctx, params)
		})
}

// NetworkDiagnosticResult contains the results of a network diagnostic
type NetworkDiagnosticResult struct {
	Timestamp       time.Time         `json:"timestamp"`
	ProcessID       int               `json:"process_id,omitempty"`
	TargetHost      string            `json:"target_host,omitempty"`
	TargetPort      int               `json:"target_port,omitempty"`
	Interface       string            `json:"interface,omitempty"`
	Connections     []ConnectionInfo  `json:"connections"`
	NetworkStats    NetworkStats      `json:"network_stats"`
	Issues          []NetworkIssue    `json:"issues"`
	Recommendations []string          `json:"recommendations"`
}

// ConnectionInfo contains information about a network connection
type ConnectionInfo struct {
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	State         string `json:"state"`
	ProcessName   string `json:"process_name"`
	PID           int    `json:"pid"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
}

// NetworkStats contains network statistics
type NetworkStats struct {
	TotalConnections int `json:"total_connections"`
	ActiveConnections int `json:"active_connections"`
	TimeWaitCount     int `json:"time_wait_count"`
	CloseWaitCount    int `json:"close_wait_count"`
	HighNetworkUsage  bool `json:"high_network_usage"`
}

// NetworkIssue represents a network issue found during diagnosis
type NetworkIssue struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // low, medium, high, critical
	Suggestion  string `json:"suggestion"`
}

// runNetworkDiagnosis performs network diagnostics based on the provided parameters
func runNetworkDiagnosis(ctx context.Context, params NetworkDiagnoseParams) (fantasy.ToolResponse, error) {
	var result strings.Builder
	
	result.WriteString("=== NETWORK DIAGNOSTIC RESULT ===\n")
	result.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	if params.PID > 0 {
		// Diagnose network activity for specific process
		result.WriteString(fmt.Sprintf("Diagnosing network activity for PID: %d\n\n", params.PID))
		
		connections, err := getProcessNetworkConnections(ctx, params.PID)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get network connections for PID %d: %v", params.PID, err)), nil
		}
		
		stats := analyzeNetworkConnections(connections)
		
		result.WriteString(fmt.Sprintf("Found %d connections:\n", len(connections)))
		for _, conn := range connections {
			result.WriteString(fmt.Sprintf("  %s -> %s [%s]\n", conn.LocalAddress, conn.RemoteAddress, conn.State))
		}
		
		result.WriteString("\nNetwork Statistics:\n")
		result.WriteString(fmt.Sprintf("  Total Connections: %d\n", stats.TotalConnections))
		result.WriteString(fmt.Sprintf("  Active Connections: %d\n", stats.ActiveConnections))
		result.WriteString(fmt.Sprintf("  TIME_WAIT Count: %d\n", stats.TimeWaitCount))
		result.WriteString(fmt.Sprintf("  CLOSE_WAIT Count: %d\n", stats.CloseWaitCount))
		result.WriteString(fmt.Sprintf("  High Network Usage: %t\n", stats.HighNetworkUsage))
		
		// Identify issues
		issues := identifyNetworkIssues(connections, stats)
		if len(issues) > 0 {
			result.WriteString("\nIssues Found:\n")
			for _, issue := range issues {
				result.WriteString(fmt.Sprintf("  [%s] %s: %s\n    Suggestion: %s\n",
					issue.Severity, issue.Type, issue.Description, issue.Suggestion))
			}
		} else {
			result.WriteString("\nNo significant network issues detected.\n")
		}

		// Provide recommendations
		recommendations := getNetworkRecommendations(issues)
		if len(recommendations) > 0 {
			result.WriteString("\nRecommendations:\n")
			for _, rec := range recommendations {
				result.WriteString(fmt.Sprintf("  - %s\n", rec))
			}
		}
	} else if params.Host != "" {
		// Diagnose connectivity to specific host
		result.WriteString(fmt.Sprintf("Diagnosing connectivity to host: %s", params.Host))
		if params.Port > 0 {
			result.WriteString(fmt.Sprintf(":%d", params.Port))
		}
		result.WriteString("\n\n")
		
		connectivityResult, err := diagnoseConnectivity(ctx, params.Host, params.Port)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to diagnose connectivity to %s:%d: %v", params.Host, params.Port, err)), nil
		}
		
		result.WriteString(connectivityResult)
	} else if params.Interface != "" {
		// Diagnose specific network interface
		result.WriteString(fmt.Sprintf("Diagnosing network interface: %s\n\n", params.Interface))
		
		interfaceStats, err := getInterfaceStats(ctx, params.Interface)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get interface stats for %s: %v", params.Interface, err)), nil
		}
		
		result.WriteString(interfaceStats)
	} else {
		// Diagnose general network activity
		result.WriteString("Diagnosing general network activity\n\n")
		
		allConnections, err := getAllNetworkConnections(ctx)
		if err != nil {
			return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get network connections: %v", err)), nil
		}
		
		stats := analyzeNetworkConnections(allConnections)
		
		result.WriteString(fmt.Sprintf("Found %d total connections\n", len(allConnections)))
		result.WriteString(fmt.Sprintf("Active connections: %d\n", stats.ActiveConnections))
		result.WriteString(fmt.Sprintf("TIME_WAIT connections: %d\n", stats.TimeWaitCount))
		result.WriteString(fmt.Sprintf("CLOSE_WAIT connections: %d\n", stats.CloseWaitCount))
		
		// Show top connection counts by process
		connByProcess := groupConnectionsByProcess(allConnections)
		result.WriteString("\nTop processes by connection count:\n")
		for _, proc := range connByProcess {
			result.WriteString(fmt.Sprintf("  PID %d (%s): %d connections\n", proc.pid, proc.name, proc.count))
		}
		
		// Identify issues
		issues := identifyNetworkIssues(allConnections, stats)
		if len(issues) > 0 {
			result.WriteString("\nIssues Found:\n")
			for _, issue := range issues {
				result.WriteString(fmt.Sprintf("  [%s] %s: %s\n    Suggestion: %s\n", 
					issue.Severity, issue.Type, issue.Description, issue.Suggestion))
			}
		} else {
			result.WriteString("\nNo significant network issues detected.\n")
		}
		
		// Provide recommendations
		recommendations := getNetworkRecommendations(issues)
		if len(recommendations) > 0 {
			result.WriteString("\nRecommendations:\n")
			for _, rec := range recommendations {
				result.WriteString(fmt.Sprintf("  - %s\n", rec))
			}
		}
	}

	return fantasy.NewTextResponse(result.String()), nil
}

// getProcessNetworkConnections gets network connections for a specific process
func getProcessNetworkConnections(ctx context.Context, pid int) ([]ConnectionInfo, error) {
	var cmd *exec.Cmd
	
	// Use lsof to get network connections for the process
	cmd = exec.CommandContext(ctx, "lsof", "-i", "-p", strconv.Itoa(pid), "-n", "-P", "-t", "-a")
	output, err := cmd.Output()
	if err != nil {
		// lsof might not be available, try using netstat instead
		cmd = exec.CommandContext(ctx, "netstat", "-tulnp", fmt.Sprintf("--pid=%d", pid))
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get network connections: %w", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var connections []ConnectionInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse connection information from lsof or netstat output
		conn, err := parseConnectionInfo(line, pid)
		if err != nil {
			continue // Skip malformed lines
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// getAllNetworkConnections gets all network connections on the system
func getAllNetworkConnections(ctx context.Context) ([]ConnectionInfo, error) {
	var cmd *exec.Cmd
	
	// Try lsof first
	cmd = exec.CommandContext(ctx, "lsof", "-i", "-n", "-P", "-t", "-a")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to netstat
		cmd = exec.CommandContext(ctx, "netstat", "-tulnp")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get network connections: %w", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var connections []ConnectionInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse connection information
		conn, err := parseConnectionInfo(line, 0) // PID 0 means we'll extract it from the line
		if err != nil {
			continue // Skip malformed lines
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

// parseConnectionInfo parses a line from lsof or netstat output to extract connection information
func parseConnectionInfo(line string, defaultPID int) (ConnectionInfo, error) {
	fields := strings.Fields(line)
	
	if len(fields) < 2 {
		return ConnectionInfo{}, fmt.Errorf("invalid connection line format")
	}
	
	conn := ConnectionInfo{
		PID: defaultPID,
	}

	// Different parsing based on command output format
	if strings.Contains(line, "lsof") || len(fields) > 3 {
		// This looks like lsof output: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		// NAME format: TCP *:22 (LISTEN) or TCP 192.168.1.10:54321->192.168.1.1:80 (ESTABLISHED)

		// Find the network connection part (usually the last field)
		for i, field := range fields {
			if strings.Contains(field, "->") || strings.Contains(field, ":") {
				parts := strings.Split(field, "->")
				if len(parts) == 2 {
					conn.LocalAddress = strings.TrimSpace(parts[0])
					conn.RemoteAddress = strings.TrimSpace(parts[1])
				} else {
					conn.LocalAddress = field
				}

				// Extract PID from lsof output
				if i > 1 {
					if pid, err := strconv.Atoi(fields[1]); err == nil {
						conn.PID = pid
					}
				}

				// Extract process name
				conn.ProcessName = fields[0]

				// Extract state if available
				if strings.Contains(field, "(") && strings.Contains(field, ")") {
					start := strings.Index(field, "(")
					end := strings.Index(field, ")")
					if start != -1 && end != -1 && end > start {
						conn.State = strings.TrimSpace(field[start+1:end])
					}
				}

				break
			}
		}
	} else {
		// This looks like netstat output
		// Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
		if len(fields) >= 7 {
			conn.LocalAddress = fields[3]
			conn.RemoteAddress = fields[4]
			conn.State = fields[5]

			// Extract PID and program name from last field
			pidProgram := fields[6]
			parts := strings.Split(pidProgram, "/")
			if len(parts) == 2 {
				if pid, err := strconv.Atoi(parts[0]); err == nil {
					conn.PID = pid
				}
				conn.ProcessName = parts[1]
			}
		}
	}

	return conn, nil
}

// analyzeNetworkConnections analyzes network connections to identify potential issues
func analyzeNetworkConnections(connections []ConnectionInfo) NetworkStats {
	stats := NetworkStats{
		TotalConnections:  len(connections),
		ActiveConnections: 0,
		TimeWaitCount:     0,
		CloseWaitCount:    0,
		HighNetworkUsage:  false,
	}
	
	activeStates := map[string]bool{
		"ESTABLISHED": true,
		"SYN_SENT":    true,
		"SYN_RECV":    true,
		"FIN_WAIT1":   true,
		"FIN_WAIT2":   true,
		"CLOSING":     true,
		"LAST_ACK":    true,
		"TIME_WAIT":   true,
	}
	
	for _, conn := range connections {
		if activeStates[strings.ToUpper(conn.State)] {
			stats.ActiveConnections++
		}
		
		if strings.ToUpper(conn.State) == "TIME_WAIT" {
			stats.TimeWaitCount++
		} else if strings.ToUpper(conn.State) == "CLOSE_WAIT" {
			stats.CloseWaitCount++
		}
	}
	
	// Heuristic for high network usage: if there are many connections (> 100) or many TIME_WAIT/CLOSE_WAIT connections
	if stats.TotalConnections > 100 || stats.TimeWaitCount > 50 || stats.CloseWaitCount > 10 {
		stats.HighNetworkUsage = true
	}
	
	return stats
}

// identifyNetworkIssues identifies potential network issues from connections and stats
func identifyNetworkIssues(connections []ConnectionInfo, stats NetworkStats) []NetworkIssue {
	var issues []NetworkIssue
	
	// Check for connection leaks (too many TIME_WAIT or CLOSE_WAIT connections)
	if stats.TimeWaitCount > 50 {
		issues = append(issues, NetworkIssue{
			Type:        "ConnectionLeak",
			Description: fmt.Sprintf("High number of TIME_WAIT connections (%d), indicating potential connection leaks", stats.TimeWaitCount),
			Severity:    "high",
			Suggestion:  "Review application code for proper connection closing, consider adjusting TCP keepalive settings",
		})
	}
	
	if stats.CloseWaitCount > 10 {
		issues = append(issues, NetworkIssue{
			Type:        "ConnectionLeak",
			Description: fmt.Sprintf("High number of CLOSE_WAIT connections (%d), indicating potential connection leaks", stats.CloseWaitCount),
			Severity:    "high",
			Suggestion:  "Review application code for proper connection closing, ensure connections are closed in all code paths",
		})
	}
	
	// Check for too many connections
	if stats.TotalConnections > 200 {
		issues = append(issues, NetworkIssue{
			Type:        "HighConnectionCount",
			Description: fmt.Sprintf("Very high number of connections (%d), which may cause resource exhaustion", stats.TotalConnections),
			Severity:    "medium",
			Suggestion:  "Consider implementing connection pooling, limiting concurrent connections, or optimizing connection reuse",
		})
	}
	
	// Check for connections to suspicious hosts
	for _, conn := range connections {
		if isSuspiciousHost(conn.RemoteAddress) {
			issues = append(issues, NetworkIssue{
				Type:        "SuspiciousConnection",
				Description: fmt.Sprintf("Connection to potentially suspicious host: %s", conn.RemoteAddress),
				Severity:    "high",
				Suggestion:  "Investigate this connection, it may indicate unauthorized access or malware",
			})
		}
	}
	
	// Check for connections in problematic states
	problematicStates := map[string]string{
		"CLOSE_WAIT": "Remote side closed connection but local side hasn't acknowledged",
		"FIN_WAIT1":  "Connection termination initiated but not completed",
		"FIN_WAIT2":  "Connection termination in progress",
		"CLOSING":    "Both sides trying to terminate connection simultaneously",
	}
	
	for state, description := range problematicStates {
		count := 0
		for _, conn := range connections {
			if strings.ToUpper(conn.State) == state {
				count++
			}
		}
		
		if count > 10 { // More than 10 connections in problematic state
			issues = append(issues, NetworkIssue{
				Type:        "ProblematicConnectionState",
				Description: fmt.Sprintf("%d connections in %s state: %s", count, state, description),
				Severity:    "medium",
				Suggestion:  "Investigate network connectivity issues or application connection handling",
			})
		}
	}
	
	return issues
}

// isSuspiciousHost checks if a host address is potentially suspicious
func isSuspiciousHost(address string) bool {
	// Check for private IP ranges that shouldn't be connecting externally
	privateRanges := []string{
		"10.",
		"172.16.",
		"172.17.",
		"172.18.",
		"172.19.",
		"172.20.",
		"172.21.",
		"172.22.",
		"172.23.",
		"172.24.",
		"172.25.",
		"172.26.",
		"172.27.",
		"172.28.",
		"172.29.",
		"172.30.",
		"172.31.",
		"192.168.",
	}
	
	for _, privateRange := range privateRanges {
		if strings.Contains(address, privateRange) {
			// This might be suspicious if it's an outbound connection to a private IP
			// In a real implementation, we'd need more context to determine if this is truly suspicious
			return strings.Contains(address, "->") // Only flag if it's an outbound connection
		}
	}
	
	// Check for known bad ports or patterns
	badPatterns := []string{
		":22",   // SSH - might be suspicious if not expected
		":3389", // RDP - might be suspicious if not expected
		":445",  // SMB - often exploited
		":139",  // NetBIOS - often exploited
	}
	
	for _, pattern := range badPatterns {
		if strings.Contains(address, pattern) {
			return true
		}
	}
	
	return false
}

// getNetworkRecommendations provides recommendations based on identified issues
func getNetworkRecommendations(issues []NetworkIssue) []string {
	var recommendations []string
	
	seenTypes := make(map[string]bool)
	for _, issue := range issues {
		if !seenTypes[issue.Type] {
			seenTypes[issue.Type] = true
			
			switch issue.Type {
			case "ConnectionLeak":
				recommendations = append(recommendations, 
					"Implement proper connection lifecycle management with defer statements to ensure connections are closed",
					"Use connection pooling to reuse connections efficiently",
					"Set appropriate timeouts for connections to prevent indefinite hanging")
			case "HighConnectionCount":
				recommendations = append(recommendations,
					"Implement connection limits to prevent resource exhaustion",
					"Use connection pooling to reduce the number of concurrent connections",
					"Optimize application logic to reduce unnecessary connections")
			case "SuspiciousConnection":
				recommendations = append(recommendations,
					"Review firewall rules to restrict outbound connections",
					"Implement network monitoring to detect unauthorized access",
					"Investigate the application code that initiated this connection")
			case "ProblematicConnectionState":
				recommendations = append(recommendations,
					"Review network connectivity and stability",
					"Implement proper error handling for connection failures",
					"Adjust TCP timeout settings if appropriate")
			}
		}
	}
	
	// Add general recommendations if no specific issues were found
	if len(issues) == 0 {
		recommendations = append(recommendations,
			"Monitor network connections regularly for anomalies",
			"Implement proper connection lifecycle management",
			"Use secure protocols (TLS/SSL) for all network communications")
	}
	
	return recommendations
}

// diagnoseConnectivity diagnoses connectivity to a specific host/port
func diagnoseConnectivity(ctx context.Context, host string, port int) (string, error) {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("Testing connectivity to %s", host))
	if port > 0 {
		result.WriteString(fmt.Sprintf(":%d", port))
	}
	result.WriteString("\n\n")
	
	// Test DNS resolution
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		result.WriteString(fmt.Sprintf("❌ DNS resolution failed: %v\n", err))
		return result.String(), nil
	}
	
	result.WriteString(fmt.Sprintf("✓ DNS resolution: %s -> %s\n", host, ipAddr.IP.String()))
	
	// Test port connectivity if port is specified
	if port > 0 {
		address := fmt.Sprintf("%s:%d", ipAddr.IP.String(), port)
		
		conn, err := net.DialTimeout("tcp", address, 10*time.Second)
		if err != nil {
			result.WriteString(fmt.Sprintf("❌ Connection to %s failed: %v\n", address, err))
		} else {
			result.WriteString(fmt.Sprintf("✓ Connection to %s successful\n", address))
			conn.Close()
		}
	} else {
		result.WriteString("No specific port to test, only DNS resolution performed.\n")
	}
	
	return result.String(), nil
}

// getInterfaceStats gets statistics for a specific network interface
func getInterfaceStats(ctx context.Context, iface string) (string, error) {
	var result strings.Builder
	
	// Try to get interface statistics using ifconfig or ip command
	var cmd *exec.Cmd
	if isMacOS() {
		cmd = exec.CommandContext(ctx, "ifconfig", iface)
	} else {
		cmd = exec.CommandContext(ctx, "ip", "addr", "show", iface)
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get interface stats: %w", err)
	}
	
	result.WriteString(fmt.Sprintf("Interface %s statistics:\n", iface))
	result.WriteString(string(output))
	
	// Additional interface statistics using netstat or ss
	var statsCmd *exec.Cmd
	if isMacOS() {
		statsCmd = exec.CommandContext(ctx, "netstat", "-i", "-b")
	} else {
		statsCmd = exec.CommandContext(ctx, "ss", "-i")
	}
	
	statsOutput, err := statsCmd.Output()
	if err == nil {
		result.WriteString("\nDetailed interface statistics:\n")
		result.WriteString(string(statsOutput))
	}
	
	return result.String(), nil
}


// groupConnectionsByProcess groups connections by process
func groupConnectionsByProcess(connections []ConnectionInfo) []struct {
	pid   int
	name  string
	count int
} {
	processMap := make(map[int]struct {
		name  string
		count int
	})

	for _, conn := range connections {
		if conn.PID > 0 {
			if data, exists := processMap[conn.PID]; exists {
				data.count++
				processMap[conn.PID] = data
			} else {
				processMap[conn.PID] = struct {
					name  string
					count int
				}{name: conn.ProcessName, count: 1}
			}
		}
	}

	// Convert to slice and sort by count
	var processes []struct {
		pid   int
		name  string
		count int
	}

	for pid, data := range processMap {
		processes = append(processes, struct {
			pid   int
			name  string
			count int
		}{pid: pid, name: data.name, count: data.count})
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].count > processes[j].count
	})

	return processes
}

