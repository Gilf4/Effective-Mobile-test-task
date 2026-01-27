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
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, startPeriod, endPeriod time.Time) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req models.CreateSubscriptionRequest) (*models.Subscription, error) {
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
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		sub.StartDate = date
	}
	if req.EndDate != nil {
		date, err := parseDate(*req.EndDate)
		if err != nil {
			return nil, err
		}
		sub.EndDate = &date
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, userID uuid.UUID, serviceName string, startStr, endStr string) (int, error) {
	start, err := parseDate(startStr)
	if err != nil {
		return 0, err
	}
	end, err := parseDate(endStr)
	if err != nil {
		return 0, err
	}

	return s.repo.GetTotalCost(ctx, userID, serviceName, start, end)
}

func parseDate(dateStr string) (time.Time, error) {
	t, err := time.Parse("01-2006", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}
	return t, nil
}
