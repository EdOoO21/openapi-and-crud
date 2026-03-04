package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := chi.NewRouter()

	// Middleware: логирование, request ID
	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)

	// Подключаем сгенерированный OpenAPI сервер
	// apiHandler := api.NewHandler() // здесь ты реализуешь интерфейс из gen.go
	// api.RegisterHandlers(router, apiHandler)

	log.Printf("Starting server on :%s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
