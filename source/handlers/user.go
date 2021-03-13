package handlers

import (
	"database/sql"
	"go-rbac/db/models"
	"go-rbac/source/dtos/request"
	"go-rbac/source/dtos/response"
	"go-rbac/source/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetUsers ...
func (h *Handler) GetUsers(c *fiber.Ctx) error {
	//query param
	email := strings.ToLower(c.Query("email"))

	var result models.UserSlice
	var err error

	if email != "" {
		result, err = models.Users(models.UserWhere.Email.EQ(email), qm.Load(models.UserRels.Role)).All(c.Context(), h.DB)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "")
		}
	} else {
		result, err = models.Users(qm.Load(models.UserRels.Role)).All(c.Context(), h.DB)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "")
		}
	}

	//users response
	var users []*response.UserResponse

	for _, user := range result {
		users = append(users, &response.UserResponse{
			UserID:    user.UserID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			Role: &response.RoleResponse{
				Name: user.R.Role.Name,
			},
		})
	}

	return utils.SendResponse(c, fiber.StatusOK, users)
}

//GetUserByID from the database
func (h *Handler) GetUserByID(c *fiber.Ctx) error {
	paramID := c.Params("id")

	//convert string to int
	id, err := strconv.Atoi(paramID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	user, err := models.Users(models.UserWhere.UserID.EQ(int64(id)), qm.Load(models.UserRels.Role)).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "user not found")
		}

		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	return utils.SendResponse(c, fiber.StatusOK, &response.UserResponse{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		Role: &response.RoleResponse{
			Name: user.R.Role.Name,
		},
	})
}

//MyProfile returns current user profile
func (h *Handler) MyProfile(c *fiber.Ctx) error {
	userID := c.Get("user")

	//convert string to int
	currentUser, _ := strconv.Atoi(userID)

	user, err := models.Users(models.UserWhere.UserID.EQ(int64(currentUser)), qm.Load(models.UserRels.Role)).One(c.Context(), h.DB)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	return utils.SendResponse(c, fiber.StatusOK, &response.UserResponse{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt,
		Role: &response.RoleResponse{
			Name: user.R.Role.Name,
		},
	})
}

//ChangePassword lets logged in user to change their password
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	//current user id
	userFromCtx := c.Get("user")
	userID, _ := strconv.Atoi(userFromCtx)

	var req request.ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//validate
	errors := utils.ValidateStruct(req)
	if errors != nil {
		return utils.SendValidationError(c, errors)
	}

	//fetch user
	user, err := models.Users(models.UserWhere.UserID.EQ(int64(userID))).One(c.Context(), h.DB)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//match user's current password with given current password
	err = utils.MatchPassword(user.Password, req.CurrentPassword)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "current password invalid")
	}

	//match user' current password with given new password
	//if password matched then return because user entered their current password as the new password
	err = utils.MatchPassword(user.Password, req.NewPassword)
	if err == nil {
		return utils.SendError(c, fiber.StatusUnprocessableEntity, "new password is same as the current password")
	}

	//hash new password
	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	//set new password as user password
	user.Password = string(newHashedPassword)

	//update password
	user.Update(c.Context(), h.DB, boil.Infer())

	return utils.SendResponse(c, fiber.StatusOK, "password successfully changed")
}
