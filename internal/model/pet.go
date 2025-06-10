package model

import (
	"gorm.io/gorm"
)

type Pet struct {
	gorm.Model
	Name           string `json:"name"`
	Species        string `json:"species"`
	Breed          string `json:"breed"`
	OwnerID        uint   `json:"owner_id"`
	MedicalHistory string `json:"medical_history"`
}
