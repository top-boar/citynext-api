package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"citynext/internal/api/models"
	"citynext/internal/database"
	"citynext/internal/services"

	"github.com/danielgtaylor/huma/v2"
)

type AppointmentHandler struct {
	appointmentService *services.AppointmentService
	logger             *slog.Logger
}

func NewAppointmentHandler(appointmentService *services.AppointmentService, logger *slog.Logger) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
		logger:             logger,
	}
}

func (h *AppointmentHandler) CreateAppointment(ctx context.Context, input *models.CreateAppointmentInput) (*models.CreateAppointmentOutput, error) {
	h.logger.Info("Received appointment creation request",
		"first_name", input.Body.FirstName,
		"last_name", input.Body.LastName,
		"visit_date", input.Body.VisitDate.String())

	req := &services.CreateAppointmentRequest{
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
		VisitDate: input.Body.VisitDate,
	}

	appointment, err := h.appointmentService.CreateAppointment(ctx, req)
	if err != nil {
		h.logger.Error("Failed to create appointment",
			"error", err,
			"first_name", input.Body.FirstName,
			"last_name", input.Body.LastName,
			"visit_date", input.Body.VisitDate.String())

		// map domain errors to HTTP errors
		switch err {
		case services.ErrDateInPast:
			return nil, huma.Error422UnprocessableEntity("Visit date cannot be in the past")
		case services.ErrDateIsHoliday:
			return nil, huma.Error422UnprocessableEntity("Visit date is a public holiday")
		case services.ErrDateIsWeekend:
			return nil, huma.Error422UnprocessableEntity("Visit date is a weekend")
		case database.ErrDuplicateAppointment:
			return nil, huma.Error422UnprocessableEntity("An appointment already exists for this date")
		case services.ErrInvalidInput:
			return nil, huma.Error422UnprocessableEntity("Invalid input data")
		default:
			return nil, huma.Error500InternalServerError(fmt.Sprintf("Internal server error: %v", err))
		}
	}

	output := &models.CreateAppointmentOutput{}
	output.Body.ID = appointment.ID
	output.Body.FirstName = appointment.FirstName
	output.Body.LastName = appointment.LastName
	output.Body.VisitDate = appointment.VisitDate
	output.Body.CreatedAt = appointment.CreatedAt.Format("2006-01-02T15:04:05Z")

	h.logger.Info("Appointment created successfully via API",
		"id", appointment.ID,
		"first_name", appointment.FirstName,
		"last_name", appointment.LastName,
		"visit_date", appointment.VisitDate.String())

	return output, nil
}
