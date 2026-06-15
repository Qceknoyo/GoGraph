package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	sessionKey contextKey = "session_token"
	userKey    contextKey = "user_id"
)

func withSession(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session_id")

		if err != nil {

			userID := uuid.New().String()
			sessionToken := uuid.New().String()

			_, err = db.Exec(`
                INSERT INTO users (id)
                VALUES ($1)
            `, userID)

			if err != nil {
				http.Error(w, "Ошибка создания user", 500)
				return
			}

			_, err = db.Exec(`
                INSERT INTO user_sessions (token, user_id)
                VALUES ($1, $2)
            `, sessionToken, userID)

			if err != nil {
				http.Error(w, "Ошибка создания session", 500)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionToken,
				MaxAge:   int((30 * 24 * time.Hour).Seconds()),
				HttpOnly: true,
				Path:     "/",
			})

			ctx := context.WithValue(r.Context(), sessionKey, sessionToken)
			ctx = context.WithValue(ctx, userKey, userID)

			next(w, r.WithContext(ctx))
			return
		}

		var userID string

		err = db.QueryRow(`
            SELECT user_id
            FROM user_sessions
            WHERE token = $1
        `, cookie.Value).Scan(&userID)

		if err != nil {

			userID = uuid.New().String()
			sessionToken := uuid.New().String()

			_, err = db.Exec(`
                INSERT INTO users (id)
                VALUES ($1)
            `, userID)

			if err != nil {
				http.Error(w, "Ошибка создания user", 500)
				return
			}

			_, err = db.Exec(`
                INSERT INTO user_sessions (token, user_id)
                VALUES ($1, $2)
            `, sessionToken, userID)

			if err != nil {
				http.Error(w, "Ошибка создания session", 500)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionToken,
				MaxAge:   int((30 * 24 * time.Hour).Seconds()),
				HttpOnly: true,
				Path:     "/",
			})

			ctx := context.WithValue(r.Context(), sessionKey, sessionToken)
			ctx = context.WithValue(ctx, userKey, userID)

			next(w, r.WithContext(ctx))
			return
		}

		db.Exec(`UPDATE users SET last_seen = NOW() WHERE id = $1`, userID)

		ctx := context.WithValue(r.Context(), sessionKey, cookie.Value)
		ctx = context.WithValue(ctx, userKey, userID)

		next(w, r.WithContext(ctx))
	}
}

// фоновая чистка протухших сессий
func cleanupExpiredSessions(db *sql.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		result, err := db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW()`)
		if err != nil {
			log.Printf("Ошибка очистки сессий: %v", err)
			continue
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Printf("Удалено %d устаревших сессий", rows)
		}
	}
}

// генерация ссылки
func syncGenerateHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userKey).(string)
	token := uuid.New().String()

	_, err := db.Exec(`
        INSERT INTO sync_tokens (token, user_id)
        VALUES ($1, $2)
    `, token, userID)

	if err != nil {
		http.Error(w, "Ошибка генерации токена", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// активация ссылки
func syncActivateHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var userID string
	err := db.QueryRow(`
        SELECT user_id FROM sync_tokens
        WHERE token = $1 AND expires_at > NOW()
    `, token).Scan(&userID)

	if err != nil {
		http.Error(w, "Токен не найден или истёк", http.StatusForbidden)
		return
	}

	db.Exec(`DELETE FROM sync_tokens WHERE token = $1`, token)

	sessionToken := uuid.New().String()
	db.Exec(`
        INSERT INTO user_sessions (token, user_id)
        VALUES ($1, $2)
    `, sessionToken, userID)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionToken,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
