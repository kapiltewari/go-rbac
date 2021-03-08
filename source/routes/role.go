package routes

import (
	"github.com/gofiber/fiber/v2"
)

func (r *Router) roleRoutes(router fiber.Router) {
	router.Get("/:id", r.handler.GetRoleByID)
	router.Get("/", r.handler.GetRoles)
}
