package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/crush/internal/message"
)

// RunDiagnosticWithAI runs a diagnostic and enhances it with AI analysis
func (app *App) RunDiagnosticWithAI(ctx context.Context, pid int) (string, error) {
	// This would integrate with the diagnose package to run diagnostics
	// and then enhance with AI analysis

	// For now, this is a placeholder that would connect the diagnostic
	// results with the AI enhancement
	return "", nil
}

// AskLLM sends a prompt to the LLM service and returns the response
func (app *App) AskLLM(ctx context.Context, prompt string) (string, error) {
	if app.AgentCoordinator == nil {
		return "", fmt.Errorf("agent coordinator not initialized")
	}

	// Create a temporary session for the LLM request
	session, err := app.Sessions.Create(ctx, "ai-diagnostic-request")
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Run the agent coordinator with the diagnostic prompt
	result, err := app.AgentCoordinator.Run(ctx, session.ID, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to run agent: %w", err)
	}

	if result == nil {
		return "", fmt.Errorf("no response from agent")
	}

	// Get the messages for this session to extract the response
	messages, err := app.Messages.List(ctx, session.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get messages: %w", err)
	}

	// Extract the response from the messages
	response := ""
	for _, msg := range messages {
		if msg.Role == message.Assistant {
			response += msg.Content().String()
		}
	}

	if response == "" {
		return "", fmt.Errorf("no assistant response found")
	}

	return response, nil
}