package service

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type SubscriptionRepo interface {
	Create(ctx context.Context, serviceName string, userId uuid.UUID, startDate time.Time) error
}

type SubscriptionService struct {
	repo SubscriptionRepo
}

func (s *SubscriptionService) Create(ctx context.Context,
	serviceName string, userId uuid.UUID, startDate time.Time) error {
	err := s.repo.Create(ctx, serviceName, userId, startDate)
	if err != nil {
		
	}
}
