package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Gilf4/effective-mobile-task/internal/config"
	apperrors "github.com/Gilf4/effective-mobile-task/internal/errors"
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
		return apperrors.NewInternal(err)
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
			return nil, apperrors.NewNotFound("subscription not found", err)
		}
		return nil, apperrors.NewInternal(err)
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
		return apperrors.NewInternal(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.NewNotFound("subscription not found", nil)
	}

	return nil
}

// Delete удаляет подписку по ID
func (s *SubscriptionStorage) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	cmdTag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return apperrors.NewInternal(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.NewNotFound("subscription not found", nil)
	}

	return nil
}

// ListByUserID возвращает список подписок пользователя
func (s *SubscriptionStorage) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, apperrors.NewInternal(err)
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
			return nil, apperrors.NewInternal(err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.NewInternal(err)
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
		return nil, apperrors.NewInternal(err)
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
			return nil, apperrors.NewInternal(err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.NewInternal(err)
	}

	return subscriptions, nil
}

// GetTotalCost считает сумму стоимости подписок за период
// Параметр serviceName опциональный
// Вариант A: бессрочные подписки (без end_date) учитываются полностью, если активны в запрошенном периоде
func (r *SubscriptionStorage) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, startPeriod, endPeriod time.Time) (int, error) {
	query := `
		SELECT SUM(price) as total
		FROM subscriptions
		WHERE user_id = $1
		  AND start_date <= $3
		  AND (
		    end_date IS NULL 
		    OR end_date >= $2
		  )
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
		return 0, apperrors.NewInternal(err)
	}

	if total == nil {
		return 0, apperrors.NewNotFound("no subscriptions found for the specified criteria", nil)
	}

	return *total, nil
}
