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

	"github.com/MSaiAswin/pet-clinic-management-system/internal/model"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/service"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/validators"
	"github.com/gorilla/mux"
)

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

func (h *handlerService) CreatePetHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside CreatePetHandler")
	l.Info().Msg("Incoming request to create a new pet")
	var pet model.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
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

	var pet model.Pet
	if err := json.NewDecoder(r.Body).Decode(&pet); err != nil {
		h.respond(w, err, http.StatusBadRequest)
		return
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

func (h *handlerService) GetAllPetsHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetAllPetsHandler")
	l.Info().Msg("Incoming request to fetch all pets")
	pets, err := h.petService.GetAllPets(r.Context())
	if err != nil {
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	h.respond(w, pets, http.StatusOK)
}

func (h *handlerService) GetPetsByOwnerHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetsByOwnerHandler")
	l.Info().Msg("Incoming request to fetch pets by owner")
	pets, err := h.petService.GetPetsByOwner(r.Context())
	if err != nil {
		h.respond(w, err, http.StatusInternalServerError)
		return
	}
	h.respond(w, pets, http.StatusOK)
}

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

	l.Info().Uint("petID", petID).Str("fileName", fileName).Msg("Pet document uploaded successfully")
	h.respond(w, map[string]string{"message": "Pet document uploaded successfully", "fileName": fileName + fileExension}, http.StatusCreated)

}

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

func (h *handlerService) GetPetDocumentByNameHandler(w http.ResponseWriter, r *http.Request) {
	l := zerolog.Ctx(r.Context())
	l.Trace().Msg("Inside GetPetDocumentByIDHandler")
	vars := mux.Vars(r)
	petID, err := h.petIDValidate(&vars)
	if err != nil {
		h.respond(w, err, http.StatusBadRequest)
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
