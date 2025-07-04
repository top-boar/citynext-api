package routes

import (
	"net/http"

	"citynext/internal/api/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func RegisterRoutes(router *http.ServeMux, appointmentHandler *handlers.AppointmentHandler) {

	api := humago.New(router, huma.DefaultConfig("CityNext Appointment API", "1.0.0"))

	// expose the appointment creation endpoint
	huma.Post(api, "/appointments", appointmentHandler.CreateAppointment)
}
