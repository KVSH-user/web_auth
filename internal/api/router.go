package api

import (
	"net/http"
	"web_auth/internal/modules/auth"
	"web_auth/internal/modules/messages"

	"github.com/go-chi/chi/v5"
)

func NewRouter(authService *auth.Auth, messageService *messages.MessageService) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", RegisterHandler(authService))
	r.Post("/login", LoginHandler(authService))

	r.Get("/users", ListUsersHandler(authService))
	r.Get("/users/{userID}", GetUserHandler(authService))
	r.Post("/users/{userID}/block", BlockUserHandler(authService))

	r.Get("/users/{userID}/messages", GetUserMessagesHandler(messageService))

	return r
}
