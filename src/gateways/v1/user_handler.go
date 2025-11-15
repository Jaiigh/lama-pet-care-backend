package gateways

import (
	"lama-backend/domain/entities"
	"lama-backend/src/middlewares"
	"lama-backend/src/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// @Summary find user by id
// @Description find user by id and role from JWT token
// @Tags user
// @Produce json
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 404 {object} string "User not found."
// @Router /user/ [get]
// @Security BearerAuth
func (h *HTTPGateway) FindUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
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
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 500 {object} string "Internal Server Error"
// @Router /user/ [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeleteUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	deletedUser, err := h.UsersService.DeleteUsersByID(token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found, details: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "user deleted successfully",
		Data:    deletedUser,
		Status:  fiber.StatusOK,
	})
}

// @Summary delete user by admin
// @Description admin delete user by specifying user ID
// @Tags user
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "Invalid user ID"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 500 {object} entities.ResponseMessage "Internal Server Error"
// @Router /admin/users/{userID} [delete]
// @Security BearerAuth
func (h *HTTPGateway) DeleteUserByAdmin(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	userID := ctx.Params("userID")
	if userID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid user ID"})
	}

	deletedUser, err := h.UsersService.DeleteUsersByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
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
// @Accept json
// @Produce json
// @Param body body entities.UpdateUserModel true "update user data"
// @Success 200 {object} entities.UserDataModel "Request successful"
// @Failure 400 {object} string "Invalid json body"
// @Failure 401 {object} string "Unauthorization Token."
// @Failure 500 {object} string "Internal Server Error"
// @Router /user/ [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateUserByID(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	var updateData entities.UpdateUserModel
	if err := ctx.BodyParser(&updateData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if err := validator.New().Struct(updateData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: utils.FormatValidationError(err)})
	}

	updatedUser, err := h.UsersService.UpdateUsersByID(token.UserID, updateData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}

// @Summary update user picture
// @Description update user picture by id and role from JWT token
// @Tags user
// @Accept multipart/form-data
// @Produce json
// @Param profile formData file true "user profile picture"
// @Success 200 {object} entities.UserDataModel "Request successful"
// @Failure 400 {object} entities.ResponseMessage "picture file is required"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 500 {object} entities.ResponseMessage "Internal Server Error"
// @Router /user/profile [patch]
// @Security BearerAuth
func (h *HTTPGateway) UpdateUserPicture(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}

	file, err := ctx.FormFile("profile")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "picture file is required",
		})
	}

	pictureString, err := utils.UploadToSupabase(file, token.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "failed to upload picture: " + err.Error(),
		})
	}

	var updateData entities.UpdateUserModel
	updateData.Profile = &pictureString

	updatedUser, err := h.UsersService.UpdateUsersByID(token.UserID, updateData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(updatedUser)
}

// @Summary list users
// @Description Admin-only endpoint that returns paginated users with an optional role filter.
// @Tags user
// @Produce json
// @Param role query string false "Filter by role (admin, owner, caretaker, doctor)"
// @Param page query int false "Page number for pagination" [optional default: 1]
// @Param limit query int false "Number of users per page" [optional default: 20]
// @Success 200 {object} entities.ResponseModel "Request successful"
// @Failure 401 {object} entities.ResponseMessage "Unauthorization Token."
// @Failure 403 {object} entities.ResponseMessage "Invalid role"
// @Failure 500 {object} entities.ResponseMessage "Internal server error"
// @Router /admin/users [get]
// @Security BearerAuth
func (h *HTTPGateway) GetAllUsers(ctx *fiber.Ctx) error {
	token, err := middlewares.DecodeJWTToken(ctx)
	if err != nil || token.Purpose != "access" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
	}
	if token.Role != "admin" {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: "Invalid role"})
	}

	role := ctx.Query("role")
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 20)

	users, err := h.UsersService.FindAllUsers(role, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data: fiber.Map{
			"page":   page,
			"limit":  limit,
			"amount": len(users),
			"users":  users,
		},
		Status: fiber.StatusOK,
	})
}
