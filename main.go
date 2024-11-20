package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Film struct {
	ID        int     `json:"id"`
	Nev       string  `json:"nev"`
	Tipus     string  `json:"tipus"`
	Ertekeles float64 `json:"ertekeles"`
}

var db *sql.DB

func initDB() {
	var err error

	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/Films")

	if err != nil {
		log.Fatal("Hiba az adatbázishoz való csatlakozáskor: ", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Nem sikerült elérni az adatbázist: ", err)
	}

	log.Println("Sikeresen kapcsolódott az adatbázishoz!")
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getFilms(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query("SELECT * FROM Filmek")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer rows.Close()

	var films []Film

	for rows.Next() {
		var film Film

		if err := rows.Scan(&film.ID, &film.Nev, &film.Tipus, &film.Ertekeles); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		films = append(films, film)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(films)
}

func addFilm(w http.ResponseWriter, r *http.Request) {
	var film Film

	if err := json.NewDecoder(r.Body).Decode(&film); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err := db.Exec("INSERT INTO Filmek (nev,tipus,ertekeles) VALUES (?, ?, ?)", film.Nev, film.Tipus, film.Ertekeles)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(film)

}

func deleteFilm(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Query().Get("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM Filmek WHERE id = ?", id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func getFilm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var film Film

	err = db.QueryRow("SELECT * FROM Filmek WHERE id = ?", id).Scan(&film.ID, &film.Nev, &film.Tipus, &film.Ertekeles)

	if err == sql.ErrNoRows {
		http.Error(w, "Film was not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(film)
}

func main() {
	initDB()

	defer db.Close()

	http.Handle("/films", enableCors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getFilms(w, r)
		case "POST":
			addFilm(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/film", enableCors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getFilm(w, r)
		case "DELETE":
			deleteFilm(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	log.Println("A szerver fut a 8080-as porton")
	http.ListenAndServe(":8080", nil)
}
