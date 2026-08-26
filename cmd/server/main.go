package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Result struct {
	ID	   int	         `json:"id, omitempty"`
	URL        string        `json:"url"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency_ms"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	CreatedAt	 time.Time		  `json:"created_at,omitempty"`
}

var db *sql.DB

func initDB() { 
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/pinger?sslmode=disable"
	}

	var err error


	for i := 0; i < 5; i++ { 
		db, err = sql.Open("postgres", dbURL)
		if err == nil && db.Ping() == nil { 
			fmt.Println("Successfully connected to Database!")
			break
		}
		fmt.Println("Waiting for Database to be ready...")
		time.Sleep(2 * time.Second)
	}

	query := `
		CREATE TABLE IF NOT EXISTS ping_results ( 
			id SERIAL PRIMARY KEY,
			url TEXT NOT NULL,
			status_code INT,
			latency_ms BIGINT,
			success BOOLEAN,
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(query)
	if err != nil { 
		log.Fatalf("Failed to create table: %v", err)
	}
}

func pingURL(url string, client *http.Client, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	res, err := client.Get(url)
	latency := time.Since(start)

	if err != nil {
		ch <- Result{
			URL:     url,
			Latency: latency / time.Millisecond,
			Success: false,
			Error:   err.Error(),
		}
		return
	}
	defer res.Body.Close()

	ch <- Result{
		URL:        url,
		StatusCode: res.StatusCode,
		Latency:    latency / time.Millisecond,
		Success:    res.StatusCode >= 200 && res.StatusCode < 300,
	}
}

func saveResult(r Result) { 
	query := `INSERT INTO ping_results (url, status_code, latency_ms, success, error) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(query, r.URL, r.StatusCode, int64(r.Latency), r.Success, r.Error)
	if err != nil { 
		fmt.Printf("Error saving result to DB: %v\n", err)
	}
}

func main() {
	initDB()

	targets := []string{
		"https://google.com",
		"https://github.com",
		"https://httpbin.org/status/404",
		"https://httpbin.org/delay/1",
	}

	client := &http.Client{Timeout: 5 * time.Second}
	ch := make(chan Result, len(targets))
	var wg sync.WaitGroup

	for _, url := range targets { 
		wg.Add(1)
		go pingURL(url, client, ch, &wg)
	}

	wg.Wait()
	close(ch)

	for res := range ch { 
			saveResult(res)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) { 
		w.Header().Set("Content-Type", "application/json")
		rows, err := db.Query("SELECT id, url, status_code, latency_ms, success, COALESCE(error, ''), created_at FROM ping_results ORDER BY id DESC LIMIT 20")
		if err != nil { 
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []Result
		for rows.Next() {
			var r Result
			var latency int64
			if err := rows.Scan(&r.ID, &r.URL, &r.StatusCode, &r.Latency, &r.Success, &r.Error, &r.CreatedAt); err != nil {
				continue
			}
			r.Latency = time.Duration(latency)
			results = append(results, r)
		}
		json.NewEncoder(w).Encode(results)
	})

	fmt.Printf("Server listening on port %s... \n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}

