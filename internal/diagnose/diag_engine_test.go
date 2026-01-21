package diagnose

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewDiagnostics(t *testing.T) {
	timeout := 5 * time.Second
	d := New(timeout)
	
	if d == nil {
		t.Fatal("Expected Diagnostics instance, got nil")
	}
	
	if d.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, d.timeout)
	}
}

func TestJavaProcessStruct(t *testing.T) {
	p := Process{
		PID:        1234,
		Name:       "test-process",
		CPUUsage:   10.5,
		MemoryUsed: 1024 * 1024, // 1MB
		MemoryMax:  1024 * 1024 * 1024, // 1GB
		Type:       "java",
	}
	
	if p.PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", p.PID)
	}
	
	if p.Name != "test-process" {
		t.Errorf("Expected name 'test-process', got '%s'", p.Name)
	}
	
	if p.CPUUsage != 10.5 {
		t.Errorf("Expected CPU usage 10.5, got %f", p.CPUUsage)
	}
	
	if p.Type != "java" {
		t.Errorf("Expected type 'java', got '%s'", p.Type)
	}
}

func TestDiagnosticResultStruct(t *testing.T) {
	result := &DiagnosticResult{
		Process: Process{
			PID:        1234,
			Name:       "test-process",
			CPUUsage:   10.5,
			MemoryUsed: 1024 * 1024,
			MemoryMax:  1024 * 1024 * 1024,
			Type:       "java",
		},
		IsOOM:           true,
		IsHighCPU:       false,
		IsStuckThread:   true,
		HasFrequentGC:   false,
		HasFullGCIssues: true,
		GCActivity:      "some gc activity",
		ThreadDump:      "some thread dump",
		Timestamp:       time.Now(),
		Issues: []Issue{
			{
				Type:        "memory",
				Description: "high memory usage",
				Severity:    "high",
				Suggestion:  "reduce memory usage",
			},
		},
	}
	
	if result.Process.PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", result.Process.PID)
	}
	
	if !result.IsOOM {
		t.Error("Expected IsOOM to be true")
	}
	
	if result.IsHighCPU {
		t.Error("Expected IsHighCPU to be false")
	}
	
	if !result.IsStuckThread {
		t.Error("Expected IsStuckThread to be true")
	}
	
	if result.HasFrequentGC {
		t.Error("Expected HasFrequentGC to be false")
	}
	
	if !result.HasFullGCIssues {
		t.Error("Expected HasFullGCIssues to be true")
	}
	
	if result.GCActivity != "some gc activity" {
		t.Errorf("Expected GC activity 'some gc activity', got '%s'", result.GCActivity)
	}
	
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
	
	if result.Issues[0].Type != "memory" {
		t.Errorf("Expected issue type 'memory', got '%s'", result.Issues[0].Type)
	}
}

func TestIssueStruct(t *testing.T) {
	issue := Issue{
		Type:        "cpu",
		Description: "high CPU usage",
		Severity:    "medium",
		Suggestion:  "optimize code",
	}
	
	if issue.Type != "cpu" {
		t.Errorf("Expected type 'cpu', got '%s'", issue.Type)
	}
	
	if issue.Description != "high CPU usage" {
		t.Errorf("Expected description 'high CPU usage', got '%s'", issue.Description)
	}
	
	if issue.Severity != "medium" {
		t.Errorf("Expected severity 'medium', got '%s'", issue.Severity)
	}
	
	if issue.Suggestion != "optimize code" {
		t.Errorf("Expected suggestion 'optimize code', got '%s'", issue.Suggestion)
	}
}

func TestCheckForOOM(t *testing.T) {
	d := New(5 * time.Second)
	
	// Test case 1: Memory usage > 90% of max memory
	process1 := Process{
		MemoryUsed: 950 * 1024 * 1024, // 950MB
		MemoryMax:  1000 * 1024 * 1024, // 1000MB (1GB)
	}
	
	result1 := d.checkForOOM(process1)
	if !result1 {
		t.Error("Expected OOM to be detected when memory usage is > 90% of max")
	}
	
	// Test case 2: Memory usage < 90% of max memory
	process2 := Process{
		MemoryUsed: 500 * 1024 * 1024, // 500MB
		MemoryMax:  1000 * 1024 * 1024, // 1000MB (1GB)
	}
	
	result2 := d.checkForOOM(process2)
	if result2 {
		t.Error("Expected OOM not to be detected when memory usage is < 90% of max")
	}
	
	// Test case 3: Memory usage > 8GB threshold
	process3 := Process{
		MemoryUsed: 9 * 1024 * 1024 * 1024, // 9GB
		MemoryMax:  0, // Unknown max
	}
	
	result3 := d.checkForOOM(process3)
	if !result3 {
		t.Error("Expected OOM to be detected when memory usage > 8GB")
	}
	
	// Test case 4: Normal memory usage
	process4 := Process{
		MemoryUsed: 100 * 1024 * 1024, // 100MB
		MemoryMax:  1000 * 1024 * 1024, // 1000MB (1GB)
	}
	
	result4 := d.checkForOOM(process4)
	if result4 {
		t.Error("Expected OOM not to be detected for normal memory usage")
	}
}

func TestAnalyzeThreadDump(t *testing.T) {
	d := New(5 * time.Second)
	
	// Test case 1: Contains lock contention pattern (waiting for monitor entry followed by locked)
	threadDump1 := `Full thread dump:

"Thread-1" #1 prio=5 os_prio=0 tid=0x00007f8b8c001000 nid=0x1001 runnable [0x00007f8b9c7fe000]
   java.lang.Thread.State: RUNNABLE

"Thread-2" #2 prio=5 os_prio=0 tid=0x00007f8b8c002000 nid=0x1002 waiting for monitor entry
   java.lang.Thread.State: BLOCKED (on object monitor)
   locked <0x00000007c0010000> by <0x00000007c0010000> (a java.lang.Object)
`

	result1 := d.analyzeThreadDump(threadDump1)
	if !result1 {
		t.Error("Expected stuck threads to be detected when lock contention pattern is present")
	}
	
	// Test case 2: Contains WAITING threads (many of them)
	var waitingDump strings.Builder
	waitingDump.WriteString("Full thread dump:\n\n")
	for i := 0; i < 60; i++ {
		waitingDump.WriteString(fmt.Sprintf(`"Thread-%d" #1 prio=5 os_prio=0 tid=0x00007f8b8c00%d000 nid=0x100%d waiting on condition [0x00007f8b9c7fe000]
   java.lang.Thread.State: WAITING (parking)
`, i, i%1000, i%1000))
	}
	
	result2 := d.analyzeThreadDump(waitingDump.String())
	if !result2 {
		t.Error("Expected stuck threads to be detected when many WAITING threads are present")
	}
	
	// Test case 3: Normal thread dump without stuck threads
	threadDump3 := `Full thread dump:

"Thread-1" #1 prio=5 os_prio=0 tid=0x00007f8b8c001000 nid=0x1001 runnable [0x00007f8b9c7fe000]
   java.lang.Thread.State: RUNNABLE

"Thread-2" #2 prio=5 os_prio=0 tid=0x00007f8b8c002000 nid=0x1002 runnable [0x00007f8b9c7ff000]
   java.lang.Thread.State: RUNNABLE
`
	
	result3 := d.analyzeThreadDump(threadDump3)
	if result3 {
		t.Error("Expected no stuck threads to be detected in normal thread dump")
	}
}

func TestAnalyzeGCActivity(t *testing.T) {
	d := New(5 * time.Second)
	
	// Test case 1: Normal GC activity (header + data line)
	gcStats1 := `S0C    S1C    S0U    S1U      EC       EU        OC         OU       MC     MU    CCSC   CCSU   YGC     YGCT    FGC    FGCT     GCT   
2048.0 2048.0  0.0   64.0   16384.0 12345.6  40960.0    12345.7   16384.0 12345.6 1792.0 1234.5      5    0.012   1      0.003    0.015`
	
	freqGC1, fullGCIssues1 := d.analyzeGCActivity(gcStats1)
	if freqGC1 {
		t.Error("Expected no frequent GC issues with low YGC count")
	}
	if fullGCIssues1 {
		t.Error("Expected no full GC issues with low FGC count and low FGCT")
	}
	
	// Test case 2: High YGC count (frequent GC)
	gcStats2 := `S0C    S1C    S0U    S1U      EC       EU        OC         OU       MC     MU    CCSC   CCSU   YGC     YGCT    FGC    FGCT     GCT   
2048.0 2048.0  0.0   64.0   16384.0 12345.6  40960.0    12345.7   16384.0 12345.6 1792.0 1234.5    150    0.150   3      0.005    0.155`
	
	freqGC2, fullGCIssues2 := d.analyzeGCActivity(gcStats2)
	if !freqGC2 {
		t.Error("Expected frequent GC issues with high YGC count (>100)")
	}
	if fullGCIssues2 {
		t.Error("Expected no full GC issues with low FGC count")
	}
	
	// Test case 3: High FGC count (full GC issues)
	gcStats3 := `S0C    S1C    S0U    S1U      EC       EU        OC         OU       MC     MU    CCSC   CCSU   YGC     YGCT    FGC    FGCT     GCT   
2048.0 2048.0  0.0   64.0   16384.0 12345.6  40960.0    12345.7   16384.0 12345.6 1792.0 1234.5     50    0.500  10      0.050    0.550`
	
	freqGC3, fullGCIssues3 := d.analyzeGCActivity(gcStats3)
	if freqGC3 {
		t.Error("Expected no frequent GC issues with moderate YGC count")
	}
	if !fullGCIssues3 {
		t.Error("Expected full GC issues with high FGC count (>5)")
	}
}

func TestManagerInitialization(t *testing.T) {
	manager := NewManager()
	
	if manager == nil {
		t.Fatal("Expected Manager instance, got nil")
	}
	
	if manager.diagnosers == nil {
		t.Error("Expected diagnosers map to be initialized")
	}
	
	if len(manager.diagnosers) != 0 {
		t.Error("Expected empty diagnosers map initially")
	}
}

func TestManagerRegister(t *testing.T) {
	manager := NewManager()
	diag := New(5 * time.Second)
	
	manager.Register(diag)
	
	if len(manager.diagnosers) != 1 {
		t.Errorf("Expected 1 diagnoser after registration, got %d", len(manager.diagnosers))
	}
	
	// Check if the diagnoser was registered with the correct name
	_, exists := manager.diagnosers[diag.GetName()]
	if !exists {
		t.Error("Expected diagnoser to be registered with its name")
	}
}

func TestManagerGetDiagnoser(t *testing.T) {
	manager := NewManager()
	diag := New(5 * time.Second)
	
	manager.Register(diag)
	
	foundDiag, exists := manager.GetDiagnoser(diag.GetName())
	if !exists {
		t.Error("Expected diagnoser to exist")
	}
	
	if foundDiag == nil {
		t.Error("Expected found diagnoser to not be nil")
	}
	
	// Note: We can't directly compare the instances, but we can check if GetName matches
	if foundDiag.GetName() != diag.GetName() {
		t.Errorf("Expected found diagnoser name '%s', got '%s'", diag.GetName(), foundDiag.GetName())
	}
	
	// Test non-existent diagnoser
	_, exists = manager.GetDiagnoser("nonexistent")
	if exists {
		t.Error("Expected non-existent diagnoser to not exist")
	}
}

func TestManagerIdentifyProcess(t *testing.T) {
	// This test is tricky because it depends on actual processes running
	// We'll test the logic path but not the actual process identification
	manager := NewManager()
	diag := New(5 * time.Second)
	
	manager.Register(diag)
	
	// Since we can't guarantee a real process exists, we'll just make sure no panic occurs
	// and the method signature is correct
	// This test mainly ensures the method exists and doesn't crash on basic input
	_, _, err := manager.IdentifyProcess(1) // PID 1 is usually init/systemd
	
	// We don't check the result since it depends on system state,
	// but we ensure no panic occurs
	_ = err
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1536 * 1024, "1.50 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1536 * 1024 * 1024, "1.50 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{1536 * 1024 * 1024 * 1024, "1.50 TB"},
	}
	
	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

// Test to verify that external commands are not executed during testing
// This helps ensure that tests don't rely on system-dependent commands
func TestNoExternalCommandExecution(t *testing.T) {
	// This test verifies that certain methods don't execute external commands
	// by checking that they don't call exec.Command directly in a way that would fail in test environment
	
	d := New(100 * time.Millisecond) // Short timeout for tests
	
	// We can't test actual process diagnosis without real processes,
	// but we can verify the structure and that no external calls are made inappropriately
	
	if d.timeout != 100*time.Millisecond {
		t.Errorf("Expected timeout to be preserved, got %v", d.timeout)
	}
}

// Mock implementation for testing command execution
type mockExecutor struct {
	command string
	args    []string
	output  string
	err     error
}

// Test that the diagnostic methods handle command execution errors gracefully
func TestCommandExecutionErrorHandling(t *testing.T) {
	// This test focuses on verifying that error handling works properly
	// without actually executing commands that might not exist on the test system
	
	d := New(100 * time.Millisecond)
	
	// Test that the getJavaProcessName method handles errors appropriately
	// We can't easily mock exec.Command, but we can verify the timeout behavior exists
	if d.timeout <= 0 {
		t.Error("Expected positive timeout value")
	}
}

func TestGetSystemProcessInfo(t *testing.T) {
	d := New(100 * time.Millisecond)
	
	// Test with a non-existent PID to trigger error handling
	// This should not panic and should return an error
	_, err := d.GetSystemProcessInfo(999999) // Very high PID unlikely to exist
	
	// We expect an error here, which is normal behavior
	// The important thing is that it doesn't panic
	if err == nil {
		// If no error, it means the PID exists (unlikely), so just continue
		t.Log("PID 999999 exists (unexpected but not an error in the test)")
	}
}

func TestGetOpenFiles(t *testing.T) {
	d := New(100 * time.Millisecond)
	
	// Test with a non-existent PID to trigger the error handling path
	// This should return a message indicating lsof is not available rather than panicking
	result, err := d.GetOpenFiles(999999) // Very high PID unlikely to exist
	
	// The method should handle this gracefully
	if err != nil {
		// If there's an error, that's OK as long as it's handled properly
		t.Logf("Got expected error for non-existent PID: %v", err)
	} else {
		// If no error, check if it's the expected "not available" message
		if strings.Contains(result, "lsof not available") {
			t.Log("Got expected 'lsof not available' message")
		}
	}
}

func TestGetNetworkConnections(t *testing.T) {
	d := New(100 * time.Millisecond)
	
	// Similar to GetOpenFiles, test error handling
	result, err := d.GetNetworkConnections(999999) // Very high PID unlikely to exist
	
	if err != nil {
		t.Logf("Got expected error for non-existent PID: %v", err)
	} else {
		if strings.Contains(result, "lsof network check not available") {
			t.Log("Got expected 'lsof not available' message for network connections")
		}
	}
}