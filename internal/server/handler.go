package server

import (
	"context"
	"effectiveMobile_test/internal/domain/entity"
	"effectiveMobile_test/internal/domain/service"
	api "effectiveMobile_test/internal/server/api"
	"effectiveMobile_test/pkg/utils"
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"
)

type SubscriptionsService interface {
	Create(ctx context.Context, input service.CreateSubscriptionInput) (entity.Subscription, error)
	Get(ctx context.Context, id int) (entity.Subscription, error)
	Update(ctx context.Context, id int, input service.UpdateSubscriptionInput) (entity.Subscription, error)
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context, userId *uuid.UUID, serviceName *string) ([]entity.Subscription, error)
	TotalCost(ctx context.Context, input service.TotalCostInput) (int, error)
}

type Handler struct {
	svc SubscriptionsService
}

func NewHandler(svc SubscriptionsService) *Handler {
	return &Handler{svc: svc}
}

var _ api.StrictServerInterface = (*Handler)(nil)

func (h *Handler) GetSubscriptions(ctx context.Context, req api.GetSubscriptionsRequestObject) (api.GetSubscriptionsResponseObject, error) {
	var uid *uuid.UUID
	if req.Params.UserId != nil {
		u := uuid.UUID(*req.Params.UserId)
		uid = &u
	}

	items, err := h.svc.GetAll(ctx, uid, req.Params.ServiceName)
	if err != nil {
		return api.GetSubscriptions500JSONResponse{
			InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error:   strPtr("internal_error"),
				Message: strPtr("internal server error"),
			},
		}, nil
	}

	out := make([]api.Subscription, 0, len(items))
	for _, item := range items {
		out = append(out, toAPISub(&item))
	}
	return api.GetSubscriptions200JSONResponse(out), nil
}

func (h *Handler) PostSubscriptions(ctx context.Context, req api.PostSubscriptionsRequestObject) (api.PostSubscriptionsResponseObject, error) {
	if req.Body == nil {
		return api.PostSubscriptions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
			Error:   strPtr("bad_request"),
			Message: strPtr("empty body"),
		}}, nil
	}

	start, err := utils.ParseMonthYear(req.Body.StartDate)
	if err != nil {
		return api.PostSubscriptions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
			Error:   strPtr("invalid_start_date"),
			Message: strPtr(err.Error()),
		}}, nil
	}

	var endPtr *time.Time
	if req.Body.EndDate != nil {
		ed, err := utils.ParseMonthYear(*req.Body.EndDate)
		if err != nil {
			return api.PostSubscriptions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error:   strPtr("invalid_end_date"),
				Message: strPtr(err.Error()),
			}}, nil
		}
		endPtr = &ed
	}

	u := uuid.UUID(req.Body.UserId)

	in := service.CreateSubscriptionInput{
		ServiceName: strings.TrimSpace(req.Body.ServiceName),
		Price:       req.Body.Price,
		UserId:      u,
		StartDate:   start,
		EndDate:     endPtr,
	}

	created, err := h.svc.Create(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidServiceName),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrInvalidDateRange):
			return api.PostSubscriptions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error:   strPtr("validation_error"),
				Message: strPtr(err.Error()),
			}}, nil
		default:
			return api.PostSubscriptions500JSONResponse{InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error:   strPtr("internal_error"),
				Message: strPtr("internal server error"),
			}}, nil
		}
	}

	return api.PostSubscriptions201JSONResponse(toAPISub(&created)), nil
}

func (h *Handler) GetSubscriptionsTotalCost(ctx context.Context, req api.GetSubscriptionsTotalCostRequestObject) (api.GetSubscriptionsTotalCostResponseObject, error) {
	start, err := utils.ParseMonthYear(req.Params.StartDate)
	if err != nil {
		return api.GetSubscriptionsTotalCost400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
			Error:   strPtr("invalid_start_date"),
			Message: strPtr(err.Error()),
		}}, nil
	}
	end, err := utils.ParseMonthYear(req.Params.EndDate)
	if err != nil {
		return api.GetSubscriptionsTotalCost400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
			Error:   strPtr("invalid_end_date"),
			Message: strPtr(err.Error()),
		}}, nil
	}

	var uid *uuid.UUID
	if req.Params.UserId != nil {
		u := uuid.UUID(*req.Params.UserId)
		uid = &u
	}

	in := service.TotalCostInput{
		StartDate:   start,
		EndDate:     end,
		UserId:      uid,
		ServiceName: req.Params.ServiceName,
	}

	total, err := h.svc.TotalCost(ctx, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDateRange):
			return api.GetSubscriptionsTotalCost400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error:   strPtr("validation_error"),
				Message: strPtr(err.Error()),
			}}, nil
		default:
			return api.GetSubscriptionsTotalCost500JSONResponse{InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error:   strPtr("internal_error"),
				Message: strPtr("internal server error"),
			}}, nil
		}
	}

	resp := api.TotalCostResponse{
		TotalCost:   intPtr(total),
		PeriodStart: &req.Params.StartDate,
		PeriodEnd:   &req.Params.EndDate,
		UserId:      req.Params.UserId,
		ServiceName: req.Params.ServiceName,
	}
	return api.GetSubscriptionsTotalCost200JSONResponse(resp), nil
}

func (h *Handler) DeleteSubscriptionsId(ctx context.Context, req api.DeleteSubscriptionsIdRequestObject) (api.DeleteSubscriptionsIdResponseObject, error) {
	if err := h.svc.Delete(ctx, req.Id); err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			return api.DeleteSubscriptionsId404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{
				Error:   strPtr("not_found"),
				Message: strPtr("subscription not found"),
			}}, nil
		}
		return api.DeleteSubscriptionsId500JSONResponse{InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
			Error:   strPtr("internal_error"),
			Message: strPtr("internal server error"),
		}}, nil
	}
	return api.DeleteSubscriptionsId204Response{}, nil
}

func (h *Handler) GetSubscriptionsId(ctx context.Context, req api.GetSubscriptionsIdRequestObject) (api.GetSubscriptionsIdResponseObject, error) {
	sub, err := h.svc.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			return api.GetSubscriptionsId404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{
				Error:   strPtr("not_found"),
				Message: strPtr("subscription not found"),
			}}, nil
		}
		return api.GetSubscriptionsId500JSONResponse{InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
			Error:   strPtr("internal_error"),
			Message: strPtr(err.Error()),
		}}, nil
	}
	return api.GetSubscriptionsId200JSONResponse(toAPISub(&sub)), nil
}

func (h *Handler) PutSubscriptionsId(ctx context.Context, req api.PutSubscriptionsIdRequestObject) (api.PutSubscriptionsIdResponseObject, error) {
	if req.Body == nil {
		return api.PutSubscriptionsId400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
			Error:   strPtr("bad_request"),
			Message: strPtr("empty body"),
		}}, nil
	}
	var startPtr *time.Time
	if req.Body.StartDate != nil {
		d, err := utils.ParseMonthYear(strings.TrimSpace(*req.Body.StartDate))
		if err != nil {
			return api.PutSubscriptionsId400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error:   strPtr("invalid_start_date"),
				Message: strPtr(err.Error()),
			}}, nil
		}
		startPtr = &d
	}
	var endPtr *time.Time
	clearEnd := false
	if req.Body.EndDate != nil {
		ed := strings.TrimSpace(*req.Body.EndDate)
		if ed == "" {
			clearEnd = true
		} else {
			d, err := utils.ParseMonthYear(ed)
			if err != nil {
				return api.PutSubscriptionsId400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
					Error:   strPtr("invalid_end_date"),
					Message: strPtr(err.Error()),
				}}, nil
			}
			endPtr = &d
		}
	}
	var uidPtr *uuid.UUID
	if req.Body.UserId != nil {
		u := uuid.UUID(*req.Body.UserId)
		uidPtr = &u
	}
	var serviceNamePtr *string
	if req.Body.ServiceName != nil {
		s := strings.TrimSpace(*req.Body.ServiceName)
		serviceNamePtr = &s
	}
	in := service.UpdateSubscriptionInput{
		ServiceName:      serviceNamePtr,
		Price:            req.Body.Price,
		UserId:           uidPtr,
		StartDate:        startPtr,
		EndDate:          endPtr,
		SetEndDateToNull: clearEnd,
	}
	updated, err := h.svc.Update(ctx, req.Id, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidServiceName),
			errors.Is(err, service.ErrInvalidPrice),
			errors.Is(err, service.ErrInvalidDateRange):
			return api.PutSubscriptionsId400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{
				Error:   strPtr("validation_error"),
				Message: strPtr(err.Error()),
			}}, nil
		case errors.Is(err, service.ErrSubscriptionNotFound):
			return api.PutSubscriptionsId404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{
				Error:   strPtr("not_found"),
				Message: strPtr("subscription not found"),
			}}, nil
		default:
			return api.PutSubscriptionsId500JSONResponse{InternalServerErrorJSONResponse: api.InternalServerErrorJSONResponse{
				Error:   strPtr("internal_error"),
				Message: strPtr("internal server error"),
			}}, nil
		}
	}
	return api.PutSubscriptionsId200JSONResponse(toAPISub(&updated)), nil
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
