package handlers

import (
	"database/sql"
	"go-rbac/db/models"
	"go-rbac/source/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetRoles ...
func (h *Handler) GetRoles(c *fiber.Ctx) error {
	//query param
	name := strings.ToLower(c.Query("name"))

	var roles models.RoleSlice
	var err error

	if name != "" {
		roles, err = models.Roles(models.RoleWhere.Name.EQ(name)).All(c.Context(), h.DB)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "")
		}
	} else {
		roles, err = models.Roles().All(c.Context(), h.DB)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "")
		}
	}

	return utils.SendResponse(c, fiber.StatusOK, roles)
}

//GetRoleByID ...
func (h *Handler) GetRoleByID(c *fiber.Ctx) error {
	//string param id
	paramID := c.Params("id")

	//convert to int
	id, err := strconv.Atoi(paramID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "")
	}

	//fetch role with id
	role, err := models.Roles(models.RoleWhere.RoleID.EQ(int64(id))).One(c.Context(), h.DB)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.SendError(c, fiber.StatusNotFound, "role does not exist")
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "")
	}

	return utils.SendResponse(c, fiber.StatusOK, role)
}
