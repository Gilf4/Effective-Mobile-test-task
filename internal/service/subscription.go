package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Gilf4/effective-mobile-task/internal/models"
	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID *uuid.UUID) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, startPeriod, endPeriod time.Time) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req models.CreateSubscriptionRequest) (*models.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return models.NewSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.SubscriptionResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return models.NewSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.SubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		date, err := parseDate(*req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		sub.StartDate = date
	}
	if req.EndDate != nil {
		date, err := parseDate(*req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		sub.EndDate = &date
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return models.NewSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID) ([]models.SubscriptionResponse, error) {
	subscriptions, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	responses := make([]models.SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		responses[i] = *models.NewSubscriptionResponse(&sub)
	}

	return responses, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, userID uuid.UUID, serviceName string, startStr, endStr string) (int, error) {
	start, err := parseDate(startStr)
	if err != nil {
		return 0, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := parseDate(endStr)
	if err != nil {
		return 0, fmt.Errorf("invalid end_date: %w", err)
	}

	total, err := s.repo.GetTotalCost(ctx, userID, serviceName, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cost: %w", err)
	}

	return total, nil
}

func parseDate(dateStr string) (time.Time, error) {
	t, err := time.Parse("01-2006", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}
	return t, nil
}
