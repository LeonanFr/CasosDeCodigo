package main

import (
	"casos-de-codigo-api/internal/auth"
	"casos-de-codigo-api/internal/db"
	"casos-de-codigo-api/internal/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if err := auth.InitJWT(secret); err != nil {
		log.Fatalf("Erro ao inicializar JWT: %v", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDBName := os.Getenv("MONGO_DB")
	if mongoDBName == "" {
		mongoDBName = "casos_de_codigo"
	}

	mongoManager, err := db.NewMongoManager(mongoURI, mongoDBName)
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer mongoManager.Close()

	sqliteFactory := db.NewSQLiteFactory()

	authHandler := handlers.NewAuthHandler(mongoManager)
	caseHandler := handlers.NewCaseHandler(mongoManager)
	gameHandler := handlers.NewGameHandler(mongoManager, sqliteFactory)

	router := mux.NewRouter()

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.Handle("/api/auth/profile", auth.Middleware(http.HandlerFunc(authHandler.Profile))).Methods("GET")

	router.Handle("/api/cases", auth.Middleware(http.HandlerFunc(caseHandler.GetAllCases))).Methods("GET")
	router.Handle("/api/cases/{id}", auth.Middleware(http.HandlerFunc(caseHandler.GetCase))).Methods("GET")
	router.Handle("/api/cases/initialize", auth.Middleware(http.HandlerFunc(caseHandler.InitializeCase))).Methods("POST")

	router.Handle("/api/game/execute", auth.Middleware(http.HandlerFunc(gameHandler.ExecuteCommand))).Methods("POST")
	router.Handle("/api/game/progress", auth.Middleware(http.HandlerFunc(gameHandler.GetProgress))).Methods("GET")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Guest-ID"},
		AllowCredentials: true,
		Debug:            false,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Servidor iniciado na porta %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler.Handler(router)))
}
