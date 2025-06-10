package initializers

import (
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
)

func MigrateDB() error {

	err := DB.AutoMigrate(
		&model.User{},
		&model.Pet{},
		&model.Appointment{},
	)

	return err
}
