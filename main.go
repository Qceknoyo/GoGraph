package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type PageData struct {
	Filename string
	Result   string
}

var spaTemplate = template.Must(template.ParseFiles("templates/spa.html"))

var db *sql.DB

func spa(w http.ResponseWriter, r *http.Request) {
	spaTemplate.Execute(w, PageData{})
}

// скип ngroka (если есть желание через него хостить)
func skipNgrokWarning(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ngrok-skip-browser-warning", "true")
		next.ServeHTTP(w, r)
	})
}

func main() {

	fs := http.FileServer(http.Dir("./ui/static/"))
	http.Handle("/ui/static/", http.StripPrefix("/ui/static/", fs))

	if err := godotenv.Load(); err != nil {
		log.Println(".env не найден")
	}

	connStr := getConnectionString()

	var err error
	fmt.Println("Подключаемся к БД...")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("БД подключена")
	defer db.Close()

	go cleanupExpiredSessions(db)

	http.HandleFunc("/", withSession(db, spa))
	http.HandleFunc("/upload", withSession(db, upload))
	http.HandleFunc("/history", withSession(db, historyHandler))
	http.HandleFunc("/clearHistory", withSession(db, clearHistoryHandler))
	http.HandleFunc("/export", withSession(db, exportHandler))
	http.HandleFunc("/me", withSession(db, meHandler))
	http.HandleFunc("/sync/generate", withSession(db, syncGenerateHandler))
	http.HandleFunc("/sync", syncActivateHandler)
	http.HandleFunc("/connectDB", withSession(db, connectDB))

	log.Fatal(http.ListenAndServe(":8181", skipNgrokWarning(http.DefaultServeMux)))
}
