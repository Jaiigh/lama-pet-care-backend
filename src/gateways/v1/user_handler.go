package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

// @Summary find user by id
// @Description find user by id and role from JWT token
// @Tags user
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 403 {object} string "Invalid role"
// @Failure 404 {object} string "User not found."
// @Router /user/ [get]
// @Security BearerAuth
func (h *HTTPGateway) FindUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	user, err := h.UsersService.FindUsersByID(token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found - details: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "user found",
		Data:    user,
		Status:  fiber.StatusOK,
	})
}

// @Summary delete user by id
// @Description delete user by id and role from JWT token
// @Tags user
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 403 {object} string "Invalid role"
// @Failure 500 {object} string "Internal Server Error"
// @Router /user/ [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeleteUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	deletedUser, err := h.UsersService.DeleteUsersByID(token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "user deleted successfully",
		Data:    deletedUser,
		Status:  fiber.StatusOK,
	})
}

// @Summary delete user by id
// @Description delete user by id and role from JWT token
// @Tags user
// @Produce json
// @Param Authorization header string true "Bearer <JWT token>"
// @Param body body entities.UpdateUserModel true "update user data"
// @Success 200 {object} entities.UserDataModel "Request successful"
// @Failure 400 {object} string "Invalid json body"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 403 {object} string "Invalid role"
// @Failure 500 {object} string "Internal Server Error"
// @Router /user/ [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	var updateData entities.UpdateUserModel
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updatedUser, err := h.UsersService.UpdateUsersByID(token.UserID, updateData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}
