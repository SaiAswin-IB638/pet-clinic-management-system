package handlers

import (
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
)

type handlerService struct {
	petService         *service.PetService
	appointmentService *service.AppointmentService
	userService        *service.UserService
}

func NewService() *handlerService {
	petService := service.NewPetService()
	appointmentService := service.NewAppointmentService()
	userService := service.NewUserService()
	return &handlerService{
		petService:         petService,
		appointmentService: appointmentService,
		userService:        userService,
	}
}
