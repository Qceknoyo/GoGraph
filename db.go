package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// подключение через .env
func getConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
}

// подключение к произвольной бд (демонстрационная функция)
func connectDB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	host := r.FormValue("host")
	dbname := r.FormValue("dbname")
	port := r.FormValue("port")
	user := r.FormValue("user")
	password := r.FormValue("password")

	w.Header().Set("Content-Type", "application/json")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Ошибка драйвера"})
		return
	}
	defer testDB.Close()

	if err := testDB.Ping(); err != nil {
		fmt.Printf("Ошибка из базы: %v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("База недоступна: %v", err),
		})
		return
	}

	fmt.Println("Связь с базой установлена!")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Связь установлена!",
	})
}
