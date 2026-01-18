package diagnose

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/crush/internal/app"
)

// AppConnector provides a bridge between the diagnose package and the app package
// for AI-enhanced diagnostics
type AppConnector struct {
	appInstance *app.App
	aiEnhancer  *AIEnhancer
}

// NewAppConnector creates a new connector between diagnose and app packages
func NewAppConnector(appInstance *app.App) *AppConnector {
	return &AppConnector{
		appInstance: appInstance,
		aiEnhancer:  NewAIEnhancer(30 * time.Second),
	}
}

// RunAIDiagnostic runs a diagnostic with AI enhancement
func (ac *AppConnector) RunAIDiagnostic(ctx context.Context, pid int) (*EnhancedDiagnosticResult, error) {
	// First, run the basic Java diagnostic
	javaDiag := New(10 * time.Second)
	basicResult, err := javaDiag.DiagnoseProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to run basic diagnostic: %w", err)
	}

	// Then enhance with AI analysis
	enhancedResult, err := ac.enhanceWithAI(ctx, basicResult)
	if err != nil {
		return nil, fmt.Errorf("failed to enhance diagnostic with AI: %w", err)
	}

	return enhancedResult, nil
}

// enhanceWithAI enhances a diagnostic result with AI analysis
func (ac *AppConnector) enhanceWithAI(ctx context.Context, result *DiagnosticResult) (*EnhancedDiagnosticResult, error) {
	// Define the AI service call function that uses the app instance
	aiServiceCall := func(ctx context.Context, prompt string) (string, error) {
		return ac.appInstance.AskLLM(ctx, prompt)
	}

	// Use the AI enhancer to analyze the result
	enhancedResult, err := ac.aiEnhancer.AnalyzeWithAI(ctx, result, aiServiceCall)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze with AI: %w", err)
	}

	return enhancedResult, nil
}

// RunAIDiagnosticForAll runs AI diagnostics on all available processes
func (ac *AppConnector) RunAIDiagnosticForAll(ctx context.Context) ([]*EnhancedDiagnosticResult, error) {
	manager := NewManager()
	javaDiag := New(10 * time.Second)
	manager.Register(javaDiag)

	// Get all processes
	processes, err := manager.GetAllProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var results []*EnhancedDiagnosticResult
	for _, process := range processes {
		result, err := ac.RunAIDiagnostic(ctx, process.PID)
		if err != nil {
			// Log error but continue with other processes
			continue
		}
		results = append(results, result)
	}

	return results, nil
}