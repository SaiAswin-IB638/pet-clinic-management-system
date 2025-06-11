package routes

import (
	"net/http"

	_ "github.com/MSaiAswin/pet-clinic-management-system/cmd/api/docs"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/handlers"
	"github.com/MSaiAswin/pet-clinic-management-system/internal/middleware"
	"github.com/gorilla/mux"
	corsHandler "github.com/gorilla/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()

	handlerService := handlers.NewService()
	router.Use(middleware.RequestLogger)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			next.ServeHTTP(w, r)
		})
	})
	cors := corsHandler.CORS(
		corsHandler.AllowedOrigins([]string{"*"}),
		corsHandler.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		corsHandler.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		corsHandler.AllowCredentials(),
		corsHandler.MaxAge(3600),
	)
	router.Use(cors)

	

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	router.HandleFunc("/signup", handlerService.SignupHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", handlerService.LoginHandler).Methods("POST", "OPTIONS")

	protectedRouter := router.PathPrefix("/").Subrouter()
	protectedRouter.Use(middleware.ValidateJWT)

	adminRouter := protectedRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.ProtectAdminRoute)

	staffRouter := protectedRouter.PathPrefix("/staff").Subrouter()
	staffRouter.Use(middleware.ProtectStaffRoute)

	ownerRouter := protectedRouter.PathPrefix("/").Subrouter()
	ownerRouter.Use(middleware.ProtectOwnerRoute)

	ownerRouter.HandleFunc("/owners", handlerService.GetUserByIDHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/owners", handlerService.UpdateUserHandler).Methods("PUT", "OPTIONS")
	ownerRouter.HandleFunc("/owners", handlerService.DeleteUserHandler).Methods("DELETE", "OPTIONS")

	staffRouter.HandleFunc("/pets", handlerService.GetAllPetsHandler).Methods("GET", "OPTIONS")
	staffRouter.HandleFunc("/pets/{id}/upload", handlerService.UploadPetDocumentHandler).Methods("POST", "OPTIONS")
	ownerRouter.HandleFunc("/pets", handlerService.GetPetsByOwnerHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/pets", handlerService.CreatePetHandler).Methods("POST", "OPTIONS")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.GetPetByIDHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.UpdatePetHandler).Methods("PUT", "OPTIONS")
	ownerRouter.HandleFunc("/pets/{id}", handlerService.DeletePetHandler).Methods("DELETE", "OPTIONS")
	ownerRouter.HandleFunc("/pets/{id}/documents", handlerService.GetPetDocumentsHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/pets/{id}/documents/{docName}", handlerService.GetPetDocumentByNameHandler).Methods("GET", "OPTIONS")

	staffRouter.HandleFunc("/appointments/upcoming", handlerService.GetUpcomingAppointmentsHandler).Methods("GET", "OPTIONS")
	staffRouter.HandleFunc("/appointments/today", handlerService.GetTodayAppointmentsHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/appointments", handlerService.GetUpcomingAppointmentsByOwnerHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/appointments", handlerService.CreateAppointmentHandler).Methods("POST", "OPTIONS")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.GetAppointmentByIDHandler).Methods("GET", "OPTIONS")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.UpdateAppointmentHandler).Methods("PUT", "OPTIONS")
	ownerRouter.HandleFunc("/appointments/{id}", handlerService.DeleteAppointmentHandler).Methods("DELETE", "OPTIONS")

	return router
}
