package database

import (
	"context"
	"log/slog"
	"sync"
	"time"

	apiModels "citynext/internal/api/models"
	dbModels "citynext/internal/database/models"
)

// implements AppointmentRepository interface using in-memory storage for testing
type MemoryAppointmentRepository struct {
	appointments map[string]*dbModels.Appointment // date string -> appointment
	mutex        sync.RWMutex
	nextID       uint
	logger       *slog.Logger
}

func NewMemoryAppointmentRepository(logger *slog.Logger) *MemoryAppointmentRepository {
	return &MemoryAppointmentRepository{
		appointments: make(map[string]*dbModels.Appointment),
		nextID:       1,
		logger:       logger,
	}
}

func (r *MemoryAppointmentRepository) Create(ctx context.Context, appointment *dbModels.Appointment) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	dateKey := appointment.VisitDate.String()

	r.logger.Info("Creating appointment in memory",
		"first_name", appointment.FirstName,
		"last_name", appointment.LastName,
		"visit_date", dateKey)

	// Check if appointment already exists
	if _, exists := r.appointments[dateKey]; exists {
		r.logger.Warn("Appointment already exists for date",
			"date", dateKey)
		return ErrDuplicateAppointment
	}

	appointment.ID = r.nextID
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()

	r.appointments[dateKey] = appointment
	r.nextID++

	r.logger.Info("Appointment created successfully in memory",
		"id", appointment.ID,
		"date", dateKey)
	return nil
}

func (r *MemoryAppointmentRepository) GetByDate(ctx context.Context, date apiModels.Date) (*dbModels.Appointment, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	dateKey := date.String()

	r.logger.Debug("Getting appointment by date from memory", "date", dateKey)

	appointment, exists := r.appointments[dateKey]
	if !exists {
		r.logger.Debug("No appointment found for date in memory", "date", dateKey)
		return nil, ErrAppointmentNotFound
	}

	r.logger.Debug("Appointment found in memory", "id", appointment.ID, "date", dateKey)
	return appointment, nil
}

func (r *MemoryAppointmentRepository) ExistsByDate(ctx context.Context, date apiModels.Date) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	dateKey := date.String()

	r.logger.Debug("Checking if appointment exists for date in memory", "date", dateKey)

	_, exists := r.appointments[dateKey]
	r.logger.Debug("Appointment existence check result in memory",
		"date", dateKey,
		"exists", exists)
	return exists, nil
}
