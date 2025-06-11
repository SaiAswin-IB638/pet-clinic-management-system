package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type AppointmentParams struct {
	Slot   time.Time `json:"slot" example:"2023-10-01T10:00:00Z"`
	Reason string    `json:"reason" example:"Regular checkup"`
	PetID  uint      `json:"pet_id" example:"1"`
}

// GetAppointmentByIDHandler godoc
// @Summary Get Appointment by ID
// @Description Fetches an appointment by its ID.
// @Tags Appointment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Appointment ID"
// @Success 200 {object} model.Appointment "Appointment details"
// @Failure 400 {object} ErrorResponse "Invalid appointment ID"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments/{id} [get]
func (h *handlerService) GetAppointmentByIDHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetAppointmentByIDHandler")
	l.Info().Msg("Fetching appointment by ID")
	vars := mux.Vars(r)
	appointmentID, err := h.appointmentIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Uint("appointmentID", appointmentID).Msg("Fetching appointment by ID")
	appointment, err := h.appointmentService.GetAppointment(appointmentID, r.Context())
	if err != nil {
		if errors.As(err, &service.AppointmentNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch appointment by ID")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Debug().Interface("appointment", appointment).Send()
	l.Info().Uint("appointmentID", appointmentID).Msg("Appointment fetched successfully")
	h.respond(w, appointment, http.StatusOK)
}

// CreateAppointmentHandler godoc
// @Summary Create Appointment
// @Description Creates a new appointment.
// @Tags Appointment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body AppointmentParams true "Appointment parameters"
// @Success 201 {object} model.Appointment "Appointment created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /appointments [post]

func (h *handlerService) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside CreateAppointmentHandler")
	l.Info().Msg("Incoming request to create a new appointment")
	var appointmentParams AppointmentParams
	if err := json.NewDecoder(r.Body).Decode(&appointmentParams); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	appointment := model.Appointment{
		Slot:   appointmentParams.Slot,
		Reason: appointmentParams.Reason,
		PetID:  appointmentParams.PetID,
	}
	if err := h.appointmentService.AddAppointment(&appointment, r.Context()); err != nil {
		if errors.As(err, &service.AppointmentFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidSlot) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to create appointment")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Msg("Appointment created successfully")
	l.Debug().Interface("appointment", appointment).Msg("Created appointment data")
	h.respond(w, appointment, http.StatusCreated)
}

// UpdateAppointmentHandler godoc
// @Summary Update Appointment
// @Description Updates an existing appointment by its ID.
// @Tags Appointment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Appointment ID"
// @Param body body AppointmentParams true "Appointment parameters"
// @Success 200 {object} model.Appointment "Appointment updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /appointments/{id} [put]
func (h *handlerService) UpdateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside UpdateAppointmentHandler")
	l.Info().Msg("Incoming request to update appointment")
	vars := mux.Vars(r)
	appointmentID, err := h.appointmentIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Uint("appointmentID", appointmentID).Msg("Updating appointment by ID")
	var appointmentParams AppointmentParams
	if err := json.NewDecoder(r.Body).Decode(&appointmentParams); err != nil {
		h.respond(w, err.Error(), http.StatusBadRequest)
		return
	}
	appointment := model.Appointment{
		Slot:   appointmentParams.Slot,
		Reason: appointmentParams.Reason,
		PetID:  appointmentParams.PetID,
	}
	if err := h.appointmentService.UpdateAppointment(appointmentID, &appointment, r.Context()); err != nil {
		if errors.As(err, &service.AppointmentNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &service.AppointmentFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidSlot) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to update appointment")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Uint("appointmentID", appointmentID).Msg("Appointment updated successfully")
	l.Debug().Interface("appointment", appointment).Msg("Updated appointment data")
	h.respond(w, appointment, http.StatusOK)
}

// DeleteAppointmentHandler godoc
// @Summary Delete Appointment
// @Description Deletes an appointment by its ID.
// @Tags Appointment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Appointment ID"
// @Success 204 "Appointment deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid appointment ID"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /appointments/{id} [delete]
func (h *handlerService) DeleteAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside DeleteAppointmentHandler")
	l.Info().Msg("Incoming request to delete appointment")
	vars := mux.Vars(r)
	appointmentID, err := h.appointmentIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Uint("appointmentID", appointmentID).Msg("Deleting appointment by ID")
	if err := h.appointmentService.DeleteAppointment(appointmentID, r.Context()); err != nil {
		if errors.As(err, &service.AppointmentNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to delete appointment")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Uint("appointmentID", appointmentID).Msg("Appointment deleted successfully")
	h.respond(w, nil, http.StatusNoContent)
}

// GetUpcomingAppointmentsHandler godoc
// @Summary Get Upcoming Appointments
// @Description Fetches all upcoming appointments.
// @Description This endpoint is restricted to staff users.
// @Tags Appointment
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Appointment "List of upcoming appointments"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /staff/appointments/upcoming [get]
func (h *handlerService) GetUpcomingAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetUpcomingAppointmentsHandler")
	l.Info().Msg("Fetching upcoming appointments")
	appointments, err := h.appointmentService.GetUpcomingAppointments()
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch upcoming appointments")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Debug().Interface("appointments", appointments).Msg("Fetched upcoming appointments")
	l.Info().Msg("Upcoming appointments fetched successfully")
	h.respond(w, appointments, http.StatusOK)
}

func (h *handlerService) GetTodayAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetTodayAppointmentsHandler")
	l.Info().Msg("Fetching today's appointments")
	appointments, err := h.appointmentService.GetTodayAppointments()
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch today's appointments")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Debug().Interface("appointments", appointments).Msg("Fetched today's appointments")
	l.Info().Msg("Today's appointments fetched successfully")
	h.respond(w, appointments, http.StatusOK)
}

// GetUpcomingAppointmentsByOwnerHandler godoc
// @Summary Get Upcoming Appointments by Owner
// @Description Fetches all upcoming appointments for the authenticated owner.
// @Description This endpoint is restricted to staff users.
// @Tags Appointment
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Appointment "List of upcoming appointments for owner"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /staff/appointments/today [get]
func (h *handlerService) GetUpcomingAppointmentsByOwnerHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetUpcomingAppointmentsByOwnerHandler")
	l.Info().Msg("Fetching upcoming appointments for owner")
	appointments, err := h.appointmentService.GetUpcomingAppointmentsByOwner(r.Context())
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch upcoming appointments for owner")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Debug().Interface("appointments", appointments).Msg("Fetched upcoming appointments for owner")
	l.Info().Msg("Upcoming appointments for owner fetched successfully")
	h.respond(w, appointments, http.StatusOK)
}
