package rw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Client is the Rundown-Workers SDK client.
type Client struct {
	Host string
}

// NewClient creates a new SDK client.
func NewClient(host string) *Client {
	if host == "" {
		host = "http://localhost:8181"
	}
	return &Client{Host: host}
}

// Enqueue submits a new job to the engine.
func (c *Client) Enqueue(queue, payload string, timeout, maxRetries int, tag string) (*Job, error) {
	reqBody := map[string]interface{}{
		"queue":       queue,
		"tag":         tag,
		"payload":     payload,
		"timeout":     timeout,
		"max_retries": maxRetries,
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(c.Host+"/enqueue", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %d", resp.StatusCode)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Job represents a unit of work.
type Job struct {
	ID         string `json:"id"`
	Queue      string `json:"queue"`
	Tag        string `json:"tag,omitempty"`
	Payload    string `json:"payload"`
	Status     string `json:"status"`
	Retries    int    `json:"retries"`
	MaxRetries int    `json:"max_retries"`
	Timeout    int    `json:"timeout"`
}

// worker represents a registered queue and its handler.
type worker struct {
	Queue        string
	Handler      func(payload string) bool
	PollInterval time.Duration
}

var (
	registry []worker
	mu       sync.Mutex
)

// Queue registers a new worker function for a specific queue.
func Queue(name string, handler func(string) bool, interval time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, worker{
		Queue:        name,
		Handler:      handler,
		PollInterval: interval,
	})
}

// Run starts all registered workers in their own goroutines.
func Run(host string) {
	if host == "" {
		host = "http://localhost:8181"
	}
	client := NewClient(host)

	fmt.Printf("[*] Rundown-Workers starting with %d workers...\n", len(registry))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for _, w := range registry {
		wg.Add(1)
		go startPolling(ctx, &wg, client, w)
	}

	<-sigChan
	fmt.Println("\n[*] Shutting down workers...")
	cancel()
	wg.Wait()
	fmt.Println("[*] All workers stopped.")
}

func startPolling(ctx context.Context, wg *sync.WaitGroup, client *Client, w worker) {
	defer wg.Done()
	fmt.Printf("[*] Poller started for queue: %s\n", w.Queue)

	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pollAndExecute(client, w)
		}
	}
}

func pollAndExecute(client *Client, w worker) {
	// 1. Poll for job
	data, _ := json.Marshal(map[string]string{"queue": w.Queue})
	resp, err := http.Post(client.Host+"/poll", "application/json", bytes.NewReader(data))
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		resp.Body.Close()
		return
	}
	resp.Body.Close()

	fmt.Printf("[*] Job %s assigned. Executing...\n", job.ID)

	// 2. Execute with local timeout
	success := false
	done := make(chan bool, 1)

	// Set a default timeout if none provided (or use job.Timeout)
	timeout := time.Duration(job.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	go func() {
		success = w.Handler(job.Payload)
		done <- true
	}()

	select {
	case <-done:
		// Task finished
		if success {
			_ = reportStatus(client, "/complete", job.ID)
		} else {
			_ = reportStatus(client, "/fail", job.ID)
		}
	case <-time.After(timeout):
		// Task timed out locally
		fmt.Printf("[!] Job %s timed out after %v\n", job.ID, timeout)
		_ = reportStatus(client, "/fail", job.ID)
	}
}

func reportStatus(client *Client, path, id string) error {
	data, _ := json.Marshal(map[string]string{"id": id})
	resp, err := http.Post(client.Host+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
