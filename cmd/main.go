package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"music-session-app/internal/handlers"
	appmiddleware "music-session-app/internal/middleware"
	"music-session-app/internal/store"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("DB connect error:", err)
	}
	defer db.Close()

	// SCS session manager backed by Postgres
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	userStore := &store.UserStore{DB: db}

	authHandler := &handlers.AuthHandler{Users: userStore, Sessions: sessionManager}
	playerHandler := &handlers.PlayerHandler{Sessions: sessionManager}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// SCS middleware wraps every request to load/save session
	r.Use(sessionManager.LoadAndSave)

	// Static assets (audio tracks, etc.)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Public routes
	r.Get("/login", authHandler.ShowLogin)
	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.RequireAuth(sessionManager))
		r.Get("/player", playerHandler.ShowPlayer)
		r.Post("/player/save", playerHandler.SaveProgress)
		r.Post("/logout", authHandler.Logout)
	})

	port := os.Getenv("PORT")
	log.Printf("Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
