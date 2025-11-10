package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary create pet
// @Description owner can create a pet (only role == "owner")
// @Tags pet
// @Accept json
// @Produce json
// @Param body body entities.CreatedPetModel true "pet payload"
// @Success 201 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid json body"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /pets [post]
// @Security BearerAuth
func (h *HTTPGateway) CreatePet(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	// Only owner can create pet (OID FK constraint will be one's own ID)
	if token.Role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	var pet entities.CreatedPetModel
	if err := ctx.BodyParser(&pet); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	// Ensure pet is created under the authenticated owner
	pet.OwnerID = token.UserID

	if err := validator.New().Struct(pet); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	created, err := h.PetService.InsertPet(pet)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot create pet: " + err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseModel{
		Message: "pet created",
		Data:    created,
		Status:  fiber.StatusCreated,
	})
}

// @Summary get owner's pets
// @Description owner can get their own pets. This endpoint is owner-only and the owner ID is
// taken from the JWT token (do not provide owner ID in path or body).
// @Tags pet
// @Produce json
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid owner ID"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /pets/owner [get]
// @Security BearerAuth
func (h *HTTPGateway) FindByOwnerID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	// This endpoint is owner-only: ownerID must come from the JWT token.
	if token.Role != "owner" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	ownerID := token.UserID
	if ownerID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid owner ID"})
	}

	pets, err := h.PetService.FindByOwnerID(ownerID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"amount": len(pets),
			"pets":   pets,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary get all pets
// @Description admin can fetch all pets
// @Tags pet
// @Produce json
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /pets [get]
// @Security BearerAuth
func (h *HTTPGateway) FindAllPets(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	// Only admin allowed to fetch all pets
	if token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	pets, err := h.PetService.FindAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"amount": len(pets),
			"pets":   pets,
		},
		Status: fiber.StatusOK,
	})
}

// @Summary update pet
// @Description owner or admin can update a pet. If role is owner, the pet must belong to the owner.
// @Tags pet
// @Accept json
// @Produce json
// @Param petID path string true "pet id"
// @Param body body entities.UpdatePetModel true "pet payload"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid pet ID or no fields to update"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role or not owner's pet"
// @Failure 404 {object} entities.ResponseMessage "pet not found"
// @Failure 422 {object} entities.ResponseMessage "Validation error"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /pets/{petID} [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdatePet(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	// only owner or admin can update pets
	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	petID := ctx.Params("petID")
	if petID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid pet ID"})
	}

	var req entities.UpdatePetModel
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	// make sure there's something to update
	if req.Breed == nil && req.Name == nil && req.BirthDate == nil && req.Weight == nil && req.Kind == nil && req.Sex == nil && req.OwnerID == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "no fields to update"})
	}

	if err := validator.New().Struct(req); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	// If owner, ensure they own the pet
	if token.Role == "owner" {
		ownerPets, err := h.PetService.FindByOwnerID(token.UserID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
		}
		owned := false
		for _, p := range ownerPets {
			if p.PetID == petID {
				owned = true
				break
			}
		}
		if !owned {
			return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "You do not own this pet"})
		}
	}

	updated, err := h.PetService.UpdatePet(petID, req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "pet not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "pet updated",
		Data:    updated,
		Status:  fiber.StatusOK,
	})
}

// @Summary delete pet
// @Description owner or admin can delete a pet. If role is owner, the pet must belong to the owner.
// @Tags pet
// @Produce json
// @Param petID path string true "pet id"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid pet ID"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role or not owner's pet"
// @Failure 404 {object} entities.ResponseMessage "pet not found"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /pets/{petID} [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeletePet(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	if token.Role != "owner" && token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	petID := ctx.Params("petID")
	if petID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid pet ID"})
	}

	// If owner role, ensure they own the pet
	if token.Role == "owner" {
		ownerPets, err := h.PetService.FindByOwnerID(token.UserID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
		}
		owned := false
		for _, p := range ownerPets {
			if p.PetID == petID {
				owned = true
				break
			}
		}
		if !owned {
			return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "You do not own this pet"})
		}
	}

	deleted, err := h.PetService.DeletePet(petID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "pet not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "pet deleted",
		Data:    deleted,
		Status:  fiber.StatusOK,
	})
}
