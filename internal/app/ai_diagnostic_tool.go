package app

import (
	"context"
	"fmt"
)

// RunAIDiagnostic runs an AI-guided diagnostic based on issue description
// This method now serves as a wrapper to inform users that diagnostics are available as an agent tool
func (app *App) RunAIDiagnostic(ctx context.Context, issueDescription string) error {
	return fmt.Errorf("AI diagnostics are now available as an agent tool. Use the 'app_diagnostics' tool in interactive mode to diagnose application issues.")
}