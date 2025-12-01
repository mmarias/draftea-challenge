package mocks

import (
	"context"
	"log"

	"github.com/mmarias/golearn/internal/domain"
)

// CommandRepository is a mock implementation of the domain.CommandRepository interface.
type CommandRepository struct{}

// NewCommandRepository creates a new mock CommandRepository.
func NewCommandRepository() *CommandRepository {
	return &CommandRepository{}
}

// Save mocks the Save method of the CommandRepository.
// It logs the data that would have been saved to the database.
func (m *CommandRepository) Save(ctx context.Context, status domain.CommandStatus, command domain.CommandEvent, payload any) error {
	log.Printf(
		"[Mock CommandRepository] Saving command: EventType=%s, Status=%s, traceID=%s",
		command.EventType,
		status,
		command.CommandEventMetadata.TraceID,
	)
	// In a real implementation, this would save the command details to a database.
	// For the mock, we do nothing and return no error.
	return nil
}
