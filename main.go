package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	//	"net/textproto"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tsperry/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	queries        database.Queries
}

func (config *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (config *apiConfig) requestCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf(`	
	<html>
		  <body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		  </body>
		</html>
		`, config.fileServerHits.Load())
	w.Write([]byte(hits))
}

func (config *apiConfig) resetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	config.fileServerHits.Store(0)
	hits := fmt.Sprintf("Hits: %v", config.fileServerHits.Load())

	w.Write([]byte(hits))
}

func replaceBadWords(s string) string {

	words := strings.Split(s, " ")
	for i, word := range words {
		if strings.ToLower(word) == "kerfuffle" {
			words[i] = "****"
		}
		if strings.ToLower(word) == "sharbert" {
			words[i] = "****"
		}
		if strings.ToLower(word) == "fornax" {
			words[i] = "****"
		}

	}
	return strings.Join(words, " ")

}

func (config *apiConfig) chirpValidator(w http.ResponseWriter, r *http.Request) {

	type chirpBody struct {
		Body string `json:"body"`
	}

	type errorBody struct {
		Error string `json:"error"`
	}

	type validBody struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpBody{}
	err := decoder.Decode(&chirp)

	if err != nil {
		log.Printf("Error parsing body: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(chirp.Body) > 140 {
		w.Header().Set("Content-Type", "text/josn; charset=utf-8")
		w.WriteHeader(400)
		respBody := errorBody{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("error marshalling json: %s", err)
		}
		w.Write(dat)
	} else {
		w.Header().Set("Content-Type", "text/josn; charset=utf-8")
		w.WriteHeader(200)

		type cleanChirp struct {
			Cleaned_body string `json:"cleaned_body"`
		}

		cleaned_chirp := cleanChirp{}

		cleaned_chirp.Cleaned_body = replaceBadWords(chirp.Body)

		jsonData, err := json.Marshal(cleaned_chirp)

		if err != nil {
			log.Printf("error marshalling json: %s", err)
		}

		w.Write(jsonData)

	}
}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Printf("error with db open: %s", err)
	}

	dbQueries := database.New(db)

	fmt.Println("server running... at localhost:8080")
	mux := http.NewServeMux()
	var server http.Server

	server.Handler = mux
	server.Addr = ":8080"

	var config apiConfig

	config.queries = *dbQueries

	mux.Handle("/app/", (&config).middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", readinessHandler)

	mux.HandleFunc("GET /admin/metrics", (&config).requestCounter)
	mux.HandleFunc("POST /admin/reset", (&config).resetCounter)
	mux.HandleFunc("POST /api/validate_chirp", (&config).chirpValidator)

	server.ListenAndServe()

}
