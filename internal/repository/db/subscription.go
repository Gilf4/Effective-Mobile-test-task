package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gilf4/effective-mobile-task/internal/config"
	"github.com/Gilf4/effective-mobile-task/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionStorage struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(ctx context.Context, dbCfg *config.DBConfig) (*SubscriptionStorage, error) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &SubscriptionStorage{db: pool}, nil
}

// Create создает новую запись о подписке
func (s *SubscriptionStorage) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID получает подписку по ID
func (s *SubscriptionStorage) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var sub models.Subscription
	err := s.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

// Update обновляет данные подписки
func (s *SubscriptionStorage) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = NOW()
		WHERE id = $5
	`

	cmdTag, err := s.db.Exec(ctx, query,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

// Delete удаляет подписку по ID
func (s *SubscriptionStorage) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	cmdTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

// ListByUserID возвращает список активных подписок пользователя
func (s *SubscriptionStorage) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// List возвращает список всех подписок (опционально с фильтром по user_id)
func (s *SubscriptionStorage) List(ctx context.Context, userID *uuid.UUID) ([]models.Subscription, error) {
	var query string
	var rows pgx.Rows
	var err error

	if userID != nil {
		query = `
			SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
			FROM subscriptions
			WHERE user_id = $1
			ORDER BY created_at DESC
		`
		rows, err = s.db.Query(ctx, query, *userID)
	} else {
		query = `
			SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
			FROM subscriptions
			ORDER BY created_at DESC
		`
		rows, err = s.db.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// GetTotalCost считает сумму стоимсоти подписок за период
// Параметр serviceName опциональный
func (r *SubscriptionStorage) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, startPeriod, endPeriod time.Time) (int, error) {
	query := `
		SELECT SUM(price) as total
		FROM subscriptions
		WHERE user_id = $1
		  AND start_date <= $3
		  AND (end_date IS NULL OR end_date >= $2)
	`

	args := []any{userID, startPeriod, endPeriod}
	argIdx := 4

	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, serviceName)
		argIdx++
	}

	var total *int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cost: %w", err)
	}

	if total == nil {
		return 0, ErrNoSubscriptionsFound
	}

	return *total, nil
}
