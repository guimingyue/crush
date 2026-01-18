package diagnose

import (
	"fmt"
	"sync"
)

// Manager manages multiple diagnosers for different languages/technologies
type Manager struct {
	diagnosers map[string]Diagnoser
	mutex      sync.RWMutex
}

// NewManager creates a new diagnostic manager
func NewManager() *Manager {
	return &Manager{
		diagnosers: make(map[string]Diagnoser),
	}
}

// Register adds a new diagnoser to the manager
func (m *Manager) Register(diagnoser Diagnoser) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	name := diagnoser.GetName()
	m.diagnosers[name] = diagnoser
}

// GetDiagnoser retrieves a diagnoser by name
func (m *Manager) GetDiagnoser(name string) (Diagnoser, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	diagnoser, exists := m.diagnosers[name]
	return diagnoser, exists
}

// IdentifyProcess identifies which diagnoser should handle a specific process
func (m *Manager) IdentifyProcess(pid int) (string, Diagnoser, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for name, diagnoser := range m.diagnosers {
		isMatch, err := diagnoser.IdentifyProcess(pid)
		if err != nil {
			// Don't return error if one diagnoser fails, try others
			continue
		}
		if isMatch {
			return name, diagnoser, nil
		}
	}
	
	return "", nil, fmt.Errorf("no diagnoser found for process %d", pid)
}

// DiagnoseProcess finds the appropriate diagnoser and runs diagnostics
func (m *Manager) DiagnoseProcess(pid int) (*DiagnosticResult, error) {
	_, diagnoser, err := m.IdentifyProcess(pid)
	if err != nil {
		return nil, err
	}
	
	return diagnoser.Diagnose(pid)
}

// GetAllProcesses finds all processes that can be diagnosed by any registered diagnoser
func (m *Manager) GetAllProcesses() ([]Process, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var allProcesses []Process
	
	for _, diagnoser := range m.diagnosers {
		// For now, we'll just handle Java processes since that's the only one implemented
		// We'll need to use type assertion to access the specific methods
		if javaDiagnoser, ok := diagnoser.(*Diagnostics); ok {
			javaProcesses, err := javaDiagnoser.FindJavaProcesses()
			if err != nil {
				// Continue with other diagnosers if one fails
				continue
			}

			for _, jp := range javaProcesses {
				allProcesses = append(allProcesses, jp.Process)
			}
		}
	}
	
	return allProcesses, nil
}

// GetSupportedLanguages returns a list of supported languages/technologies
func (m *Manager) GetSupportedLanguages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	languages := make([]string, 0, len(m.diagnosers))
	for name := range m.diagnosers {
		languages = append(languages, name)
	}
	
	return languages
}