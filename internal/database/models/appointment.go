package models

import (
	"citynext/internal/api/models"
	"time"

	"gorm.io/gorm"
)

// represents an appointment in the database
type Appointment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	FirstName string         `gorm:"not null" json:"firstName"`
	LastName  string         `gorm:"not null" json:"lastName"`
	VisitDate models.Date    `gorm:"not null;uniqueIndex;type:date" json:"visitDate"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// specifies the table name for the Appointment model
func (Appointment) TableName() string {
	return "appointments"
}
