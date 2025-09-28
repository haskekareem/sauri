package sauri

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// defaultRouter built-in routes for the package
func (s *Sauri) defaultRouter() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	if s.DebugMode {
		mux.Use(middleware.Logger)
	}

	mux.Use(middleware.Recoverer)
	mux.Use(s.SessionLoad) // load and save session data
	mux.Use(s.NoSurf)

	return mux
}
