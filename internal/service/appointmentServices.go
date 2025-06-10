package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/initializers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type AppointmentNotFoundError struct {
	AppointmentID uint
}

func (e AppointmentNotFoundError) Error() string {
	return fmt.Sprintf("appointment not found: %v", e.AppointmentID)
}

type AppointmentFoundError struct {
	AppointmentID uint
}

func (e AppointmentFoundError) Error() string {
	return fmt.Sprintf("appointment found: %v", e.AppointmentID)
}

var ErrInvalidSlot = errors.New("invalid appointment slot")

func (appointmentService *AppointmentService) GetAppointment(id uint, ctx context.Context) (model.Appointment, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetAppointment Service")
	var appointment model.Appointment
	tx := initializers.DB.Preload("Pet").First(&appointment, id)
	if err := tx.Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return model.Appointment{}, AppointmentNotFoundError{AppointmentID: id}
		default:
			return model.Appointment{}, fmt.Errorf("getting appointment %d: %w", id, err)
		}
	}
	petService := &PetService{}
	_, err := petService.GetPet(appointment.PetID, ctx)
	if err != nil {
		return model.Appointment{}, fmt.Errorf("getting appointment %d: %w", id, err)
	}
	return appointment, nil
}

func (appointmentService *AppointmentService) GetAppointmentBySlot(slot time.Time) (model.Appointment, error) {
	l := zerolog.Ctx(context.Background())
	l.Trace().Msg("Inside GetAppointmentBySlot Service")
	var appointment model.Appointment
	tx := initializers.DB.Where("slot = ?", slot).First(&appointment)
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Appointment{}, AppointmentNotFoundError{AppointmentID: 0}
		}
		return model.Appointment{}, fmt.Errorf("getting appointment by slot %v: %w", slot, err)
	}
	return appointment, nil
}

func (appointmentService *AppointmentService) AddAppointment(appointment *model.Appointment, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside AddAppointment Service")

	if err := appointmentService.ValidateAppointment(appointment, ctx); err != nil {
		return fmt.Errorf("adding appointment: %w", err)
	}

	if tx := initializers.DB.Preload("Pet").Create(appointment); tx.Error != nil {
		return fmt.Errorf("adding appointment: %w", tx.Error)
	}
	return nil
}

func (appointmentService *AppointmentService) UpdateAppointment(id uint, appointment *model.Appointment, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside UpdateAppointment Service")
	existingAppointment, err := appointmentService.GetAppointment(id, ctx)
	if err != nil {
		return fmt.Errorf("updating appointment: %w", err)
	}

	if appointment.Slot != (time.Time{}) {
		existingAppointment.Slot = appointment.Slot
	}
	if appointment.Reason != "" {
		existingAppointment.Reason = appointment.Reason
	}

	if err := appointmentService.ValidateAppointment(&existingAppointment, ctx); err != nil {
		return fmt.Errorf("updating appointment: %w", err)
	}


	if tx := initializers.DB.Model(&existingAppointment).Updates(existingAppointment); tx.Error != nil {
		return fmt.Errorf("updating appointment: %w", tx.Error)
	}
	*appointment = existingAppointment
	return nil
}

func (appointmentService *AppointmentService) DeleteAppointment(id uint, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside DeleteAppointment Service")
	existingAppointment, err := appointmentService.GetAppointment(id, ctx)
	if err != nil {
		return fmt.Errorf("deleting appointment: %w", err)
	}

	if tx := initializers.DB.Delete(&existingAppointment); tx.Error != nil {
		return fmt.Errorf("deleting appointment: %w", tx.Error)
	}

	return nil
}

func (appointmentService *AppointmentService) GetUpcomingAppointments() ([]model.Appointment, error) {
	l := zerolog.Ctx(context.Background())
	l.Trace().Msg("Inside GetUpcomingAppointments Service")
	var appointments []model.Appointment
	tx := initializers.DB.Where("slot > ?", time.Now()).Preload("Pet").Order("slot ASC").Find(&appointments)
	if tx.Error != nil {
		return nil, fmt.Errorf("getting all upcoming appointments: %w", tx.Error)
	}
	
	return appointments, nil
}

func (appointmentService *AppointmentService) GetTodayAppointments() ([]model.Appointment, error) {
	l := zerolog.Ctx(context.Background())
	l.Trace().Msg("Inside GetTodayAppointments Service")
	var appointments []model.Appointment
	today := time.Now().Truncate(24 * time.Hour)
	tx := initializers.DB.Where("slot >= ? AND slot < ?", today, today.Add(24*time.Hour)).Preload("Pet").Order("slot ASC").Find(&appointments)
	if tx.Error != nil {
		return nil, fmt.Errorf("getting today's appointments: %w", tx.Error)
	}
	return appointments, nil
}

func (appointmentService *AppointmentService) GetUpcomingAppointmentsByOwner(ctx context.Context) ([]model.Appointment, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetUpcomingAppointmentsByOwner Service")
	var appointments []model.Appointment
	ownerID, ok := ctx.Value("user_id").(uint)
	if !ok {
		return nil, fmt.Errorf("getting upcoming appointments by owner: user_id not found in context")
	}
	tx := initializers.DB.Where("slot > ? AND owner_id = ?", time.Now(), ownerID).Preload("Pet").Find(&appointments)
	if tx.Error != nil {
		return nil, fmt.Errorf("getting upcoming appointments by owner %d: %w", ownerID, tx.Error)
	}
	return appointments, nil
}

func (appointmentService *AppointmentService) ValidateAppointment(appointment *model.Appointment, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside ValidateAppointment Service")
	petService := &PetService{}
	if _, err := petService.GetPet(appointment.PetID, ctx); err != nil {
		return fmt.Errorf("adding appointment: %w", err)
	}

	existingAppointment, err := appointmentService.GetAppointmentBySlot(appointment.Slot)
	if err == nil {
		if appointment.ID == existingAppointment.ID {
			return nil
		}
		return &AppointmentFoundError{AppointmentID: existingAppointment.ID}
	} else if !errors.As(err, &AppointmentNotFoundError{}) {
		return fmt.Errorf("validating appointment: %w", err)
	}

	if appointment.Slot.Before(time.Now()) {
		return fmt.Errorf("validating appointment: slot in past: %w", ErrInvalidSlot)
	}
	if appointment.Slot.Hour() < 9 || appointment.Slot.Hour() > 17 {
		return fmt.Errorf("validating appointment: slot outside working hours: %w", ErrInvalidSlot)
	}
	if appointment.Slot.Minute() != 0 && appointment.Slot.Minute() != 30 {
		return fmt.Errorf("validating appointment: slot not on the hour or half-hour: %w", ErrInvalidSlot)
	}
	return nil
}
