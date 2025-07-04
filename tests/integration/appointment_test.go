package integration

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"citynext/internal/api/handlers"
	apiModels "citynext/internal/api/models"
	"citynext/internal/api/routes"
	"citynext/internal/database"
	"citynext/internal/services"

	"github.com/stretchr/testify/assert"
)

func TestAppointmentAPI_Integration(t *testing.T) {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Setup in-memory repository for testing
	repo := database.NewMemoryAppointmentRepository(logger)
	holidayService := services.NewHolidayService("https://date.nager.at/api/v3", logger)
	appointmentService := services.NewAppointmentService(repo, holidayService, logger)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService, logger)

	router := http.NewServeMux()
	routes.RegisterRoutes(router, appointmentHandler)

	createDate := func(t time.Time) apiModels.Date {
		return apiModels.Date{Time: t.UTC()}
	}

	// Helper to get the next N unique weekdays (skipping weekends)
	nextWeekdays := func(start time.Time, n int) []apiModels.Date {
		var dates []apiModels.Date
		current := start
		for len(dates) < n {
			current = current.AddDate(0, 0, 1)
			if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
				dates = append(dates, createDate(current))
			}
		}
		return dates
	}

	weekdays := nextWeekdays(time.Now(), 3)

	t.Run("CreateAppointment_Success", func(t *testing.T) {
		futureDate := weekdays[0]
		requestBody := apiModels.CreateAppointmentInput{}
		requestBody.Body.FirstName = "John"
		requestBody.Body.LastName = "Doe"
		requestBody.Body.VisitDate = futureDate

		bodyBytes, _ := json.Marshal(requestBody.Body)
		req := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response apiModels.CreateAppointmentOutput
		err := json.Unmarshal(w.Body.Bytes(), &response.Body)
		assert.NoError(t, err)
		assert.Equal(t, "John", response.Body.FirstName)
		assert.Equal(t, "Doe", response.Body.LastName)
		assert.Equal(t, futureDate.String(), response.Body.VisitDate.String())
		assert.NotZero(t, response.Body.ID)
	})

	t.Run("CreateAppointment_DuplicateDate", func(t *testing.T) {
		futureDate := weekdays[1]
		requestBody := apiModels.CreateAppointmentInput{}
		requestBody.Body.FirstName = "Jane"
		requestBody.Body.LastName = "Smith"
		requestBody.Body.VisitDate = futureDate

		bodyBytes, _ := json.Marshal(requestBody.Body)
		// First request should succeed
		req1 := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Second request with same date should fail
		req2 := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusUnprocessableEntity, w2.Code)
	})

	t.Run("CreateAppointment_InvalidInput", func(t *testing.T) {
		futureDate := weekdays[2]
		requestBody := map[string]interface{}{
			"lastName":  "Doe",
			"visitDate": futureDate.String(),
		}

		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("CreateAppointment_PastDate", func(t *testing.T) {
		pastDate := createDate(time.Now().AddDate(0, 0, -1)) // yesterday
		requestBody := apiModels.CreateAppointmentInput{}
		requestBody.Body.FirstName = "John"
		requestBody.Body.LastName = "Doe"
		requestBody.Body.VisitDate = pastDate

		bodyBytes, _ := json.Marshal(requestBody.Body)
		req := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("CreateAppointment_WeekendDate", func(t *testing.T) {
		// Find the next Saturday (weekday 6)
		now := time.Now()
		daysUntilSaturday := (6 - int(now.Weekday()) + 7) % 7
		if daysUntilSaturday == 0 {
			daysUntilSaturday = 7 // If today is Saturday, use next Saturday
		}
		weekendDate := createDate(now.AddDate(0, 0, daysUntilSaturday))

		requestBody := apiModels.CreateAppointmentInput{}
		requestBody.Body.FirstName = "John"
		requestBody.Body.LastName = "Doe"
		requestBody.Body.VisitDate = weekendDate

		bodyBytes, _ := json.Marshal(requestBody.Body)
		req := httptest.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
