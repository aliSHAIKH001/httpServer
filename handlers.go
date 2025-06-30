// handlers.go
// This file contains all the application-specific logic. Each HandlerFunc
// represents an action for a specific route (e.g., show the homepage,
// process a form). The logging middleware is also defined here as it's
// part of the application's behavior.

package main

import (
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// --- Middleware ---

func loggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w ResponseWriter, r *Request) {
		startTime := time.Now()
		next(w, r)
		duration := time.Since(startTime)
		log.Printf(
			`Request: "%s %s" | Response: "%d %s" | Duration: %s`,
			r.Method, r.Path, w.Status(), StatusText(w.Status()), duration,
		)
	}
}

// --- Page Handlers ---

func homeHandler(w ResponseWriter, r *Request) {
	w.SetHeader("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Welcome to the homepage!"))
}

func aboutHandler(w ResponseWriter, r *Request) {
	w.SetHeader("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("This is the about page."))
}

func submitHandler(w ResponseWriter, r *Request) {
	responseMessage := fmt.Sprintf("Received your POST request with body:\n%s", r.Body)
	w.SetHeader("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(responseMessage))
}

// --- File & Error Handlers ---

func serveStaticFile(w ResponseWriter, r *Request) {
	// This handler is now used as a fallback. We only serve files for GET requests.
	if r.Method != "GET" {
		httpError(w, 405) // Method Not Allowed
		return
	}
	
	cleanPath := filepath.Clean(strings.TrimPrefix(r.Path, "/"))
	if strings.HasPrefix(cleanPath, "..") {
		httpError(w, 400) // Bad Request
		return
	}

	filePath := filepath.Join("public", cleanPath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		// If the file doesn't exist, this is a 404.
		httpError(w, 404)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.SetHeader("Content-Type", contentType)
	w.WriteHeader(200)
	w.Write(data)
}