package service

import (
	"context"
	"fmt"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/initializers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/utils"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"errors"
)

type UserNotFoundError struct {
	ID uint
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("user with ID %d not found", e.ID)
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidUserInput = errors.New("all fields are required")
var ErrPasswordTooShort = errors.New("password must be at least 8 characters long")

type UserSignupParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Contact  string `json:"contact"`
}

func (userService *UserService) Login(username, password string) (string, error) {
	user := &model.User{Username: username, Password: password}
	if user.Username == "" || user.Password == "" {
		return "", ErrInvalidUserInput
	}
	if tx := initializers.DB.Where("username = ?", user.Username).First(user); tx.Error != nil {
		return "", tx.Error
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", ErrInvalidCredentials
	}
	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (userService *UserService) Signup(userParams *UserSignupParams) (string, error) {
	user := &model.User{
		Name:     userParams.Name,
		Email:    userParams.Email,
		Contact:  userParams.Contact,
		Username: userParams.Username,
		Password: userParams.Password,
		Role:     model.UserTypeOwner,
	}
	if user.Username == "" || user.Password == "" || user.Name == "" || user.Email == "" || user.Contact == "" {
		return "", ErrInvalidUserInput
	}
	if len(user.Password) < 8 {
		return "", ErrPasswordTooShort
	}
	if tx := initializers.DB.Where("username = ?", user.Username).First(&model.User{}); tx.Error == nil {
		return "", ErrInvalidCredentials
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return "", err
	}
	user.Password = hashedPassword

	if err := initializers.DB.Create(user).Error; err != nil {
		return "", err
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (userService *UserService) GetUser(id uint, ctx context.Context) (model.User, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetUser Service")
	var user model.User
	tx := initializers.DB.
		Preload("Pets").
		First(&user, id)
	if err := tx.Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return model.User{}, UserNotFoundError{ID: id}
		default:
			return model.User{}, fmt.Errorf("getting user %d: %w", id, err)
		}
	}

	user.Password = ""

	if err := validators.ValidateResourceOwner(user.ID, ctx); err != nil {
		return model.User{}, fmt.Errorf("getting user %d: %w", id, err)
	}

	return user, nil
}

func (userService *UserService) UpdateUser(id uint, user *model.User, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside UpdateUser Service")
	existingUser, err := userService.GetUser(id, ctx)
	if err != nil {
		return fmt.Errorf("updating user %d: %w", id, err)
	}

	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Contact != "" {
		existingUser.Contact = user.Contact
	}
	if user.Username != "" {
		if tx := initializers.DB.Where("username = ?", user.Username).First(&model.User{}); tx.Error == nil {
			return ErrInvalidCredentials
		}
		existingUser.Username = user.Username
	}

	tx := initializers.DB.Model(&existingUser).Updates(existingUser)
	if err := tx.Error; err != nil {
		return fmt.Errorf("updating user %d: %w", id, err)
	}
	*user = existingUser
	return nil
}

func (userService *UserService) DeleteUser(id uint, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside DeleteUser Service")
	existingUser, err := userService.GetUser(id, ctx)
	if err != nil {
		return fmt.Errorf("deleting user %d: %w", id, err)
	}

	tx := initializers.DB.Delete(&existingUser)
	if err := tx.Error; err != nil {
		return fmt.Errorf("deleting user %d: %w", id, err)
	}

	return nil
}
