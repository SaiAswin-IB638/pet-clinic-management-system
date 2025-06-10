package model

import (
	"time"

	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	Slot   time.Time `json:"slot" gorm:"not null"`
	Reason string    `json:"reason"`
	PetID  uint      `json:"pet_id" gorm:"not null"`
	Pet    Pet       `json:"pet" gorm:"foreignKey:PetID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
