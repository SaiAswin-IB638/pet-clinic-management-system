package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/gorilla/mux"
)

type CreatePetRequest struct {
	Name           string `json:"name" example:"Buddy"`
	Species        string `json:"species" example:"Dog"`
	Breed          string `json:"breed" example:"Golden Retriever"`
	MedicalHistory string `json:"medical_history" example:"Healthy"`
}

type UploadPetDocumentResponse struct {
	Message string `json:"message" example:"Pet document uploaded successfully"`
	FileName string `json:"file_name" example:"document.pdf"`
}

// GetPetByIDHandler godoc
// @Summary Get Pet by ID
// @Description Fetches a pet by its ID.
// @Tags Pet
// @Produce json
// Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {object} model.Pet "Pet fetched successfully"
// @Failure 400 {object} ErrorResponse "Invalid Pet ID"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets/{id} [get]
func (h *handlerService) GetPetByIDHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetByIDHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Debug().Uint("petID", petID).Msg("Fetching pet by ID")
	pet, err := h.petService.GetPet(petID, r.Context())
	if err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch pet by ID")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Debug().Interface("pet", pet).Send()
	l.Info().Uint("petID", petID).Msg("Pet fetched successfully")
	h.respond(w, pet, http.StatusOK)
}

// CreatePetHandler godoc
// @Summary Create a new Pet
// @Description Creates a new pet with the provided details.
// @Tags Pet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreatePetRequest true "Create pet request body"
// @Success 201 {object} model.Pet "Pet created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets [post]
func (h *handlerService) CreatePetHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside CreatePetHandler")
	l.Info().Msg("Incoming request to create a new pet")
	var petParams CreatePetRequest
	if err := json.NewDecoder(r.Body).Decode(&petParams); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}

	pet := model.Pet{
		Name:           petParams.Name,
		Species:        petParams.Species,
		Breed:          petParams.Breed,
		MedicalHistory: petParams.MedicalHistory,
	}

	if pet.Name == "" || pet.Species == "" || pet.Breed == "" {
		err := errors.New("name, species, and breed are required fields")
		h.respond(w, err, http.StatusBadRequest)
	}

	l.Debug().Interface("pet", pet).Msg("Decoded pet data from request body")

	if err := h.petService.AddPet(&pet, r.Context()); err != nil {
		if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to create pet")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Msg("Pet created successfully")
	l.Debug().Interface("pet", pet).Msg("Created pet data")
	h.respond(w, pet, http.StatusCreated)
}

// UpdatePetHandler godoc
// @Summary Update Pet
// @Description Updates an existing pet by its ID.
// @Tags Pet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param body body CreatePetRequest true "Update pet request body"
// @Success 200 {object} model.Pet "Pet updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets/{id} [put]
func (h *handlerService) UpdatePetHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside UpdatePetHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Info().Uint("petID", petID).Msg("Incoming request to update pet")

	var petParams CreatePetRequest
	if err := json.NewDecoder(r.Body).Decode(&petParams); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	pet := model.Pet{
		Name:           petParams.Name,
		Species:        petParams.Species,
		Breed:          petParams.Breed,
		MedicalHistory: petParams.MedicalHistory,
	}

	l.Debug().Uint("petID", petID).Interface("pet", pet).Msg("Decoded pet data for update")

	if err := h.petService.UpdatePet(petID, &pet, r.Context()); err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		}
		if errors.As(err, &service.UserNotFoundError{}) {
			h.respond(w, err, http.StatusBadRequest)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to update pet")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Uint("petID", petID).Msg("Pet updated successfully")
	l.Debug().Interface("pet", pet).Msg("Updated pet data")
	h.respond(w, pet, http.StatusOK)
}

// DeletePetHandler godoc
// @Summary Delete Pet
// @Description Deletes a pet by its ID.
// @Tags Pet
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 204 "Pet deleted successfully"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets/{id} [delete]
func (h *handlerService) DeletePetHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside DeletePetHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Info().Uint("petID", petID).Msg("Incoming request to delete pet")
	if err := h.petService.DeletePet(petID, r.Context()); err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to delete pet")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Uint("petID", petID).Msg("Pet deleted successfully")
	h.respond(w, nil, http.StatusNoContent)
}

// GetAllPetsHandler godoc
// @Summary Get All Pets
// @Description Fetches all pets.
// @Description This endpoint is restricted to staff users only.
// @Tags Pet
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Pet "List of pets"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /staff/pets [get]
func (h *handlerService) GetAllPetsHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetAllPetsHandler")
	l.Info().Msg("Incoming request to fetch all pets")
	pets, err := h.petService.GetAllPets(r.Context())
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch all pets")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	h.respond(w, pets, http.StatusOK)
}

// GetPetsByOwnerHandler godoc
// @Summary Get Pets by Owner
// @Description Fetches all pets owned by the authenticated user.
// @Tags Pet
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Pet "List of pets owned by the user"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets [get]
func (h *handlerService) GetPetsByOwnerHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetsByOwnerHandler")
	l.Info().Msg("Incoming request to fetch pets by owner")
	pets, err := h.petService.GetPetsByOwner(r.Context())
	if err != nil {
		l.Error().Err(err).Msg("Failed to fetch pets by owner")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	h.respond(w, pets, http.StatusOK)
}

// UploadPetDocumentHandler godoc
// @Summary Upload Pet Document
// @Description Uploads a document for a specific pet.
// @Description This endpoint is restricted to staff users only.
// @Tags Pet
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param file formData file true "File to upload"
// @Param name formData string true "File name"
// @Success 201 {object} UploadPetDocumentResponse "Pet document uploaded successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /staff/pets/{id}/upload [post]
func (h *handlerService) UploadPetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside UploadPetDocumentHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Info().Uint("petID", petID).Msg("Incoming request to upload pet document")

	// check if the pet exists
	if _, err := h.petService.GetPet(petID, r.Context()); err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch pet for document upload")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		l.Error().Err(err).Msg("Failed to parse multipart form")
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	fileName := r.Form.Get("name")
	if fileName == "" {
		err := errors.New("file name is required")
		l.Error().Err(err).Msg("File name not provided in request")
		h.respond(w, err, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		l.Error().Err(err).Msg("Failed to get file from form")
		h.respond(w, err, http.StatusBadRequest)
		return
	}

	defer file.Close()

	l.Debug().Str("fileName", fileName).Msg("File name received for upload")

	fileExension := strings.ToLower(filepath.Ext(handler.Filename))

	path := filepath.Join("uploads", "pets")
	if err := os.MkdirAll(filepath.Join(path, strconv.Itoa(int(petID))), os.ModePerm); err != nil {
		l.Error().Err(err).Msg("Failed to create pet document directory")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	filePath := filepath.Join(path, strconv.Itoa(int(petID)), fileName+fileExension)

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		l.Error().Err(err).Msg("Failed to open file for writing")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		l.Error().Err(err).Msg("Failed to copy file content")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}

	response := UploadPetDocumentResponse{
		Message: "Pet document uploaded successfully",
		FileName: fileName + fileExension,
	}

	l.Info().Uint("petID", petID).Str("fileName", fileName).Msg("Pet document uploaded successfully")
	h.respond(w, response, http.StatusCreated)

}

// GetPetDocumentsHandler godoc
// @Summary Get Pet Documents
// @Description Fetches all documents for a specific pet.
// @Tags Pet
// @Produce json
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Success 200 {array} string "List of pet document names"
// @Failure 400 {object} ErrorResponse "Invalid Pet ID"
// @Failure 404 {object} ErrorResponse "Pet not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets/{id}/documents [get]
func (h *handlerService) GetPetDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetDocumentsHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	l.Info().Uint("petID", petID).Msg("Incoming request to fetch pet documents")
	// check if the pet exists
	if _, err := h.petService.GetPet(petID, r.Context()); err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch pet for document retrieval")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	documents, err := h.petService.GetPetDocuments(petID, r.Context())
	if err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch pet documents")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	l.Info().Uint("petID", petID).Msg("Pet documents fetched successfully")
	h.respond(w, documents, http.StatusOK)
}

// GetPetDocumentByNameHandler godoc
// @Summary Get Pet Document by Name
// @Description Fetches a specific document for a pet by its name.
// @Tags Pet
// @Produce octet-stream
// @Security BearerAuth
// @Param id path int true "Pet ID"
// @Param docName path string true "Document name"
// @Success 200 {string} binary "Pet document file"
// @Failure 400 {object} ErrorResponse "Invalid Pet ID or Document Name"
// @Failure 404 {object} ErrorResponse "Pet document not found"
// @Failure 403 {object} ErrorResponse "Resource not owned"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /pets/{id}/documents/{docName} [get]
func (h *handlerService) GetPetDocumentByNameHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetDocumentByIDHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
	}
	if _, err := h.petService.GetPet(petID, r.Context()); err != nil {
		if errors.As(err, &service.PetNotFoundError{}) {
			h.respond(w, err, http.StatusNotFound)
			return
		} else if errors.As(err, &validators.ResourceNotOwnedError{}) {
			h.respond(w, err, http.StatusForbidden)
			return
		}
		l.Error().Err(err).Msg("Failed to fetch pet for document retrieval")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	docName := vars["docName"]
	l.Info().Uint("petID", petID).Str("docName", docName).Msg("Incoming request to fetch pet document by ID")

	file := filepath.Join("uploads", "pets", strconv.Itoa(int(petID)), docName)
	l.Debug().Str("file", file).Msg("Fetching pet document from file system")
	_, err = os.Open(file)

	if err != nil {
		if os.IsNotExist(err) {
			h.respond(w, errors.New("pet document not found"), http.StatusNotFound)
			return
		}
		l.Error().Err(err).Msg("Failed to open pet document")
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(file))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, file)
}
