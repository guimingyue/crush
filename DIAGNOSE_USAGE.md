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

## Java Tools Configuration

The diagnostic tool supports configurable paths for Java diagnostic commands to work with different Java installations:

### Automatic Detection
The tool automatically searches for Java diagnostic tools in:
1. `JAVA_HOME` environment variable
2. Standard system PATH
3. Common installation directories

### Manual Configuration
If tools are installed in a non-standard location, ensure `JAVA_HOME` is set correctly:
```bash
export JAVA_HOME=/path/to/your/java/installation
export PATH=$PATH:$JAVA_HOME/bin
```

### Supported Java Diagnostic Tools
- `jstat` - JVM statistics monitoring tool
- `jmap` - Memory mapping tool
- `jstack` - Stack trace tool
- `jps` - JVM process status tool
- `jinfo` - Configuration information tool

### Verification
Verify that Java diagnostic tools are accessible:
```bash
which jstat
which jmap
which jstack
java -version
echo $JAVA_HOME
```

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
4. **Java tools not found**: Configure JAVA_HOME or ensure JDK tools are in PATH

#### Java Tools Not Found
If Java diagnostic tools are not found:

1. **Check Java Installation**: Ensure JDK (not just JRE) is installed
2. **Set JAVA_HOME**: Point to your JDK installation directory
3. **Add to PATH**: Ensure Java tools are in your system PATH
4. **Verify Tools**: Check that individual tools are accessible

#### Verification Commands:
```bash
# Check Java installation
java -version

# Check JAVA_HOME
echo $JAVA_HOME

# Check individual tools
which jstat jmap jstack jps jinfo

# If tools are not found, install JDK:
# Ubuntu/Debian: sudo apt-get install openjdk-11-jdk
# CentOS/RHEL: sudo yum install java-11-openjdk-devel
# macOS: brew install openjdk
```

### Other Languages
- Other languages: Language-specific diagnostic tools