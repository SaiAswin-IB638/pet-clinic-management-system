package model

import (
	"gorm.io/gorm"
)

const (
	UserTypeAdmin string = "admin"
	UserTypeStaff string = "staff"
	UserTypeOwner string = "owner"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	Role     string `json:"role" gorm:"not null"`
	Name     string `json:"name"`
	Contact  string `json:"contact"`
	Email    string `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Pets     []Pet  `json:"pets" gorm:"foreignKey:OwnerID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
