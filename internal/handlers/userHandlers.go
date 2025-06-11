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

	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginSuccessResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ..."`
}

type UpdateUserRequest struct {
	Name    string `json:"name" example:"John Doe"`
	Email   string `json:"email" example:"john@doe.com"`
	Contact string `json:"contact" example:"1234567890"`
	Username string `json:"username" example:"johndoe"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// LoginHandler godoc
// @Summary User Login
// @Description Logs in a user with username and password.
// @Tags User
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login request body"
// @Success 200 {object} LoginSuccessResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /login [post]
func (h *handlerService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside LoginHandler")
	var body LoginRequest
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

// SignupHandler godoc
// @Summary User Signup
// @Description Registers a new user with name, username and password.
// @Tags User
// @Accept json
// @Produce json
// @Param body body service.UserSignupParams true "Signup request body"
// @Success 201 {object} LoginSuccessResponse "Signup successful"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /signup [post]
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

// GetUserByIDHandler godoc
// @Summary Get User by ID
// @Description Fetches user details by user ID.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.User "User details"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /owners [get]
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

// UpdateUserHandler godoc
// @Summary Update User
// @Description Updates user details by user ID.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateUserRequest true "Update user request body"
// @Success 200 {object} model.User "User updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /owners [put]
func (h *handlerService) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside UpdateUserHandler")
	userID := r.Context().Value(middleware.ContextKeyUserID).(uint)
	l.Info().Msgf("Incoming request to update user with ID: %d", userID)
	var userParams UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}

	user := model.User{
		Name:	 userParams.Name,
		Email:    userParams.Email,
		Contact:  userParams.Contact,
		Username: userParams.Username,
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

// DeleteUserHandler godoc
// @Summary Delete User
// @Description Deletes user by user ID.
// @Tags User
// @Security BearerAuth
// @Success 204 "User deleted successfully"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /owners [delete]
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
