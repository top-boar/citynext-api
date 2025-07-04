package services

import "errors"

// Custom error types for the services layer
var (
	ErrDateInPast    = errors.New("visit date cannot be in the past")
	ErrDateIsHoliday = errors.New("visit date is a public holiday")
	ErrDateIsWeekend = errors.New("visit date is a weekend")
	ErrInvalidInput  = errors.New("invalid input data")
)
