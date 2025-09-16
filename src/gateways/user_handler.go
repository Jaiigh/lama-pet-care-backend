package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) FindUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	var user *entities.UserDataModel
	switch token.Role {
	case "caretaker":
		user, err = h.CaretakerService.FindCaretakerByID(token.UserID)
	case "doctor":
		user, err = h.DoctorService.FindDoctorByID(token.UserID)
	case "owner":
		user, err = h.OwnerService.FindOwnerByID(token.UserID)
	case "admin":
		user, err = h.AdminService.FindAdminByID(token.UserID)
	default:
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "invalid role",
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":  "user not found",
			"detail": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

func (h *HTTPGateway) DeleteUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	var deletedUser *entities.UserDataModel
	switch token.Role {
	case "caretaker":
		deletedUser, err = h.CaretakerService.DeleteCaretakerByID(token.UserID)
	case "doctor":
		deletedUser, err = h.DoctorService.DeleteDoctorByID(token.UserID)
	case "owner":
		deletedUser, err = h.OwnerService.DeleteOwnerByID(token.UserID)
	case "admin":
		deletedUser, err = h.AdminService.DeleteAdminByID(token.UserID)
	default:
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "invalid role",
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "caretaker deleted successfully",
		"data":    deletedUser,
	})
}

func (h *HTTPGateway) UpdateUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	var updateData entities.UpdateUserModel
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	var updatedUser *entities.UserDataModel
	switch token.Role {
	case "caretaker":
		updatedUser, err = h.CaretakerService.UpdateCaretakerByID(token.UserID, updateData)
	case "doctor":
		updatedUser, err = h.DoctorService.UpdateDoctorByID(token.UserID, updateData)
	case "owner":
		updatedUser, err = h.OwnerService.UpdateOwnerByID(token.UserID, updateData)
	case "admin":
		updatedUser, err = h.AdminService.UpdateAdminByID(token.UserID, updateData)
	default:
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "invalid role",
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}
