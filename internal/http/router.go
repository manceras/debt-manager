package http

import (
	"debt-manager/internal/http/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewMux(s *handlers.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// public
	r.Post("/auth/signup", s.SignUp)
	r.Post("/auth/login", s.Login)
	r.Post("/auth/refresh", s.Refresh)

	// private
	r.Group(func(private chi.Router){
		private.Use(s.Auth)

		// Lists
		private.Post("/lists", s.CreateList)
		private.Get("/lists", s.GetLists)
		private.Get("/lists/{list_id}", s.GetListByID)
		private.Patch("/lists/{list_id}", s.UpdateList)
		private.Delete("/lists/{list_id}", s.DeleteList)

		// Invitations
		private.Post("/lists/{list_id}/invitations", s.CreateInvitation)
		private.Get("/lists/{list_id}/invitations", s.GetAllInvitationsForList)
		private.Get("/invitations/{hash}", s.GetInvitationByHash)
		private.Get("/invitations/{invitation_id}", s.GetInvitationByID)
		private.Delete(("/invitations/{invitation_id}"), s.RevokeInvitation)

		// Users
		private.Get("/lists/{list_id}/users", s.GetUsersFromList)
	})

	return r
}
