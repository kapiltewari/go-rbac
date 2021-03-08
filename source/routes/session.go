package routes

import (
	"go-rbac/source/middlewares"

	"github.com/gofiber/fiber/v2"
)

//session related routes
func (r *Router) sessionRoutes(router fiber.Router) {
	router.Post("/register", r.handler.RegisterUser)
	router.Post("/login", r.handler.LoginUser)
	router.Post("/activate", r.handler.ActivateAccount)
	router.Post("/new-code", r.handler.GenerateNewCode)
	router.Post("/reset", r.handler.ResetPassword)
	router.Post("/logout", middlewares.LoggedIn(), r.handler.LogoutUser)
	router.Post("/refresh-token", r.handler.RefreshTokens)
}
