package database

import "errors"

// Custom error types for the database layer
var (
	ErrDuplicateAppointment = errors.New("appointment already exists for this date")
	ErrAppointmentNotFound  = errors.New("appointment not found")
)
