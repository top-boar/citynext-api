package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// represents a public holiday from the Nager.Date API
type Holiday struct {
	Date        string   `json:"date"`
	LocalName   string   `json:"localName"`
	Name        string   `json:"name"`
	CountryCode string   `json:"countryCode"`
	Fixed       bool     `json:"fixed"`
	Global      bool     `json:"global"`
	Counties    []string `json:"counties"`
	LaunchYear  *int     `json:"launchYear"`
	Types       []string `json:"types"`
}

type HolidayClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewHolidayClient(baseURL string, logger *slog.Logger) *HolidayClient {
	return &HolidayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// fetches public holidays for a specific year and country
func (c *HolidayClient) GetPublicHolidays(ctx context.Context, year int, countryCode string) ([]Holiday, error) {
	url := fmt.Sprintf("%s/PublicHolidays/%d/%s", c.baseURL, year, countryCode)

	c.logger.Debug("Fetching public holidays",
		"url", url,
		"year", year,
		"country_code", countryCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request",
			"error", err,
			"url", url)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch public holidays",
			"error", err,
			"url", url)
		return nil, fmt.Errorf("failed to fetch public holidays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("API returned non-OK status",
			"status_code", resp.StatusCode,
			"url", url)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var holidays []Holiday
	if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
		c.logger.Error("Failed to decode response",
			"error", err,
			"url", url)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Successfully fetched public holidays",
		"year", year,
		"country_code", countryCode,
		"count", len(holidays))
	return holidays, nil
}

func (c *HolidayClient) IsPublicHoliday(ctx context.Context, date time.Time, countryCode string) (bool, error) {
	year := date.Year()
	holidays, err := c.GetPublicHolidays(ctx, year, countryCode)
	if err != nil {
		return false, err
	}

	dateStr := date.Format("2006-01-02")
	for _, holiday := range holidays {
		if holiday.Date == dateStr {
			return true, nil
		}
	}

	return false, nil
}
