package diagnose

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AIEnhancer provides AI-powered analysis of diagnostic results
type AIEnhancer struct {
	timeout time.Duration
}

// NewAIEnhancer creates a new AI enhancer instance
func NewAIEnhancer(timeout time.Duration) *AIEnhancer {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &AIEnhancer{
		timeout: timeout,
	}
}

// EnhancedDiagnosticResult extends the basic diagnostic result with AI insights
type EnhancedDiagnosticResult struct {
	BasicResult    *DiagnosticResult
	AIInsights     []AIInsight
	RootCause      string
	ActionableSteps []string
	Confidence     float64 // 0.0 to 1.0
}

// AIInsight represents an insight provided by the AI analysis
type AIInsight struct {
	Category    string  // e.g., "performance", "memory", "threading", "gc"
	Description string
	Severity    string  // "low", "medium", "high", "critical"
	Confidence  float64 // 0.0 to 1.0
}

// AnalyzeWithAI performs AI-enhanced analysis of diagnostic data
func (ae *AIEnhancer) AnalyzeWithAI(ctx context.Context, result *DiagnosticResult, aiServiceCall func(context.Context, string) (string, error)) (*EnhancedDiagnosticResult, error) {
	// Prepare diagnostic data for AI analysis
	diagnosticData := ae.prepareDiagnosticData(result)

	// Create a prompt for the AI
	prompt := ae.createAIPrompt(diagnosticData)

	// Call the AI service through the provided callback
	response, err := aiServiceCall(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	// Parse the AI response
	enhancedResult, err := ae.parseAIResponse(response, result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return enhancedResult, nil
}

// prepareDiagnosticData formats diagnostic data for AI analysis
func (ae *AIEnhancer) prepareDiagnosticData(result *DiagnosticResult) map[string]interface{} {
	data := map[string]interface{}{
		"process_info": map[string]interface{}{
			"pid":         result.Process.PID,
			"name":        result.Process.Name,
			"type":        result.Process.Type,
			"cpu_usage":   result.Process.CPUUsage,
			"memory_used": result.Process.MemoryUsed,
			"memory_max":  result.Process.MemoryMax,
		},
		"issues_found": len(result.Issues),
		"timestamp":    result.Timestamp,
		"has_oom":      result.IsOOM,
		"has_high_cpu": result.IsHighCPU,
		"has_stuck_threads": result.IsStuckThread,
		"has_frequent_gc": result.HasFrequentGC,
		"has_full_gc_issues": result.HasFullGCIssues,
		"gc_activity":  result.GCActivity,
		"thread_dump":  result.ThreadDump,
	}
	
	// Add issue details
	issues := make([]map[string]string, len(result.Issues))
	for i, issue := range result.Issues {
		issues[i] = map[string]string{
			"type":        issue.Type,
			"description": issue.Description,
			"severity":    issue.Severity,
			"suggestion":  issue.Suggestion,
		}
	}
	data["issues"] = issues
	
	return data
}

// createAIPrompt creates a prompt for the AI service
func (ae *AIEnhancer) createAIPrompt(diagnosticData map[string]interface{}) string {
	jsonData, _ := json.Marshal(diagnosticData)
	
	prompt := fmt.Sprintf(`You are a senior application performance engineer. Analyze the following diagnostic data from a running application and provide insights:

Diagnostic Data:
%s

Please provide:
1. Root cause analysis of the issues
2. Actionable steps to resolve the problems
3. Additional insights about performance, memory, threading, or garbage collection patterns
4. Confidence level in your assessment (0-100%%)

Format your response as a JSON object with these fields:
- "root_cause": string
- "actionable_steps": array of strings
- "ai_insights": array of objects with "category", "description", "severity", and "confidence" fields
- "confidence": number between 0 and 1

Focus on identifying patterns, correlations between issues, and providing specific, actionable recommendations.`, string(jsonData))
	
	return prompt
}

// parseAIResponse parses the AI response into structured data
func (ae *AIEnhancer) parseAIResponse(response string, basicResult *DiagnosticResult) (*EnhancedDiagnosticResult, error) {
	// Try to extract JSON from the response (in case the AI wraps it in text)
	jsonStr := ae.extractJSONFromResponse(response)
	
	var aiData struct {
		RootCause     string      `json:"root_cause"`
		ActionableSteps []string  `json:"actionable_steps"`
		AIInsights    []struct {
			Category    string  `json:"category"`
			Description string  `json:"description"`
			Severity    string  `json:"severity"`
			Confidence  float64 `json:"confidence"`
		} `json:"ai_insights"`
		Confidence float64 `json:"confidence"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &aiData); err != nil {
		return nil, fmt.Errorf("failed to parse AI response JSON: %w", err)
	}
	
	// Convert AI insights to our internal format
	insights := make([]AIInsight, len(aiData.AIInsights))
	for i, insight := range aiData.AIInsights {
		insights[i] = AIInsight{
			Category:    insight.Category,
			Description: insight.Description,
			Severity:    insight.Severity,
			Confidence:  insight.Confidence,
		}
	}
	
	enhancedResult := &EnhancedDiagnosticResult{
		BasicResult:     basicResult,
		AIInsights:      insights,
		RootCause:       aiData.RootCause,
		ActionableSteps: aiData.ActionableSteps,
		Confidence:      aiData.Confidence,
	}
	
	return enhancedResult, nil
}

// extractJSONFromResponse extracts JSON from AI response that might contain wrapper text
func (ae *AIEnhancer) extractJSONFromResponse(response string) string {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	
	if start != -1 && end != -1 && end > start {
		return response[start : end+1]
	}
	
	// If no JSON delimiters found, return the whole response
	// This is a fallback - ideally the AI should return clean JSON
	return response
}
