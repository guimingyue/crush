package diagnose

import (
	"os"
	"path/filepath"
	"strings"
)

// JavaToolsConfig holds the configuration for Java diagnostic tools
type JavaToolsConfig struct {
	JStatPath  string // Path to jstat command
	JMapPath   string // Path to jmap command
	JStackPath string // Path to jstack command
	JPSPath    string // Path to jps command
	JInfoPath  string // Path to jinfo command
}

// GetDefaultJavaToolsConfig returns a configuration with default paths
func GetDefaultJavaToolsConfig() *JavaToolsConfig {
	config := &JavaToolsConfig{}

	// Try to find tools in common locations
	javaHome := getJavaHome()
	if javaHome != "" {
		binDir := filepath.Join(javaHome, "bin")
		config.JStatPath = filepath.Join(binDir, "jstat")
		config.JMapPath = filepath.Join(binDir, "jmap")
		config.JStackPath = filepath.Join(binDir, "jstack")
		config.JPSPath = filepath.Join(binDir, "jps")
		config.JInfoPath = filepath.Join(binDir, "jinfo")
	}

	// If not found in JAVA_HOME, try common system locations
	if !fileExists(config.JStatPath) {
		config.JStatPath = findExecutable("jstat")
	}
	if !fileExists(config.JStatPath) {
		config.JStatPath = "jstat" // fallback to PATH
	}

	if !fileExists(config.JMapPath) {
		config.JMapPath = findExecutable("jmap")
	}
	if !fileExists(config.JMapPath) {
		config.JMapPath = "jmap" // fallback to PATH
	}

	if !fileExists(config.JStackPath) {
		config.JStackPath = findExecutable("jstack")
	}
	if !fileExists(config.JStackPath) {
		config.JStackPath = "jstack" // fallback to PATH
	}

	if !fileExists(config.JPSPath) {
		config.JPSPath = findExecutable("jps")
	}
	if !fileExists(config.JPSPath) {
		config.JPSPath = "jps" // fallback to PATH
	}

	if !fileExists(config.JInfoPath) {
		config.JInfoPath = findExecutable("jinfo")
	}
	if !fileExists(config.JInfoPath) {
		config.JInfoPath = "jinfo" // fallback to PATH
	}

	return config
}

// getJavaHome attempts to find the JAVA_HOME environment variable
func getJavaHome() string {
	// Check JAVA_HOME environment variable
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		return javaHome
	}

	// Check for java command and try to infer JAVA_HOME
	javaPath := findExecutable("java")
	if javaPath != "" {
		// If java is found, try to find its parent directories
		// java is typically in JAVA_HOME/bin/java
		dir := filepath.Dir(filepath.Dir(javaPath))
		if isValidJavaHome(dir) {
			return dir
		}
	}

	return ""
}

// isValidJavaHome checks if a directory is a valid JAVA_HOME
func isValidJavaHome(path string) bool {
	// Check if the directory contains the expected structure
	requiredPaths := []string{
		filepath.Join(path, "bin", "java"),
		filepath.Join(path, "lib"),
	}

	for _, reqPath := range requiredPaths {
		if !fileExists(reqPath) {
			return false
		}
	}

	return true
}

// findExecutable searches for an executable in common Java locations
func findExecutable(toolName string) string {
	// Common Java installation paths
	possiblePaths := []string{
		"/usr/bin/" + toolName,
		"/usr/local/bin/" + toolName,
		"/opt/java/bin/" + toolName,
		"/usr/lib/jvm/default-java/bin/" + toolName,
		"/Library/Java/JavaVirtualMachines/*/Contents/Home/bin/" + toolName, // macOS
	}

	for _, path := range possiblePaths {
		// Handle wildcard for macOS
		if strings.Contains(path, "*") {
			matches, _ := filepath.Glob(path)
			for _, match := range matches {
				if fileExists(match) {
					return match
				}
			}
		} else {
			if fileExists(path) {
				return path
			}
		}
	}

	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// Validate checks if all required Java tools are accessible
func (c *JavaToolsConfig) Validate() []string {
	var missingTools []string

	if c.JStatPath == "" || !fileExists(c.JStatPath) {
		missingTools = append(missingTools, "jstat")
	}
	if c.JMapPath == "" || !fileExists(c.JMapPath) {
		missingTools = append(missingTools, "jmap")
	}
	if c.JStackPath == "" || !fileExists(c.JStackPath) {
		missingTools = append(missingTools, "jstack")
	}
	if c.JPSPath == "" || !fileExists(c.JPSPath) {
		missingTools = append(missingTools, "jps")
	}

	return missingTools
}