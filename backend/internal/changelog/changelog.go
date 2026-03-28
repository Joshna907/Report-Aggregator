package changelog

import (
	"fmt"
	"sync"
	"time"
)

// ChangeEntry represents a single change to the aggregated report
type ChangeEntry struct {
	ID            string    `json:"id"`
	ComponentName string    `json:"componentName"`
	ComponentVer  string    `json:"componentVersion"`
	Field         string    `json:"field"`         // "license", "copyright", "supplier", "version", etc.
	OldValue      string    `json:"oldValue"`
	NewValue      string    `json:"newValue"`
	ChangedBy     string    `json:"changedBy"`
	ChangedAt     time.Time `json:"changedAt"`
	Reason        string    `json:"reason,omitempty"` // User's explanation for the change
}

// Changelog tracks all changes to the aggregated report
type Changelog struct {
	mu      sync.RWMutex
	entries []ChangeEntry
	counter int
}

// New creates a new Changelog
func New() *Changelog {
	return &Changelog{}
}

// LogChange records a change
func (cl *Changelog) LogChange(componentName, componentVer, field, oldValue, newValue, changedBy, reason string) ChangeEntry {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.counter++
	entry := ChangeEntry{
		ID:            fmt.Sprintf("change-%d", cl.counter),
		ComponentName: componentName,
		ComponentVer:  componentVer,
		Field:         field,
		OldValue:      oldValue,
		NewValue:      newValue,
		ChangedBy:     changedBy,
		ChangedAt:     time.Now(),
		Reason:        reason,
	}
	cl.entries = append(cl.entries, entry)
	return entry
}

// GetAll returns all change entries
func (cl *Changelog) GetAll() []ChangeEntry {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	result := make([]ChangeEntry, len(cl.entries))
	copy(result, cl.entries)
	return result
}

// GetByComponent returns changes for a specific component
func (cl *Changelog) GetByComponent(name string) []ChangeEntry {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	var result []ChangeEntry
	for _, e := range cl.entries {
		if e.ComponentName == name {
			result = append(result, e)
		}
	}
	return result
}

// Count returns the total number of changes
func (cl *Changelog) Count() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return len(cl.entries)
}
