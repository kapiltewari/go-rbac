package routes

import (
	"go-rbac/source/middlewares"

	"github.com/gofiber/fiber/v2"
)

//session related routes
func (r *Router) sessionRoutes(router fiber.Router) {
	router.Post("/register", r.handler.RegisterUser)
	router.Post("/account/activate", r.handler.AccountActivation)
	router.Post("/account/reverify", r.handler.AccountReverification)
	router.Post("/login", r.handler.LoginUser)
	router.Post("/password/forgot", r.handler.ForgotPassword)
	router.Post("/password/reset/:id/:token", r.handler.ResetPassword)
	router.Post("/logout", middlewares.LoggedIn(), r.handler.LogoutUser)
	router.Post("/tokens/refresh", r.handler.RefreshTokens)
}
