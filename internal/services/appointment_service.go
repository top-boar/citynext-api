package services

import (
	apiModels "citynext/internal/api/models"
	"citynext/internal/database"
	dbModels "citynext/internal/database/models"
	"context"
	"log/slog"
)

type AppointmentService struct {
	repo           database.AppointmentRepository
	holidayService HolidayServiceInterface
	logger         *slog.Logger
}

func NewAppointmentService(repo database.AppointmentRepository, holidayService HolidayServiceInterface, logger *slog.Logger) *AppointmentService {
	return &AppointmentService{
		repo:           repo,
		holidayService: holidayService,
		logger:         logger,
	}
}

type CreateAppointmentRequest struct {
	FirstName string         `json:"firstName"`
	LastName  string         `json:"lastName"`
	VisitDate apiModels.Date `json:"visitDate"`
}

// creates a new appointment with validation
func (s *AppointmentService) CreateAppointment(ctx context.Context, req *CreateAppointmentRequest) (*dbModels.Appointment, error) {
	s.logger.Info("Creating appointment",
		"first_name", req.FirstName,
		"last_name", req.LastName,
		"visit_date", req.VisitDate.String())

	if req.FirstName == "" || req.LastName == "" {
		s.logger.Warn("Invalid input: missing first or last name")
		return nil, ErrInvalidInput
	}

	if err := s.holidayService.ValidateDate(ctx, req.VisitDate); err != nil {
		s.logger.Warn("Date validation failed",
			"error", err,
			"visit_date", req.VisitDate.String())
		return nil, err
	}

	// Check for existing appointment
	exists, err := s.repo.ExistsByDate(ctx, req.VisitDate)
	if err != nil {
		s.logger.Error("Failed to check existing appointment",
			"error", err,
			"visit_date", req.VisitDate.String())
		return nil, err
	}

	if exists {
		s.logger.Warn("Duplicate appointment attempt", "visit_date", req.VisitDate.String())
		return nil, database.ErrDuplicateAppointment
	}

	appointment := &dbModels.Appointment{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		VisitDate: req.VisitDate,
	}

	if err := s.repo.Create(ctx, appointment); err != nil {
		s.logger.Error("Failed to create appointment",
			"error", err,
			"first_name", req.FirstName,
			"last_name", req.LastName,
			"visit_date", req.VisitDate.String())
		return nil, err
	}

	s.logger.Info("Appointment created successfully",
		"id", appointment.ID,
		"first_name", appointment.FirstName,
		"last_name", appointment.LastName,
		"visit_date", appointment.VisitDate.String())

	return appointment, nil
}
