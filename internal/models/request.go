package models

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidPrice       = errors.New("price must be greater than 0")
	ErrInvalidServiceName = errors.New("service name is required")
	ErrInvalidUserID      = errors.New("user id is required")
	ErrInvalidDate        = errors.New("invalid date format (expected MM-YYYY)")
)

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

func (r CreateSubscriptionRequest) Validate() error {
	if r.ServiceName == "" {
		return ErrInvalidServiceName
	}
	if r.Price <= 0 {
		return ErrInvalidPrice
	}
	if r.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if r.StartDate == "" {
		return ErrInvalidDate
	}
	return nil
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name"`
	Price       *int    `json:"price"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

func (r UpdateSubscriptionRequest) Validate() error {
	if r.Price != nil && *r.Price <= 0 {
		return ErrInvalidPrice
	}
	if r.ServiceName != nil && *r.ServiceName == "" {
		return ErrInvalidServiceName
	}
	return nil
}

type ListSubscriptionsRequest struct {
	UserID *uuid.UUID
	Limit  int
	Offset int
}

func (r ListSubscriptionsRequest) Validate() error {
	if r.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if r.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}
	if r.Offset < 0 {
		return errors.New("offset cannot be negative")
	}
	return nil
}

func (r *ListSubscriptionsRequest) SetDefaults() {
	if r.Limit == 0 {
		r.Limit = 20
	}
}
