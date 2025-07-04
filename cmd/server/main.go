package main

import (
	"net/http"
	"os"

	"citynext/internal/api/handlers"
	"citynext/internal/api/routes"
	"citynext/internal/config"
	"citynext/internal/database"
	"citynext/internal/logger"
	"citynext/internal/services"
)

func main() {

	cfg := config.Load()

	log := logger.New(cfg.LogLevel)
	log.Info("Starting CityNext Appointment API",
		"version", "1.0.0",
		"port", cfg.ServerPort,
		"db_path", cfg.DBPath,
		"log_level", cfg.LogLevel.String(),
		"nager_api_url", cfg.NagerAPIBaseURL)

	db, err := database.NewSQLiteConnection(cfg.DBPath)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.CloseConnection(db); err != nil {
			log.Error("Failed to close database connection", "error", err)
		}
	}()

	appointmentRepo := database.NewSQLiteAppointmentRepository(db, log.Logger)

	holidayService := services.NewHolidayService(cfg.NagerAPIBaseURL, log.Logger)
	appointmentService := services.NewAppointmentService(appointmentRepo, holidayService, log.Logger)

	appointmentHandler := handlers.NewAppointmentHandler(appointmentService, log.Logger)

	router := http.NewServeMux()
	routes.RegisterRoutes(router, appointmentHandler)

	http.ListenAndServe(":"+cfg.ServerPort, router)
}
