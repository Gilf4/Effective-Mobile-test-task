package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/Gilf4/effective-mobile-task/internal/errors"
	"github.com/Gilf4/effective-mobile-task/internal/models"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req models.CreateSubscriptionRequest) (*models.SubscriptionResponse, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.SubscriptionResponse, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.SubscriptionResponse, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID *uuid.UUID) ([]models.SubscriptionResponse, error)
	CalculateTotal(ctx context.Context, userID uuid.UUID, serviceName string, startStr, endStr string) (int, error)
}

type Handler struct {
	service SubscriptionService
	log     *slog.Logger
}

func NewHandler(service SubscriptionService, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		h.log.Error("application error", "error", appErr.Err, "code", appErr.Code, "message", appErr.Message)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(appErr.Code)
		json.NewEncoder(w).Encode(map[string]string{"error": appErr.Message})
		return
	}

	h.log.Error("unexpected error", "error", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /subscriptions", h.CreateSubscription)
	mux.HandleFunc("GET /subscriptions", h.ListSubscriptions)
	mux.HandleFunc("GET /subscriptions/{id}", h.GetSubscription)
	mux.HandleFunc("PUT /subscriptions/{id}", h.UpdateSubscription)
	mux.HandleFunc("DELETE /subscriptions/{id}", h.DeleteSubscription)
	mux.HandleFunc("GET /subscriptions/total", h.GetTotalCost)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
}

// @Summary Создать подписку
// @Description Создание новой подписки. Поле `end_date` опциональное. Если не указано - подписка бессрочная.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body models.CreateSubscriptionRequest true "Subscription info"
// @Success 201 {object} models.SubscriptionResponse
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid request body", err))
		return
	}

	h.log.Info("creating subscription",
		slog.String("user_id", req.UserID.String()),
		slog.String("service_name", req.ServiceName),
	)

	sub, err := h.service.CreateSubscription(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

// @Summary Получить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {string} string "Invalid ID format"
// @Failure 404 {string} string "Subscription not found"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid id format", err))
		return
	}

	h.log.Info("getting subscription", slog.String("id", id.String()))

	sub, err := h.service.GetSubscription(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// @Summary Обновить подписку
// @Description Обновление данных подписки. Все поля опциональные. При обновлении `end_date` проверяется, что `end_date >= start_date`.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param input body models.UpdateSubscriptionRequest true "Subscription update info (all fields optional)"
// @Success 200 {object} models.SubscriptionResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Subscription not found"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid id format", err))
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid request body", err))
		return
	}

	h.log.Info("updating subscription", slog.String("id", id.String()))

	sub, err := h.service.UpdateSubscription(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// @Summary Удалить подписку
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {string} string "Invalid ID format"
// @Failure 404 {string} string "Subscription not found"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid id format", err))
		return
	}

	h.log.Info("deleting subscription", slog.String("id", id.String()))

	err = h.service.DeleteSubscription(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Получить список подписок
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID (optional - returns all if not provided)"
// @Success 200 {array} models.SubscriptionResponse
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	userIDStr := r.URL.Query().Get("user_id")

	if userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.handleError(w, apperrors.NewBadRequest("invalid user_id format", err))
			return
		}
		userID = &parsedID
	}

	h.log.Info("listing subscriptions", slog.String("user_id", userIDStr))

	subs, err := h.service.ListSubscriptions(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

// @Summary Получение общей стоимости подписок за заданный период
// @Description Расчёт общей стоимости активных подписок за указанный период.<br><br>
// @Description **Логика расчёта:**<br>
// @Description - Подписка с `end_date` учитывается, если пересекается с запрошенным периодом (start_date <= end_period AND end_date >= start_period)<br>
// @Description - Подписка без `end_date` (бессрочная) учитывается полностью, если её `start_date <= end_period`<br>
// @Description - Бессрочная подписка всегда учитывается полной стоимостью, независимо от длины запрошенного периода
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "User UUID"
// @Param start_date query string true "Format: MM-YYYY"
// @Param end_date query string true "Format: MM-YYYY"
// @Param service_name query string false "Service Name filter (optional)"
// @Success 200 {object} models.TotalCostResponse
// @Failure 400 {string} string "Invalid parameters"
// @Failure 404 {string} string "No subscriptions found for the specified criteria"
// @Failure 500 {string} string "Internal server error"
// @Router /subscriptions/total [get]
func (h *Handler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.handleError(w, apperrors.NewBadRequest("invalid user_id format", err))
		return
	}

	if startDate == "" || endDate == "" {
		h.handleError(w, apperrors.NewBadRequest("start_date and end_date are required", nil))
		return
	}

	h.log.Info("calculating total cost",
		slog.String("user_id", userID.String()),
		slog.String("service_name", serviceName),
		slog.String("start_date", startDate),
		slog.String("end_date", endDate),
	)

	total, err := h.service.CalculateTotal(r.Context(), userID, serviceName, startDate, endDate)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TotalCostResponse{TotalCost: total})
}
