package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mradu/task-manager/internal/api"
	"github.com/mradu/task-manager/internal/config"
	"github.com/mradu/task-manager/internal/database"
)

func main() {
	// Initialize centralized configuration
	config.Load()

	// Connect to Database using config URL
	if err := database.Connect(config.AppConfig.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize Router
	r := chi.NewRouter()

	// Start SSE Hub
	go api.AppHub.Run()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // Can also be tied to config.AppConfig.FrontendURL in a real prod env
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, 
	}))

	// Routes
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", api.RegisterHandler)
		r.Post("/auth/login", api.LoginHandler)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(api.AuthMiddleware)
			
			r.Get("/users", api.GetUsersHandler)
			r.Get("/events", api.ServeSSE)

			r.Route("/tasks", func(r chi.Router) {
				r.Post("/", api.CreateTaskHandler)
				r.Get("/", api.GetTasksHandler)
				r.Get("/{id}", api.GetTaskByIDHandler)
				r.Patch("/{id}", api.UpdateTaskHandler)
				r.Delete("/{id}", api.DeleteTaskHandler)
				r.Post("/{id}/upload", api.UploadAttachmentHandler)
			})
		})
	})

	// Serve static uploads
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	r.Get("/uploads/*", func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		fs.ServeHTTP(w, r)
	})

	// Start Server
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", config.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
