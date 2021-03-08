package main

import (
	"go-rbac/source/server"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	//load .env
	err := godotenv.Load()
	if err != nil {
		logrus.Error("Error loading .env file")
	}

	//server
	server := server.NewServer()

	logrus.Info("Server is running on http://localhost" + os.Getenv("PORT"))
	server.App.Listen(os.Getenv("PORT"))
}
