// main.go
// This is the entry point of our application. Its primary role is to
// create a new server, register routes and middleware, and start listening
// for connections.

package main

import (
	"log"
	"os"
)

func main() {
	// Ensure the public directory exists.
	if err := os.MkdirAll("public", 0755); err != nil {
		log.Fatalf("Failed to create public directory: %v", err)
	}

	// Create a new server instance.
	server := NewServer(":8080")

	// Register middleware. Logging will wrap all handlers.
	server.Use(loggingMiddleware)

	// Register application-specific routes.
	server.Handle("GET", "/", homeHandler)
	server.Handle("GET", "/about", aboutHandler)
	server.Handle("POST", "/submit", submitHandler)

	// The static file server is now configured as the fallback for any GET
	// request that doesn't match the routes above.
	server.SetNotFoundHandler(serveStaticFile)


	// Start the server.
	log.Printf("Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		// Log.Fatal will not be called on graceful shutdown.
		log.Printf("Server stopped: %v", err)
	}
}