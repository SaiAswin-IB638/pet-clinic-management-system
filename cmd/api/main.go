package main

import (
	"net/http"
	"os"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/initializers"
	"github.com/MSaiAswin/pet-clinic-management-system/cmd/logger"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/routes"

	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
)

func init() {

	l := logger.Get()
	l.Info().Msg("Initializing application...")

	err := initializers.ConnectDB()
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to the database")
	}
	l.Info().Msg("Connected to the database successfully")

	err = initializers.MigrateDB()
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to migrate the database")
	}
	l.Info().Msg("Database migration completed successfully")

}

// @title Pet Clinic Management System API
// @version 1.0
// @description This is the API documentation for the Pet Clinic Management System.
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {

	l := logger.Get()
	router := routes.NewRouter()

	http.Handle("/", router)

	port := os.Getenv("PORT")
	l.Info().Str("port", port).Msg("Server is starting on port: " + port)
	l.Fatal().Err(http.ListenAndServe(":"+port, nil)).Msg("Server failed to start")

}
