package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

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

func (h *handlerService) CreateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside CreateAppointmentHandler")
	l.Info().Msg("Incoming request to create a new appointment")
	var appointment model.Appointment
	if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
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
	var appointment model.Appointment
	if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
		h.respond(w, err.Error(), http.StatusBadRequest)
		return
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
