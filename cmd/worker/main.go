package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/its-ernest/rundown-workers/internal/store"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// EnqueueRequest defines the parameters for adding a new job to a queue.
type EnqueueRequest struct {
	Queue      string `json:"queue"`
	Payload    string `json:"payload"`
	Timeout    int    `json:"timeout"`
	MaxRetries int    `json:"max_retries"`
}

// PollRequest defines the queue to poll for new jobs.
type PollRequest struct {
	Queue string `json:"queue"`
}

// CompleteRequest identifies a job to be marked as finished.
type CompleteRequest struct {
	ID string `json:"id"`
}

// FailRequest identifies a job to be marked as failed.
type FailRequest struct {
	ID string `json:"id"`
}

func main() {
	port := flag.Int("port", 8181, "Port to run the engine on")
	flag.Parse()

	e := echo.New()

	// Use ONLY existing middlewares in v5
	e.Use(middleware.Recover())

	// Initialize store
	s, err := store.NewSQLiteStore("rundown_v2.db")
	if err != nil {
		panic(fmt.Sprintf("Error initializing store: %v", err))
	}

	// Endpoints
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	// Live Documentation rendering with gomarkdown
	e.GET("/docs", func(c *echo.Context) error {
		fmt.Println("[DEBUG] /docs endpoint was hit!")
		content, err := os.ReadFile("DOCUMENTATION.md")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Documentation not found")
		}

		// Convert markdown to HTML
		html := markdown.ToHTML(content, nil, nil)

		// Wrap in a HTML template
		styledHTML := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Rundown-Workers Docs</title>
				<style>
					body { font-family: -apple-system, system-ui, "Segoe UI", Helvetica, Arial, sans-serif; line-height: 1.6; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #333; }
					pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
					code { font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace; font-size: 0.9em; }
					h1, h2, h3 { color: #1a1a1a; margin-top: 2em; }
					a { color: #0366d6; text-decoration: none; }
					a:hover { text-decoration: underline; }
				</style>
			</head>
			<body>%s</body>
			</html>`, html)

		return c.HTML(http.StatusOK, styledHTML)
	})

	// 1. Enqueue a job
	e.POST("/enqueue", func(c *echo.Context) error {
		var req EnqueueRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		job, err := s.Enqueue(req.Queue, req.Payload, req.Timeout, req.MaxRetries)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, job)
	})

	// 2. Poll for a job
	e.POST("/poll", func(c *echo.Context) error {
		var req PollRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		job, err := s.Poll(req.Queue)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if job == nil {
			return c.NoContent(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, job)
	})

	// 3. Mark job as complete
	e.POST("/complete", func(c *echo.Context) error {
		var req CompleteRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		err := s.Complete(req.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})

	// 4. Mark job as failed
	e.POST("/fail", func(c *echo.Context) error {
		var req FailRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		err := s.Fail(req.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.NoContent(http.StatusOK)
	})

	// Start Staleness Checker (Background recovery)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			affected, err := s.CleanupStale()
			if err != nil {
				fmt.Printf("[!] Cleanup error: %v\n", err)
			} else if affected > 0 {
				fmt.Printf("[*] Recovered %d stale jobs\n", affected)
			}
		}
	}()

	fmt.Printf("Rundown-Workers Engine v0.1.0 starting on :%d\n", *port)
	if err := e.Start(fmt.Sprintf(":%d", *port)); err != nil && err != http.ErrServerClosed {
		fmt.Printf("[!] Engine crashed: %v\n", err)
	}
}
