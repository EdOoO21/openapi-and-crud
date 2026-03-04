package main

import (
	"log"
	"net/http"
	"os"

	"github.com/EdOoO21/openapi-and-crud/internal/api"
	"github.com/EdOoO21/openapi-and-crud/internal/db"
	handlers "github.com/EdOoO21/openapi-and-crud/internal/handlers"
	"github.com/EdOoO21/openapi-and-crud/internal/middleware"
	"github.com/EdOoO21/openapi-and-crud/internal/repository"
	"github.com/EdOoO21/openapi-and-crud/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/marketplace?sslmode=disable"
	}

	sqlxDB, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer sqlxDB.Close()

	userRepo := repository.NewUserRepository(sqlxDB)
	prodRepo := repository.NewProductRepository(sqlxDB)
	orderRepo := repository.NewOrderRepository(sqlxDB)

	authSvc := service.NewAuthService(userRepo)
	prodSvc := service.NewProductService(prodRepo)
	orderSvc := service.NewOrderService(orderRepo, prodRepo, userRepo, sqlxDB)

	authHandler := handlers.NewAuthHandler(authSvc)
	prodHandler := handlers.NewProductHandler(prodSvc)
	orderHandler := handlers.NewOrderHandler(orderSvc)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging) // JSON logging
	r.Use(middleware.Auth)    // parses token if present and sets ctx user info

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/refresh", authHandler.Refresh)

	// mount OpenAPI-generated product routes (uses prodHandler)
	r.Mount("/", api.Handler(prodHandler))

	// Orders endpoints (not in generated OpenAPI file in your repo — added here)
	r.Route("/orders", func(r chi.Router) {
		r.Post("/", orderHandler.CreateOrder)            // POST /orders
		r.Put("/{id}", orderHandler.UpdateOrder)         // PUT /orders/{id}
		r.Post("/{id}/cancel", orderHandler.CancelOrder) // POST /orders/{id}/cancel
		r.Get("/{id}", orderHandler.GetOrder)            // GET /orders/{id}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
