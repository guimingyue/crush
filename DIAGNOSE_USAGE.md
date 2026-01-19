# AI Diagnostic Tool Usage Guide

The AI Diagnostic tool is an intelligent system integrated into the crush application that uses AI to identify and help resolve online application issues including OutOfMemoryError, high CPU usage, and other performance problems.

## Overview

The AI diagnostic tool leverages the LLM (Large Language Model) to analyze issue descriptions and automatically execute appropriate diagnostic tools. It follows the vibe coding pattern where users describe their issues and the AI determines and executes the appropriate diagnostic steps.

## Installation

The AI diagnostic tool is integrated into the `crush` CLI. Simply build the project:

```bash
cd /path/to/crush
go build -o crush .
```

## Usage

### Interactive Mode (Recommended)

The AI diagnostic tool is available through the interactive mode of crush:

```bash
# Start crush in interactive mode
./crush

# In the interactive mode, describe your issue:
# "My application is running slowly and seems to be using too much memory"
# The AI will analyze the issue and execute appropriate diagnostic tools
```

### Programmatic Access

The AI diagnostic functionality is also available programmatically through the app:

```go
app := // your app instance
ctx := context.Background()

// Describe the issue to the AI diagnostic tool
err := app.RunAIDiagnostic(ctx, "Application is experiencing high memory usage and slow performance")
if err != nil {
    log.Printf("Error running AI diagnostic: %v", err)
}
```

## Supported Diagnostics

The AI diagnostic tool can automatically detect and analyze:

### Java Application Issues
- **Memory Issues**: OutOfMemoryError, high memory consumption
- **CPU Issues**: High CPU usage, performance bottlenecks
- **Thread Issues**: Stuck threads, deadlocks, thread contention
- **Garbage Collection Issues**: Frequent GC, full GC problems

### Multi-Language Support
- **Java**: Full diagnostic support
- **Go, C++, Rust, etc.**: Planned expansion through the extensible framework

## How It Works

1. **Issue Description**: User describes the problem in natural language
2. **AI Analysis**: The LLM analyzes the description to determine likely causes
3. **Tool Selection**: AI selects appropriate diagnostic tools to run
4. **Execution**: Diagnostic tools are executed automatically
5. **Analysis**: Results are analyzed by the AI
6. **Recommendations**: AI provides specific recommendations to resolve the issue

## Integration with Vibe Coding

The AI diagnostic tool follows the vibe coding pattern:
- Users describe their issue in natural language
- The AI determines what diagnostic tools to run
- Tools are executed automatically
- Results are analyzed and presented to the user
- Specific recommendations are provided

## Extending the Framework

To add support for additional languages:

1. Create a new diagnoser that implements the `Diagnoser` interface in `internal/diagnose`
2. Register it with the `Manager` using `Register(diagnoser)`
3. Implement the required methods:
   - `IdentifyProcess(pid int)` - Determine if process is of this language type
   - `GetProcessInfo(pid int)` - Retrieve basic process information
   - `Diagnose(pid int)` - Perform full diagnostic
   - `GetName()` - Return the language name

The AI will automatically incorporate new diagnosers into its analysis.

## Examples

### In Interactive Mode:
```
> My Java application is running slowly and consuming too much memory
[AI analyzes the issue and runs appropriate diagnostic tools...]
[Results and recommendations are displayed]
```

### Programmatic Usage:
```go
// Run AI-guided diagnostics
err := app.RunAIDiagnostic(ctx, "Application is experiencing high memory usage")
if err != nil {
    // Handle error
}
```

## Troubleshooting

### Common Issues

1. **AI doesn't recognize the issue**: Provide more specific details about the problem
2. **Diagnostic tools fail**: Ensure required system tools are available (jstack, jstat, etc.)
3. **No processes found**: Verify that target applications are running

### Required Tools

Depending on the target applications, the following tools may be required:
- Java applications: `jstack`, `jstat`, `jmap`, `ps`
- Other languages: Language-specific diagnostic tools