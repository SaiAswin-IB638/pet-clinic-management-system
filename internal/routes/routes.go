package routes

import (
	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/handlers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/middleware"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()

	handlerService := handlers.NewService()

	router.Use(middleware.RequestLogger)

	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/signup", handlerService.SignupHandler).Methods("POST")
	router.HandleFunc("/login", handlerService.LoginHandler).Methods("POST")

	protectedRouter := router.PathPrefix("/").Subrouter()
	protectedRouter.Use(middleware.ValidateJWT)

	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.ProtectAdminRoute)

	staffRouter := protectedRouter.PathPrefix("/staff").Subrouter()
	staffRouter.Use(middleware.ProtectStaffRoute)

	ownerRouter := protectedRouter.PathPrefix("/").Subrouter()
	ownerRouter.Use(middleware.ProtectOwnerRoute)

	ownerRouter.HandleFunc("/owners", handlerService.GetUserByIDHandler).Methods("GET")
	ownerRouter.HandleFunc("/owners", handlerService.UpdateUserHandler).Methods("PUT")
	ownerRouter.HandleFunc("/owners", handlerService.DeleteUserHandler).Methods("DELETE")

	staffRouter.HandleFunc("/pets", handlerService.GetAllPetsHandler).Methods("GET")
	staffRouter.HandleFunc("/pets/{id}/upload", handlerService.UploadPetDocumentHandler).Methods("POST")
	ownerRouter.HandleFunc("/pets", handlerService.GetPetsByOwnerHandler).Methods("GET")
	ownerRouter.HandleFunc("/pets", handlerService.CreatePetHandler).Methods("POST")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.GetPetByIDHandler).Methods("GET")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.UpdatePetHandler).Methods("PUT")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.DeletePetHandler).Methods("DELETE")
	ownerRouter.HandleFunc("/pets/{id}/documents", handlerService.GetPetDocumentsHandler).Methods("GET")
	ownerRouter.HandleFunc("/pets/{id}/documents/{docName}", handlerService.GetPetDocumentByNameHandler).Methods("GET")

	staffRouter.HandleFunc("/appointments/upcoming", handlerService.GetUpcomingAppointmentsHandler).Methods("GET")
	staffRouter.HandleFunc("/appointments/today", handlerService.GetTodayAppointmentsHandler).Methods("GET")
	ownerRouter.HandleFunc("/appointments", handlerService.GetUpcomingAppointmentsByOwnerHandler).Methods("GET")
	ownerRouter.HandleFunc("/appointments", handlerService.CreateAppointmentHandler).Methods("POST")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.GetAppointmentByIDHandler).Methods("GET")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.UpdateAppointmentHandler).Methods("PUT")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.DeleteAppointmentHandler).Methods("DELETE")

	return router
}
