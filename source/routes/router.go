/**
Register application routes here.
*/

package routes

import (
	"go-rbac/source/handlers"

	"github.com/gofiber/fiber/v2"
)

//Router ...
type Router struct {
	handler *handlers.Handler
}

//NewRouter ...
func NewRouter(handler *handlers.Handler, app *fiber.App) *Router {
	router := &Router{
		handler: handler,
	}

	//api router
	apiV1Router := app.Group("/api/v1")
	router.apiRoutes(apiV1Router)

	return router
}

func (r *Router) apiRoutes(router fiber.Router) {
	//session routes
	session := router.Group("/session")
	r.sessionRoutes(session)

	//user routes
	users := router.Group("/users")
	r.userRoutes(users)

	//role routes
	roles := router.Group("/roles")
	r.roleRoutes(roles)

	//search routes
	// search := router.Group("/search")
	// r.searchRoutes(search)
}
