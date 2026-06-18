package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/its-ernest/rundown-workers/internal/store"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// EnqueueRequest modified: Payload now accepts raw structured JSON maps natively.
type EnqueueRequest struct {
	Queue      string      `json:"queue"`
	Tag        string      `json:"tag,omitempty"`
	Payload    interface{} `json:"payload"` // Changed from string to interface{} to accept json structures
	Timeout    int         `json:"timeout"`
	MaxRetries int         `json:"max_retries"`
}

// PollRequest modified: Added a limit ceiling parameter to fetch items in batches.
type PollRequest struct {
	Queue string `json:"queue"`
	Limit int    `json:"limit,omitempty"` // Allows picking multiple records simultaneously
}

type CompleteRequest struct {
	ID string `json:"id"`
}

type FailRequest struct {
	ID string `json:"id"`
}

func main() {
	defaultPort := os.Getenv("RUNDOWN_PORT")
	if defaultPort == "" {
		defaultPort = os.Getenv("PORT")
	}
	if defaultPort == "" {
		defaultPort = "8181"
	}

	defaultHost := os.Getenv("RUNDOWN_HOST")
	if defaultHost == "" {
		defaultHost = "0.0.0.0"
	}

	var portStr string
	flag.StringVar(&portStr, "port", defaultPort, "Port to run the engine on")
	host := flag.String("host", defaultHost, "Host to bind the engine on")
	flag.Parse()

	e := echo.New()

	e.Use(middleware.Recover())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogMethod:  true,
		LogURI:     true,
		LogLatency: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			statusColor := "\033[32m" // Green
			if v.Status >= 400 {
				statusColor = "\033[33m"
			}
			if v.Status >= 500 {
				statusColor = "\033[31m"
			}
			reset := "\033[0m"

			log.Printf("%s %s %s| %s%d%s | %10v | %s",
				"\033[34m[API]\033[0m", v.Method, v.URI,
				statusColor, v.Status, reset, v.Latency, c.Request().RemoteAddr)
			return nil
		},
	}))

	s, err := store.NewSQLiteStore("rundown_v2.db")
	if err != nil {
		panic(fmt.Sprintf("Error initializing store: %v", err))
	}

	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	e.GET("/docs", func(c *echo.Context) error {
		fmt.Println("[DEBUG] /docs endpoint was hit!")
		content, err := os.ReadFile("DOCUMENTATION.md")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "Documentation not found")
		}

		html := markdown.ToHTML(content, nil, nil)

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

	// 1. Enqueue an active job with direct formatting blocks
	e.POST("/enqueue", func(c *echo.Context) error {
		var req EnqueueRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		// Convert the incoming structured payload mapping directly into database text
		marshaledPayload, err := json.Marshal(req.Payload)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON structure parsed inside payload field")
		}

		job, err := s.Enqueue(req.Queue, req.Tag, string(marshaledPayload), req.Timeout, req.MaxRetries)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, job)
	})

	// 2. Poll for multiple jobs simultaneously inside a single slice array
	e.POST("/poll", func(c *echo.Context) error {
		var req PollRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		// Enforce a default processing threshold boundary if none is provided
		if req.Limit <= 0 {
			req.Limit = 1
		}

		var activeJobs []interface{}

		// Loop up to the targeted batch size parameters
		for i := 0; i < req.Limit; i++ {
			job, err := s.Poll(req.Queue)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			if job == nil {
				break // Queue drained cleanly
			}
			activeJobs = append(activeJobs, job)
		}

		if len(activeJobs) == 0 {
			return c.NoContent(http.StatusNoContent)
		}

		// Returns a plain JSON array list containing all unique job entities mapped with their IDs
		return c.JSON(http.StatusOK, activeJobs)
	})

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

	// Get job status by ID for direct tracking and reference
	e.GET("/status/:id", func(c *echo.Context) error {
		id := c.Param("id")
		job, err := s.GetJob(id) // Ensure your store has a GetJob method
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Job not found"})
		}
		return c.JSON(http.StatusOK, job)
	})

	/// Get job details by tag for easier tracking and reference
	e.GET("/details/:tag", func(c *echo.Context) error {
		tag := c.Param("tag")
		job, err := s.GetJobByTag(tag)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Job not found for this reference"})
		}
		return c.JSON(http.StatusOK, job)
	})

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

	fmt.Printf("Rundown-Workers Engine v0.2.0 starting on %s:%s\n", *host, portStr)
	if err := e.Start(fmt.Sprintf("%s:%s", *host, portStr)); err != nil && err != http.ErrServerClosed {
		fmt.Printf("[!] Engine crashed: %v\n", err)
	}
}
