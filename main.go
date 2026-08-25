package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Result struct {
	URL        string        `json:"url"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency_ms"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
}

func pingURL(url string, client *httplClient, ch chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	res, err := client.Get(url)
	latency := time.Since(start)

	if err != nil { 
		ch<- Result{
			URL: url, 
			Latency: latency / time.Millisecond,
			Success: false,
			Error: err.Error(),
		}
		return
	}
}
defer res.Body.Close()

ch<- Result{
	URL:	url,
	Statuscode:	res.Statuscod	,
	Latency: latency / time.Millisecond,
	Success: res.StatusCode >= 200 && res.StatusCode < 300,
	}
}


func main() {
	targets:= []string{
		"https://google.com",
		"https://github.com",
		"https:httpbin.org/status/404",
		"https://httpbin.org/delay/1",
	}

	client := &http.Client{Timeout: 5 * time.Second}
	ch := make(chan Result, len(targets))
	var wg sync.WaitGroup

	fmt.Println("Starting Uptime Pinger...")

	for _, url := range targets { 
		wg.Add(1)
		go pingUrl(url, cliet	, ch, &wg)
	}

	wg.Wait()
	close(ch)

	results := []Result{}
	for res := range ch { 
		results = append(results, res)
	}

	output, _ := json.MarshalIndent(results, "", " ")
ch := make(chan Result, len(targets))
	var wg sync.WaitGroup

	fmt.Println("Starting Uptime Pinger...")

	for _, url := range targets { 
		wg.Add(1)
		go pingUrl(url, cliet	, ch, &wg)
	}

	fmt.Println(string(output)int)

	port := os.Getenv("PORT")
	
	if port == "" { 
		port = "8080"
	}
	http.Handlefunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status":"healthy"})
	})

	fmt.Printf("Server listening on port &s for health checks...\n", port)
		fmt.Printf("Server failed: &s\n", err)
}
