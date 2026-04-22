package tests

import (
	"bytes"
	"encoding/json"

	//"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/its-ernest/rundown-workers/internal/store"
	"github.com/its-ernest/rundown-workers/pkg/engine"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	// 1. Setup a clean test database
	dbPath := "test_integration.db"
	defer os.Remove(dbPath)

	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)

	// 2. Setup a test Echo instance (no need to start a real server on a port)
	e := echo.New()

	// Re-register the handlers for the test
	e.POST("/enqueue", func(c *echo.Context) error {
		var req struct {
			Queue      string `json:"queue"`
			Tag        string `json:"tag,omitempty"`
			Payload    string `json:"payload"`
			Timeout    int    `json:"timeout"`
			MaxRetries int    `json:"max_retries"`
		}
		if err := c.Bind(&req); err != nil {
			return err
		}
		job, err := s.Enqueue(req.Queue, req.Tag, req.Payload, req.Timeout, req.MaxRetries)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, job)
	})

	// 3. Test Enqueue using httptest.NewRecorder
	t.Run("Enqueue", func(t *testing.T) {
		payload := map[string]interface{}{
			"queue":   "test_queue",
			"payload": "hello world",
		}
		jsonPayload, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewReader(jsonPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		// Serve the request through the Echo instance
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var job engine.Job
		err = json.Unmarshal(rec.Body.Bytes(), &job)
		require.NoError(t, err)
		assert.Equal(t, "test_queue", job.Queue)
		assert.Equal(t, "hello world", job.Payload)
		assert.Equal(t, engine.StatusPending, job.Status)
	})
}
