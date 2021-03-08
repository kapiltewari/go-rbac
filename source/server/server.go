package server

import (
	"go-rbac/db/connections"
	"go-rbac/source/handlers"
	"go-rbac/source/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

//Server ...
type Server struct {
	App *fiber.App
}

//NewServer ...
func NewServer() *Server {
	//postgresql
	db := connections.Postgres()

	//redis
	redis := connections.Redis()

	//handler has access to the db, redis
	//? you can also create a services package for your business logic and pass db, redis etc. to that package
	//? service := services.NewService(db, redis)
	//? and then pass service to handler
	//? handler := handlers.NewHandler(service)
	handler := handlers.NewHandler(db, redis)

	//new fiber app
	app := fiber.New()

	//logger
	app.Use(logger.New())

	//compress
	app.Use(compress.New())

	//cors
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000", //e.g. react app
		AllowCredentials: true,
	}))

	//files
	app.Static("/assets", "./assets")

	//passed app to NewRouter() because we can then create route groups
	//passed handler so router can have access to handler methods
	routes.NewRouter(handler, app)

	server := &Server{
		App: app,
	}

	return server
}
