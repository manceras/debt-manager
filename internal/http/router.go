package http

import (
	"debt-manager/internal/http/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMux(s *handlers.Server) *chi.Mux {
	r := chi.NewRouter()

	// public
	r.Post("/auth/signup", s.SignUp)
	r.Post("/auth/login", s.Login)
	r.Post("/auth/refresh", s.Refresh)

	// private
	r.Route("/", func(private chi.Router){
		private.Use(s.Auth)
		private.Post("/lists", s.CreateList)
		private.Get("/lists", s.GetLists)
	})

	return r
}
