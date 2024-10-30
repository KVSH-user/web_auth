package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"web_auth/internal/modules/auth"
	"web_auth/internal/modules/messages"

	"github.com/go-chi/chi/v5"
)

func ListUsersHandler(authService *auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10 // Значение по умолчанию
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0 // Значение по умолчанию
		}

		users, err := authService.ListUsers(r.Context(), limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(users)
	}
}

func GetUserHandler(authService *auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

		user, err := authService.GetUser(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}

func BlockUserHandler(authService *auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

		err := authService.BlockUser(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"successfully user blocked by id:": userID,
		})
	}
}

func GetUserMessagesHandler(messageService *messages.MessageService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10 // Значение по умолчанию
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0 // Значение по умолчанию
		}

		messages, err := messageService.GetUserMessages(r.Context(), userID, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}
