package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Valentin0851/avito-recap-backend/internal/delivery/middleware"
	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

const maxAdminCardRequestBodyBytes = 1 << 20

type AdminCardDefinitionService interface {
	Create(
		ctx context.Context,
		adminID uuid.UUID,
		definition *domain.CardDefinition,
	) (*domain.CardDefinition, error)
	List(ctx context.Context) ([]domain.CardDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdminCardDefinitionHandler struct {
	service AdminCardDefinitionService
	logger  *slog.Logger
}

type CreateCardDefinitionRequest struct {
	Name              string                       `json:"name"`
	TargetUserID      *uuid.UUID                   `json:"target_user_id"`
	Kind              domain.CardDefinitionKind    `json:"kind"`
	Metric            domain.CardMetric            `json:"metric"`
	Analysis          domain.CardAnalysis          `json:"analysis"`
	ConditionOperator domain.CardConditionOperator `json:"condition_operator"`
	ConditionValue    *float64                     `json:"condition_value"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	ValueSuffix       string                       `json:"value_suffix"`
	Layout            string                       `json:"layout"`
	Theme             string                       `json:"theme"`
	Icon              string                       `json:"icon"`
	Shareable         *bool                        `json:"shareable"`
	SortOrder         *int                         `json:"sort_order"`
	IsActive          *bool                        `json:"is_active"`
}

func NewAdminCardDefinitionHandler(
	service AdminCardDefinitionService,
	logger *slog.Logger,
) *AdminCardDefinitionHandler {
	return &AdminCardDefinitionHandler{service: service, logger: logger}
}

func (h *AdminCardDefinitionHandler) Register(mux *http.ServeMux) {
	mux.Handle(
		"GET /api/admin/card-definitions/options",
		middleware.RequireAdmin(http.HandlerFunc(h.Options)),
	)
	mux.Handle(
		"GET /api/admin/card-definitions",
		middleware.RequireAdmin(http.HandlerFunc(h.List)),
	)
	mux.Handle(
		"POST /api/admin/card-definitions",
		middleware.RequireAdmin(http.HandlerFunc(h.Create)),
	)
	mux.Handle(
		"DELETE /api/admin/card-definitions/{id}",
		middleware.RequireAdmin(http.HandlerFunc(h.Delete)),
	)
}

func (h *AdminCardDefinitionHandler) Options(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"kinds":      []string{"statistic", "highlight"},
		"metrics":    []string{"total_views", "favorites", "chats", "purchases", "sales", "listing_views", "contacts", "reviews", "activity_days", "categories", "deals"},
		"analyses":   []string{"total", "monthly_average", "monthly_max"},
		"conditions": []string{"always", "gt", "gte", "lt", "lte", "eq"},
		"layouts":    []string{"statistic", "hero"},
		"monthly_metrics": []string{
			"total_views",
			"favorites",
			"purchases",
			"sales",
		},
	})
}

func (h *AdminCardDefinitionHandler) List(w http.ResponseWriter, r *http.Request) {
	definitions, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, r, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": definitions})
}

func (h *AdminCardDefinitionHandler) Create(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.AccountIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	request, err := decodeCreateCardDefinitionRequest(w, r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	definition := request.cardDefinition()
	created, err := h.service.Create(r.Context(), adminID, &definition)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AdminCardDefinitionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "id must be a valid non-zero UUID")
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminCardDefinitionHandler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidCardDefinition):
		writeError(w, r, http.StatusBadRequest, "invalid_card_definition", err.Error())
	case errors.Is(err, apperrors.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "card definition or target user not found")
	default:
		writeServiceError(w, r, h.logger, err)
	}
}

func decodeCreateCardDefinitionRequest(
	w http.ResponseWriter,
	r *http.Request,
) (CreateCardDefinitionRequest, error) {
	var request CreateCardDefinitionRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxAdminCardRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			return request, errors.New("request body is required")
		}
		return request, errors.New("request body must be valid JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return request, errors.New("request body must contain a single JSON object")
	}
	return request, nil
}

func (request CreateCardDefinitionRequest) cardDefinition() domain.CardDefinition {
	definition := domain.CardDefinition{
		Name:              request.Name,
		TargetUserID:      request.TargetUserID,
		Kind:              request.Kind,
		Metric:            request.Metric,
		Analysis:          request.Analysis,
		ConditionOperator: request.ConditionOperator,
		ConditionValue:    request.ConditionValue,
		Title:             request.Title,
		Description:       request.Description,
		ValueSuffix:       request.ValueSuffix,
		Layout:            request.Layout,
		Theme:             request.Theme,
		Icon:              request.Icon,
		Shareable:         true,
		SortOrder:         100,
		IsActive:          true,
	}
	if definition.Kind == "" {
		definition.Kind = domain.CardKindStatistic
	}
	if definition.Analysis == "" {
		definition.Analysis = domain.CardAnalysisTotal
	}
	if definition.ConditionOperator == "" {
		definition.ConditionOperator = domain.CardConditionAlways
	}
	if definition.Layout == "" {
		definition.Layout = "statistic"
	}
	if definition.Theme == "" {
		definition.Theme = "avito-purple"
	}
	if definition.Icon == "" {
		definition.Icon = "chart"
	}
	if request.Shareable != nil {
		definition.Shareable = *request.Shareable
	}
	if request.SortOrder != nil {
		definition.SortOrder = *request.SortOrder
	}
	if request.IsActive != nil {
		definition.IsActive = *request.IsActive
	}
	return definition
}
