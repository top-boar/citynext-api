package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	apiModels "citynext/internal/api/models"
	"citynext/pkg/client"
)

// interface for holiday service operations
type HolidayServiceInterface interface {
	IsPublicHoliday(ctx context.Context, date apiModels.Date) (bool, error)
	ValidateDate(ctx context.Context, date apiModels.Date) error
}

type HolidayService struct {
	client *client.HolidayClient
	cache  map[int]map[string]bool // year -> date -> isHoliday
	mutex  sync.RWMutex
	logger *slog.Logger
}

func NewHolidayService(baseURL string, logger *slog.Logger) HolidayServiceInterface {
	return &HolidayService{
		client: client.NewHolidayClient(baseURL, logger),
		cache:  make(map[int]map[string]bool),
		logger: logger,
	}
}

func (s *HolidayService) IsPublicHoliday(ctx context.Context, date apiModels.Date) (bool, error) {
	year := date.Time.Year()
	dateStr := date.String()

	s.logger.Debug("Checking if date is public holiday",
		"date", dateStr,
		"year", year)

	// Check cache first
	s.mutex.RLock()
	if yearCache, exists := s.cache[year]; exists {
		if isHoliday, found := yearCache[dateStr]; found {
			s.logger.Debug("Cache hit for holiday check",
				"date", dateStr,
				"is_holiday", isHoliday)
			s.mutex.RUnlock()
			return isHoliday, nil
		}
	}
	s.mutex.RUnlock()

	s.logger.Debug("Cache miss for holiday check, fetching from API", "date", dateStr)

	// Fetch holidays for the year
	holidays, err := s.client.GetPublicHolidays(ctx, year, "GB")
	if err != nil {
		s.logger.Error("Failed to fetch holidays",
			"error", err,
			"year", year)
		return false, err
	}

	// Build cache for the year
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cache[year] == nil {
		s.cache[year] = make(map[string]bool)
	}

	// Populate cache and check the specific date
	isHoliday := false
	for _, holiday := range holidays {
		s.cache[year][holiday.Date] = true
		if holiday.Date == dateStr {
			isHoliday = true
		}
	}

	s.logger.Info("Updated holiday cache",
		"year", year,
		"holidays_count", len(holidays),
		"date", dateStr,
		"is_holiday", isHoliday)

	return isHoliday, nil
}

func (s *HolidayService) ValidateDate(ctx context.Context, visitDate apiModels.Date) error {
	s.logger.Debug("Validating appointment date", "date", visitDate.String())

	// Check if date is in the past
	now := time.Now().UTC().Truncate(24 * time.Hour)
	date := visitDate.Time.UTC().Truncate(24 * time.Hour)
	if date.Before(now) {
		s.logger.Warn("Attempted to book appointment in the past", "date", visitDate.String())
		return ErrDateInPast
	}

	// Check if date is a weekend (Saturday = 6, Sunday = 0)
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		s.logger.Warn("Attempted to book appointment on weekend",
			"date", visitDate.String(),
			"weekday", weekday.String())
		return ErrDateIsWeekend
	}

	// Check if date is a public holiday
	isHoliday, err := s.IsPublicHoliday(ctx, visitDate)
	if err != nil {
		s.logger.Error("Failed to check if date is holiday",
			"error", err,
			"date", visitDate.String())
		return err
	}

	if isHoliday {
		s.logger.Warn("Attempted to book appointment on public holiday", "date", visitDate.String())
		return ErrDateIsHoliday
	}

	s.logger.Debug("Date validation passed", "date", visitDate.String())
	return nil
}
