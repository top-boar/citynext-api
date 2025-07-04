package models

// represents the input for creating an appointment
type CreateAppointmentInput struct {
	Body struct {
		FirstName string `json:"firstName" example:"John" doc:"First name of the person" maxLength:"50"`
		LastName  string `json:"lastName" example:"Doe" doc:"Last name of the person" maxLength:"50"`
		VisitDate Date   `json:"visitDate" example:"2025-08-15" doc:"Visit date (YYYY-MM-DD format)"`
	}
}

// represents the output of a successful appointment creation
type CreateAppointmentOutput struct {
	Body struct {
		ID        uint   `json:"id" example:"1" doc:"Appointment ID"`
		FirstName string `json:"firstName" example:"John" doc:"First name of the person"`
		LastName  string `json:"lastName" example:"Doe" doc:"Last name of the person"`
		VisitDate Date   `json:"visitDate" example:"2025-08-15" doc:"Visit date"`
		CreatedAt string `json:"createdAt" example:"2025-08-15T10:30:00Z" doc:"Creation timestamp"`
	}
}
