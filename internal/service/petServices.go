package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/initializers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/middleware"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type PetNotFoundError struct {
	ID uint
}

func (e PetNotFoundError) Error() string {
	return fmt.Sprintf("pet with ID %d not found", e.ID)
}

func (perService *PetService) GetPet(id uint, ctx context.Context) (model.Pet, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetPet Service")
	var pet model.Pet
	tx := initializers.DB.First(&pet, id)
	if err := tx.Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return model.Pet{}, PetNotFoundError{ID: id}
		default:
			return model.Pet{}, fmt.Errorf("getting pet %d: %w", id, err)
		}
	}

	err := validators.ValidateResourceOwner(pet.OwnerID, ctx)
	if err != nil {
		return model.Pet{}, validators.ResourceNotOwnedError{}
	}

	return pet, nil

}

func (perService *PetService) AddPet(pet *model.Pet, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside AddPet Service")
	ownerID, _ := ctx.Value(middleware.ContextKeyUserID).(uint)
	pet.OwnerID = ownerID
	tx := initializers.DB.Create(pet)
	if err := tx.Error; err != nil {
		return fmt.Errorf("adding pet: %w", err)
	}
	return nil
}

func (perService *PetService) UpdatePet(id uint, pet *model.Pet, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside UpdatePet Service")
	existingPet, err := perService.GetPet(id, ctx)
	if err != nil {
		return fmt.Errorf("updating pet %d: %w", id, err)
	}

	if pet.Name != "" {
		existingPet.Name = pet.Name
	}
	if pet.Species != "" {
		existingPet.Species = pet.Species
	}
	if pet.Breed != "" {
		existingPet.Breed = pet.Breed
	}
	if pet.OwnerID != 0 {
		userService := &UserService{}
		if _, err := userService.GetUser(pet.OwnerID, ctx); err != nil {
			return fmt.Errorf("updating pet %d: %w", id, err)
		}
		existingPet.OwnerID = pet.OwnerID
	}
	if pet.MedicalHistory != "" {
		existingPet.MedicalHistory = pet.MedicalHistory
	}
	tx := initializers.DB.Model(&existingPet).Updates(existingPet)
	if err := tx.Error; err != nil {
		return fmt.Errorf("updating pet %d: %w", id, err)
	}
	*pet = existingPet
	return nil
}

func (perService *PetService) DeletePet(id uint, ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside DeletePet Service")
	pet, err := perService.GetPet(id, ctx)
	if err != nil {
		return fmt.Errorf("deleting pet %d: %w", id, err)
	}

	if err := initializers.DB.Delete(&pet).Error; err != nil {
		return fmt.Errorf("deleting pet %d: %w", id, err)
	}
	return nil
}

func (perService *PetService) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetAllPets Service")
	var pets []model.Pet
	tx := initializers.DB.Find(&pets)
	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("getting all pets: %w", err)
	}

	return pets, nil
}

func (perService *PetService) GetPetsByOwner(ctx context.Context) ([]model.Pet, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetPetsByOwner Service")
	var pets []model.Pet
	tx := initializers.DB.Where("owner_id = ?", ctx.Value(middleware.ContextKeyUserID)).Find(&pets)
	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("getting pets by owner: %w", err)
	}

	return pets, nil
}


// pet documents are stored on the server as files itself, so we can use the pet ID to fetch them
// stored path is "/uploads/pets/{petID}/{documentName}"
// and the document ID should be an xid
func (perService *PetService) GetPetDocuments(petID uint, ctx context.Context) ([]string, error) {
	l := zerolog.Ctx(ctx)
	l.Trace().Msg("Inside GetPetDocuments Service")
	documentPath := filepath.Join("uploads", "pets", fmt.Sprint(petID))
	l.Debug().Str("documentPath", documentPath).Uint("petID", petID).Msg("Fetching pet documents from path")
	// Check if the directory exists
	if _, err := os.Stat(documentPath); os.IsNotExist(err) {
		l.Debug().Uint("petID", petID).Msg("No documents found for pet")
		return []string{}, nil // No documents found
	}
	files, err := os.ReadDir(documentPath)
	if err != nil {
		l.Error().Err(err).Uint("petID", petID).Msg("Failed to read pet documents directory")
		return nil, fmt.Errorf("reading pet documents for pet %d: %w", petID, err)
	}
	documents := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		documents = append(documents, file.Name())
	}
	return documents, nil
}

