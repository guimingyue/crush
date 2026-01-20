# App Diagnostics Tool Usage Guide

The App Diagnostics tool is an intelligent agent tool integrated into the crush application that uses AI to identify and help resolve online application issues including OutOfMemoryError, high CPU usage, and other performance problems.

## Overview

The App Diagnostics tool is an agent tool that leverages the LLM (Large Language Model) to analyze issue descriptions and automatically execute appropriate diagnostic tools. It follows the vibe coding pattern where users describe their issues and the AI determines and executes the appropriate diagnostic steps.

## Installation

The App Diagnostics tool is integrated into the `crush` CLI. Simply build the project:

```bash
cd /path/to/crush
go build -o crush .
```

## Usage

### Interactive Mode (Recommended)

The App Diagnostics tool is available through the interactive mode of crush:

```bash
# Start crush in interactive mode
./crush

# In the interactive mode, describe your issue:
# "My application is running slowly and seems to be using too much memory"
# The AI will automatically use the app_diagnostics tool to analyze the issue
```

### Tool Direct Usage

The tool can be called directly by the AI agent when it determines that application diagnostics are needed:

```json
{
  "name": "app_diagnostics",
  "arguments": {
    "process_id": 1234,
    "type": "java"
  }
}
```

## Supported Diagnostics

The App Diagnostics tool can automatically detect and analyze:

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
3. **Tool Selection**: AI selects the `app_diagnostics` tool to run
4. **Execution**: Diagnostic tools are executed automatically
5. **Analysis**: Results are analyzed by the AI
6. **Recommendations**: AI provides specific recommendations to resolve the issue

## Integration with Vibe Coding

The App Diagnostics tool follows the vibe coding pattern:
- Users describe their issue in natural language
- The AI determines that diagnostic tools are needed
- The `app_diagnostics` tool is executed automatically
- Results are analyzed and presented to the user
- Specific recommendations are provided

## Java Tools Requirements

The diagnostic tool requires Java diagnostic commands to be available in the system PATH:

### Required Java Diagnostic Tools
- `jstat` - JVM statistics monitoring tool
- `jmap` - Memory mapping tool
- `jstack` - Stack trace tool
- `jps` - JVM process status tool
- `jinfo` - Configuration information tool

### Configuration
The diagnostic tool expects Java tools to be available in the system PATH:
1. Ensure JDK (not just JRE) is installed
2. Java tools should be in the system PATH
3. Common locations: `/usr/bin/`, `$JAVA_HOME/bin/`

### Verification
Verify that Java diagnostic tools are accessible:
```bash
which jstat
which jmap
which jstack
java -version
echo $JAVA_HOME
```

## Examples

### In Interactive Mode:
```
> My Java application is running slowly and consuming too much memory
[AI analyzes the issue and runs app_diagnostics tool...]
[Results and recommendations are displayed]
```

### Tool Output Example:
```
Found 2 processes:
- PID: 1234, Name: /usr/bin/java MyApp, Type: java, CPU: 85.20%, Memory: 2.15 GB/4.00 GB
  Diagnostics:
    Process: /usr/bin/java MyApp (PID: 1234)
    Type: java
    CPU Usage: 85.20%
    Memory Used: 2.15 GB
    Memory Max: 4.00 GB
    Potential OOM: false
    High CPU Usage: true
    Stuck Threads: false
    Frequent GC: true
    Full GC Issues: false
    Issues Found:
      - [high] HighCPU: Process is consuming excessive CPU resources
        Suggestion: Consider profiling the application to identify hotspots
```

## Troubleshooting

### Common Issues

1. **AI doesn't recognize the issue**: Provide more specific details about the problem
2. **Diagnostic tools fail**: Ensure required system tools are available (jstack, jstat, etc.)
3. **No processes found**: Verify that target applications are running
4. **Java tools not found**: Ensure JDK tools are in PATH

#### Java Tools Not Found
If Java diagnostic tools are not found:

1. **Check Java Installation**: Ensure JDK (not just JRE) is installed
2. **Add to PATH**: Ensure Java tools are in your system PATH
3. **Verify Tools**: Check that individual tools are accessible

#### Verification Commands:
```bash
# Check Java installation
java -version

# Check individual tools
which jstat jmap jstack jps jinfo

# If tools are not found, install JDK:
# Ubuntu/Debian: sudo apt-get install openjdk-11-jdk
# CentOS/RHEL: sudo yum install java-11-openjdk-devel
# macOS: brew install openjdk
```

### Other Languages
- Other languages: Language-specific diagnostic tools