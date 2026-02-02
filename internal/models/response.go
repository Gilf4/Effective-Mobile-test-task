package models

import (
	"time"

	"github.com/google/uuid"
)

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewSubscriptionResponse(sub *Subscription) *SubscriptionResponse {
	return &SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
		UpdatedAt:   sub.UpdatedAt,
	}
}

type PaginatedSubscriptionResponse struct {
	Data    []SubscriptionResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasMore bool                   `json:"has_more"`
}

func (p *PaginatedSubscriptionResponse) CalculateHasMore() {
	p.HasMore = int64(p.Offset+len(p.Data)) < p.Total
}
