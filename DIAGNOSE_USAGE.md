# Diagnose Tool Usage Guide

The Diagnose tool is a multi-language diagnostic framework that can identify and help resolve online application issues including OutOfMemoryError, high CPU usage, and other performance problems.

## Overview

The diagnose framework is designed to be extensible and support multiple programming languages. Currently, it includes support for Java applications with plans to expand to Go, C++, Rust, and other languages.

## Installation

The diagnose tool is integrated into the `crush` CLI. Simply build the project:

```bash
cd /path/to/crush
go build -o crush .
```

## Commands

### Basic Usage

```bash
# List all diagnosable processes
./crush java-diag

# Diagnose a specific process by PID
./crush java-diag --pid <PID>

# Diagnose all Java processes
./crush java-diag --all

# Show only OOM-related diagnostics
./crush java-diag --oom-only

# Show only CPU-related diagnostics
./crush java-diag --cpu-only
```

### Command Options

| Option | Description |
|--------|-------------|
| `--pid <PID>` | Diagnose a specific process by its Process ID |
| `--all` | Diagnose all running Java processes |
| `--oom-only` | Show only OutOfMemoryError related diagnostics |
| `--cpu-only` | Show only high CPU usage related diagnostics |
| `-h`, `--help` | Show help information |

## Supported Languages

### Java Diagnostics

The Java diagnoser can detect:

#### Memory Issues
- **OOM (OutOfMemoryError)**: Detects when memory usage approaches critical levels
- **High memory consumption**: Identifies processes using excessive memory

#### CPU Issues
- **High CPU usage**: Flags processes consuming >80% CPU consistently
- **Performance bottlenecks**: Identifies processes causing system slowdown

#### Thread Issues
- **Stuck threads**: Detects blocked or deadlocked threads
- **Thread contention**: Identifies excessive thread synchronization issues

#### Garbage Collection Issues
- **Frequent GC**: Detects excessive garbage collection activity
- **Full GC problems**: Identifies problematic full garbage collection cycles

## Output Format

The diagnostic output includes:

```
=== DIAGNOSTIC RESULT FOR PID <PID> ===
Name: <process name>
Type: <language type>
Timestamp: <timestamp>
CPU Usage: <percentage>%
Memory Used: <amount>
Memory Max: <amount>
Potential OOM: <true/false>
High CPU Usage: <true/false>
Stuck Threads Detected: <true/false>
Frequent GC Activity: <true/false>
Full GC Issues: <true/false>

Issues Found:
  [<severity>] <issue-type>: <description>
      Suggestion: <recommendation>

Recommendation: <specific action to take>
```

## Exit Codes

- `0`: Success - diagnostics completed
- `1`: Error - failed to diagnose process or invalid parameters

## Troubleshooting

### Common Issues

1. **"Process not found"**: The process may have terminated or the PID is incorrect
2. **Permission denied**: The tool may not have sufficient permissions to access process information
3. **No Java processes found**: No Java applications are currently running

### Required Tools

The Java diagnoser requires these tools to be available in the system PATH:
- `jstack` - for thread dump analysis
- `jstat` - for garbage collection statistics
- `jmap` - for heap dump generation
- `ps` - for process information

## Extending the Framework

To add support for additional languages:

1. Create a new diagnoser that implements the `Diagnoser` interface
2. Register it with the `Manager` using `Register(diagnoser)`
3. Implement the required methods:
   - `IdentifyProcess(pid int)` - Determine if process is of this language type
   - `GetProcessInfo(pid int)` - Retrieve basic process information
   - `Diagnose(pid int)` - Perform full diagnostic
   - `GetName()` - Return the language name

## Examples

### List all Java processes:
```bash
./crush java-diag
```

### Diagnose a specific Java application:
```bash
./crush java-diag --pid 12345
```

### Focus on memory issues:
```bash
./crush java-diag --pid 12345 --oom-only
```

### Focus on CPU issues:
```bash
./crush java-diag --pid 12345 --cpu-only
```

### Diagnose all Java applications:
```bash
./crush java-diag --all
```