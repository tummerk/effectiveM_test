package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"effectiveMobile_test/internal/domain/entity"
	"github.com/google/uuid"
)

//валидатор, можно было вынести не успеваю по времени и

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription entity.Subscription) (entity.Subscription, error)
	GetById(ctx context.Context, id int) (entity.Subscription, error)
	Update(ctx context.Context, subscription entity.Subscription) (entity.Subscription, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context, filter entity.ListFilter) ([]entity.Subscription, error)
	GetTotalCost(ctx context.Context, filter entity.CostFilter) (int, error)
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, input CreateSubscriptionInput) (entity.Subscription, error) {
	if err := s.validateCreateInput(input); err != nil {
		logger(ctx).Debug("validation failed", "error", err)
		return entity.Subscription{}, err
	}

	subscription := entity.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserId:      input.UserId,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	created, err := s.repo.Create(ctx, subscription)
	if err != nil {
		logger(ctx).Error("failed to create subscription", "error", err)
		return entity.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	logger(ctx).Info("subscription created", "id", created.ID)
	return created, nil
}

func (s *SubscriptionService) validateCreateInput(input CreateSubscriptionInput) error {
	if input.ServiceName == "" {
		return ErrInvalidServiceName
	}
	if input.Price <= 0 {
		return ErrInvalidPrice
	}
	if input.UserId == uuid.Nil {
		return ErrInvalidUserId
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return ErrInvalidDateRange
	}
	return nil
}

func (s *SubscriptionService) Get(ctx context.Context, id int) (entity.Subscription, error) {
	subscription, err := s.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Subscription{}, ErrSubscriptionNotFound
		}
		logger(ctx).Error("failed to get subscription", "id", id, "error", err)
		return entity.Subscription{}, fmt.Errorf("get subscription: %w", err)
	}
	return subscription, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id int, input UpdateSubscriptionInput) (entity.Subscription, error) {
	var updated entity.Subscription

	err := s.repo.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetById(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrSubscriptionNotFound
			}
			return err
		}

		if err := s.applyUpdates(&existing, input); err != nil {
			return err
		}

		updated, err = s.repo.Update(ctx, existing)
		return err
	})

	if err != nil {
		logger(ctx).Error("failed to update subscription", "id", id, "error", err)
		return entity.Subscription{}, err
	}

	logger(ctx).Info("subscription updated", "id", id)
	return updated, nil
}

func (s *SubscriptionService) applyUpdates(sub *entity.Subscription, input UpdateSubscriptionInput) error {
	if input.ServiceName != nil {
		if *input.ServiceName == "" {
			return ErrInvalidServiceName
		}
		sub.ServiceName = *input.ServiceName
	}

	if input.Price != nil {
		if *input.Price <= 0 {
			return ErrInvalidPrice
		}
		sub.Price = *input.Price
	}

	if input.UserId != nil {
		if *input.UserId == uuid.Nil {
			return ErrInvalidUserId
		}
		sub.UserId = *input.UserId
	}

	if input.StartDate != nil {
		sub.StartDate = *input.StartDate
	}

	if input.EndDate != nil {
		sub.EndDate = input.EndDate
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return ErrInvalidDateRange
	}

	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSubscriptionNotFound
		}
		logger(ctx).Error("failed to delete subscription", "id", id, "error", err)
		return fmt.Errorf("delete subscription: %w", err)
	}

	logger(ctx).Info("subscription deleted", "id", id)
	return nil
}

func (s *SubscriptionService) GetAll(ctx context.Context, userId *uuid.UUID, serviceName *string) ([]entity.Subscription, error) {
	filter := entity.ListFilter{
		UserId:      userId,
		ServiceName: serviceName,
	}

	subscriptions, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		logger(ctx).Error("failed to list subscriptions", "error", err)
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *SubscriptionService) TotalCost(ctx context.Context, input TotalCostInput) (int, error) {
	endDate := input.EndDate
	if endDate.After(time.Now()) {
		endDate = time.Now()
	}

	if endDate.Before(input.StartDate) {
		return 0, ErrInvalidDateRange
	}

	filter := entity.CostFilter{
		DateStart:   input.StartDate,
		DateEnd:     endDate,
		UserId:      input.UserId,
		ServiceName: input.ServiceName,
	}

	cost, err := s.repo.GetTotalCost(ctx, filter)
	if err != nil {
		logger(ctx).Error("failed to calculate total cost", "error", err)
		return 0, fmt.Errorf("calculate total cost: %w", err)
	}

	return cost, nil
}
