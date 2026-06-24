package service

import "messenger/backend/internal/repository"

// Service groups application use cases.
// Handlers should stay thin and delegate business rules here.
type Service struct {
	store repository.Store
}

// New creates an application service layer.
func New(store repository.Store) *Service {
	return &Service{store: store}
}
