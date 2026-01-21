package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"charm.land/fantasy"
)

// CommandExecutor provides tools for executing diagnostic commands
type CommandExecutor struct {
	timeout time.Duration
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &CommandExecutor{
		timeout: timeout,
	}
}

// PSCommandParams are parameters for the ps command
type PSCommandParams struct {
	PID int `json:"pid,omitempty" description:"Process ID to get info for (optional - if not provided, gets all processes)"`
}

// PSCommandToolName is the name of the ps command tool
const PSCommandToolName = "ps_command"

// NewPSCommandTool creates a new ps command tool
func NewPSCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		PSCommandToolName,
		`Execute the ps command to get process information. This tool can get information for a specific process or all processes.
Use this when you need to identify running processes, their resource usage, or basic process information.`,
		func(ctx context.Context, params PSCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runPSCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runPSCommand(ctx context.Context, params PSCommandParams) (fantasy.ToolResponse, error) {
	var cmd *exec.Cmd
	
	if params.PID > 0 {
		// Get info for specific process
		cmd = exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(params.PID), "-o", "pid,ppid,comm,%cpu,%mem,vsz,rss,wchan:50,etime,pcpu,pmem", "--no-headers")
	} else {
		// Get all processes
		cmd = exec.CommandContext(ctx, "ps", "axo", "pid,ppid,comm,%cpu,%mem,vsz,rss,wchan:50,etime,pcpu,pmem", "--no-headers")
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run ps command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// JStackCommandParams are parameters for the jstack command
type JStackCommandParams struct {
	PID int `json:"pid" description:"Process ID of the Java application to get thread dump for"`
}

// JStackCommandToolName is the name of the jstack command tool
const JStackCommandToolName = "jstack_command"

// NewJStackCommandTool creates a new jstack command tool
func NewJStackCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JStackCommandToolName,
		`Execute the jstack command to get a thread dump for a Java process. This tool is useful for identifying stuck threads, deadlocks, and thread contention issues.
Use this when investigating high CPU usage, application hangs, or suspected deadlocks.`,
		func(ctx context.Context, params JStackCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runJStackCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runJStackCommand(ctx context.Context, params JStackCommandParams) (fantasy.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "jstack", strconv.Itoa(params.PID))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run jstack command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// JStatCommandParams are parameters for the jstat command
type JStatCommandParams struct {
	PID    int    `json:"pid" description:"Process ID of the Java application to get statistics for"`
	Option string `json:"option" description:"jstat option to use (e.g., -gc, -gccapacity, -gcutil)"`
}

// JStatCommandToolName is the name of the jstat command tool
const JStatCommandToolName = "jstat_command"

// NewJStatCommandTool creates a new jstat command tool
func NewJStatCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JStatCommandToolName,
		`Execute the jstat command to get JVM statistics for a Java process. This tool can provide information about garbage collection, memory usage, and other JVM metrics.
Use this when investigating memory issues, garbage collection problems, or performance issues.`,
		func(ctx context.Context, params JStatCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runJStatCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runJStatCommand(ctx context.Context, params JStatCommandParams) (fantasy.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "jstat", params.Option, strconv.Itoa(params.PID))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run jstat command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// JMapCommandParams are parameters for the jmap command
type JMapCommandParams struct {
	PID int `json:"pid" description:"Process ID of the Java application to get memory information for"`
}

// JMapCommandToolName is the name of the jmap command tool
const JMapCommandToolName = "jmap_command"

// NewJMapCommandTool creates a new jmap command tool
func NewJMapCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JMapCommandToolName,
		`Execute the jmap command to get memory mapping information for a Java process. This tool can provide information about heap usage, object counts, and memory distribution.
Use this when investigating memory leaks, OutOfMemoryError issues, or high memory usage.`,
		func(ctx context.Context, params JMapCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runJMapCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runJMapCommand(ctx context.Context, params JMapCommandParams) (fantasy.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "jmap", "-heap", strconv.Itoa(params.PID))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run jmap command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// JInfoCommandParams are parameters for the jinfo command
type JInfoCommandParams struct {
	PID int `json:"pid" description:"Process ID of the Java application to get configuration info for"`
}

// JInfoCommandToolName is the name of the jinfo command tool
const JInfoCommandToolName = "jinfo_command"

// NewJInfoCommandTool creates a new jinfo command tool
func NewJInfoCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JInfoCommandToolName,
		`Execute the jinfo command to get configuration information for a Java process. This tool can provide information about JVM flags and system properties.
Use this when investigating JVM configuration issues or understanding the runtime environment.`,
		func(ctx context.Context, params JInfoCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runJInfoCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runJInfoCommand(ctx context.Context, params JInfoCommandParams) (fantasy.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "jinfo", "-flags", strconv.Itoa(params.PID))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run jinfo command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// TopCommandParams are parameters for the top command
type TopCommandParams struct {
	PID int `json:"pid,omitempty" description:"Process ID to monitor (optional - if not provided, shows all processes)"`
	Iterations int `json:"iterations,omitempty" description:"Number of iterations to run (default: 1)"`
}

// TopCommandToolName is the name of the top command tool
const TopCommandToolName = "top_command"

// NewTopCommandTool creates a new top command tool
func NewTopCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TopCommandToolName,
		`Execute the top/htop command to get real-time process information. This tool can show CPU and memory usage for processes.
Use this when investigating high CPU usage, memory consumption, or general process performance.`,
		func(ctx context.Context, params TopCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runTopCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runTopCommand(ctx context.Context, params TopCommandParams) (fantasy.ToolResponse, error) {
	var cmd *exec.Cmd
	
	if params.PID > 0 {
		// Use different command based on OS
		if runtime.GOOS == "darwin" { // macOS
			cmd = exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(params.PID), "-o", "pid,ppid,comm,%cpu,%mem,vsz,rss,time,etime", "--no-headers")
		} else { // Linux
			cmd = exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(params.PID), "-o", "pid,ppid,comm,%cpu,%mem,vsz,rss,time,etime", "--no-headers")
		}
	} else {
		// Get top processes
		if runtime.GOOS == "darwin" { // macOS
			cmd = exec.CommandContext(ctx, "ps", "aux", "-o", "pid,ppid,comm,%cpu,%mem,vsz,rss,time,etime", "--no-headers", "|", "head", "-20")
		} else { // Linux
			cmd = exec.CommandContext(ctx, "ps", "aux", "-o", "pid,ppid,comm,%cpu,%mem,vsz,rss,time,etime", "--no-headers")
		}
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run top command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// LsofCommandParams are parameters for the lsof command
type LsofCommandParams struct {
	PID int `json:"pid" description:"Process ID to get open files and network connections for"`
}

// LsofCommandToolName is the name of the lsof command tool
const LsofCommandToolName = "lsof_command"

// NewLsofCommandTool creates a new lsof command tool
func NewLsofCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		LsofCommandToolName,
		`Execute the lsof command to get information about open files and network connections for a process. This tool can help identify resource leaks and connection issues.
Use this when investigating file descriptor leaks, network connection issues, or resource usage.`,
		func(ctx context.Context, params LsofCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runLsofCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runLsofCommand(ctx context.Context, params LsofCommandParams) (fantasy.ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-p", strconv.Itoa(params.PID))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run lsof command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// NetstatCommandParams are parameters for the netstat command
type NetstatCommandParams struct {
	PID int `json:"pid,omitempty" description:"Process ID to get network connections for (optional - if not provided, shows all connections)"`
}

// NetstatCommandToolName is the name of the netstat command tool
const NetstatCommandToolName = "netstat_command"

// NewNetstatCommandTool creates a new netstat command tool
func NewNetstatCommandTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		NetstatCommandToolName,
		`Execute the netstat command to get network connection information. This tool can help identify network issues and connection patterns.
Use this when investigating network performance, connection issues, or port usage.`,
		func(ctx context.Context, params NetstatCommandParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			executor := NewCommandExecutor(10 * time.Second)
			return executor.runNetstatCommand(ctx, params)
		})
}

func (ce *CommandExecutor) runNetstatCommand(ctx context.Context, params NetstatCommandParams) (fantasy.ToolResponse, error) {
	var cmd *exec.Cmd
	
	if params.PID > 0 {
		// Get network connections for specific process
		cmd = exec.CommandContext(ctx, "lsof", "-i", "-p", strconv.Itoa(params.PID))
	} else {
		// Get all network connections
		cmd = exec.CommandContext(ctx, "netstat", "-tuln")
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to run netstat command: %v", err)), nil
	}
	
	return fantasy.NewTextResponse(string(output)), nil
}

// GetCommandTools returns all command-based diagnostic tools
func GetCommandTools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		NewPSCommandTool(),
		NewJStackCommandTool(),
		NewJStatCommandTool(),
		NewJMapCommandTool(),
		NewJInfoCommandTool(),
		NewTopCommandTool(),
		NewLsofCommandTool(),
		NewNetstatCommandTool(),
	}
}