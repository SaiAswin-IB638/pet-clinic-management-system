package validators

import (
	"context"

	"github.com/MSaiAswin/pet-clinic-management-system/internal/middleware"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
)

type ResourceNotOwnedError struct {
}

func (e ResourceNotOwnedError) Error() string {
	return "requested resource is not owned by the user"
}

func ValidateResourceOwner(resourceOwnerID uint, ctx context.Context) error {
	userID, _ := ctx.Value(middleware.ContextKeyUserID).(uint)
	role, _ := ctx.Value(middleware.ContextKeyRole).(string)
	if userID != resourceOwnerID && role == model.UserTypeOwner {
		return ResourceNotOwnedError{}
	}
	return nil
}