package routes

import (
	"go-rbac/source/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (r *Router) userRoutes(router fiber.Router) {
	router.Post("/password/change", middlewares.LoggedIn(), r.handler.ChangePassword)
	router.Get("/me", middlewares.LoggedIn(), r.handler.MyProfile)
	router.Get("/:id", r.handler.GetUserByID)
	router.Get("/", r.handler.GetUsers)
}
