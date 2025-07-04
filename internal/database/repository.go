package database

import (
	apiModels "citynext/internal/api/models"
	dbModels "citynext/internal/database/models"
	"context"
	"log/slog"

	"gorm.io/gorm"
)

// interface for appointment data operations
type AppointmentRepository interface {
	Create(ctx context.Context, appointment *dbModels.Appointment) error
	GetByDate(ctx context.Context, date apiModels.Date) (*dbModels.Appointment, error)
	ExistsByDate(ctx context.Context, date apiModels.Date) (bool, error)
}

// SQLite implementation of the AppointmentRepository interface
type SQLiteAppointmentRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewSQLiteAppointmentRepository(db *gorm.DB, logger *slog.Logger) *SQLiteAppointmentRepository {
	return &SQLiteAppointmentRepository{
		db:     db,
		logger: logger,
	}
}

// saves a new appointment to the database
func (r *SQLiteAppointmentRepository) Create(ctx context.Context, appointment *dbModels.Appointment) error {
	r.logger.Info("Creating appointment",
		"first_name", appointment.FirstName,
		"last_name", appointment.LastName,
		"visit_date", appointment.VisitDate.String())

	err := r.db.WithContext(ctx).Create(appointment).Error
	if err != nil {
		r.logger.Error("Failed to create appointment",
			"error", err,
			"first_name", appointment.FirstName,
			"last_name", appointment.LastName,
			"visit_date", appointment.VisitDate.String())
		return err
	}

	r.logger.Info("Appointment created successfully", "id", appointment.ID)
	return nil
}

// retrieves an appointment by date
func (r *SQLiteAppointmentRepository) GetByDate(ctx context.Context, date apiModels.Date) (*dbModels.Appointment, error) {
	r.logger.Debug("Getting appointment by date", "date", date.String())

	var appointment dbModels.Appointment
	err := r.db.WithContext(ctx).Where("DATE(visit_date) = DATE(?)", date.String()).First(&appointment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("No appointment found for date", "date", date.String())
			return nil, ErrAppointmentNotFound
		}
		r.logger.Error("Failed to get appointment by date",
			"error", err,
			"date", date.String())
		return nil, err
	}

	r.logger.Debug("Appointment found", "id", appointment.ID, "date", date.String())
	return &appointment, nil
}

// checks if an appointment exists for a given date
func (r *SQLiteAppointmentRepository) ExistsByDate(ctx context.Context, date apiModels.Date) (bool, error) {
	r.logger.Debug("Checking if appointment exists for date", "date", date.String())

	var count int64
	err := r.db.WithContext(ctx).Model(&dbModels.Appointment{}).Where("DATE(visit_date) = DATE(?)", date.String()).Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to check appointment existence",
			"error", err,
			"date", date.String())
		return false, err
	}

	exists := count > 0
	r.logger.Debug("Appointment existence check result",
		"date", date.String(),
		"exists", exists)
	return exists, nil
}
