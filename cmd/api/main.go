package main

import (
	"net/http"

	"github.com/MSaiAswin/pet-clinic-management-system/cmd/initializers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/routes"
	"github.com/MSaiAswin/pet-clinic-management-system/cmd/logger"
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

func main() {

	l := logger.Get()
	router := routes.NewRouter()

	http.Handle("/", router)

	port := "8001"
	l.Info().Str("port", port).Msg("Server is starting on port 8001")
	l.Fatal().Err(http.ListenAndServe(":"+port, nil)).Msg("Server failed to start")

}