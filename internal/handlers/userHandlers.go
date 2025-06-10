package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MSaiAswin/pet-clinic-management-system/internal/middleware"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func (h *handlerService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside LoginHandler")
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Str("username", body.Username).Msg("User login attempt")
	token, err := h.userService.Login(body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidUserInput) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			h.respond(w, service.ErrInvalidCredentials, http.StatusNotFound)
			return
		}
		l.Error().Err(err).Msg("Failed to login user")
		h.respond(w, err, http.StatusInternalServerError)
	}
	l.Info().Str("username", body.Username).Msg("User login successful")
	h.respond(w, map[string]string{"token": token}, http.StatusOK)
}

func (h *handlerService) SignupHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside SignupHandler")
	l.Info().Msg("Processing user signup request")
	body := &service.UserSignupParams{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Str("username", body.Username).Msg("User signup attempt")
	token, err := h.userService.Signup(body)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.respond(w, errors.New("username already exists"), http.StatusBadRequest)
			return
		} else if errors.Is(err, service.ErrInvalidUserInput) {
			h.respond(w, err, http.StatusBadRequest)
			return
		}
		l.Error().Err(err).Msg("Failed to create user")
		h.respond(w, err, http.StatusInternalServerError)
	}
	l.Info().Str("username", body.Username).Msg("User signup successful")
	h.respond(w, map[string]string{"token": token}, http.StatusCreated)
}

func (h *handlerService) GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetUserByIDHandler")
	userID := r.Context().Value(middleware.ContextKeyUserID).(uint)
	l.Info().Msgf("Incoming request to fetch user by ID: %d", userID)
	user, err := h.userService.GetUser(userID, r.Context())
	if err != nil {
		if errors.As(err, &service.UserNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch user")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}

	l.Debug().Interface("user", user).Send()
	l.Info().Msgf("User with ID %d fetched successfully", userID)

	data := user
	h.respond(w, data, http.StatusOK)
}

func (h *handlerService) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside UpdateUserHandler")
	userID := r.Context().Value(middleware.ContextKeyUserID).(uint)
	l.Info().Msgf("Incoming request to update user with ID: %d", userID)
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdateUser(userID, &user, r.Context()); err != nil {
		if errors.As(err, &service.UserNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.Is(err, service.ErrInvalidUserInput) {
			h.respond(w, errors.New("username already exists"), http.StatusBadRequest)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to update user")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}

	l.Debug().Interface("user", user).Send()
	l.Info().Msgf("User with ID %d updated successfully", userID)

	h.respond(w, user, http.StatusOK)
}

func (h *handlerService) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside DeleteUserHandler")
	userID := r.Context().Value(middleware.ContextKeyUserID).(uint)
	l.Info().Msgf("Incoming request to delete user with ID: %d", userID)
	if err := h.userService.DeleteUser(userID, r.Context()); err != nil {
		if errors.As(err, &service.UserNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to delete user")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Msgf("User with ID %d deleted successfully", userID)
	h.respond(w, nil, http.StatusNoContent)
}
