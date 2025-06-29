package main

import (
	"fmt"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)


type HandlerFunc func(request string) string

// routeRequest is our primary router. It will now also decide which handler to call.
func routeRequest(request string) string {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return buildHTTPResponse("Bad Request", "text/plain", "400 Bad Request")
	}

	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")
	if len(parts) < 2 {
		return buildHTTPResponse("Bad Request", "text/plain", "400 Bad Request")
	}

	method := parts[0]
	path := parts[1]

	switch method {
	case "GET":
		// Simple GET routes
		if path == "/" {
			return buildHTTPResponse("Welcome to the homepage!", "text/plain", "200 OK")
		} else if path == "/about" {
			return buildHTTPResponse("This is the about page.", "text/plain", "200 OK")
		} else {
			// Try to serve static file
			content, contentType, err := serveStaticFile(path)
			if err != nil {
				return buildHTTPResponse("File Not Found", "text/plain", "404 Not Found")
			}
			// Use the dynamic content type from the file
			return buildHTTPResponse(string(content), contentType, "200 OK")
		}

	
	case "POST":
		if path == "/submit" {
			// The request body is separated from headers by a double newline "\r\n\r\n"
			requestParts := strings.SplitN(request, "\r\n\r\n", 2)
			body := ""
			if len(requestParts) > 1 {
				body = requestParts[1]
			}
			responseMessage := fmt.Sprintf("Received your POST request with body:\n%s", body)
			return buildHTTPResponse(responseMessage, "text/plain", "200 OK")
		}
		// Fall through to method not allowed if path is not /submit
		fallthrough

	default:
		return buildHTTPResponse("Method Not Allowed", "text/plain", "405 Method Not Allowed")
	}
}

// buildHTTPResponse has become more flexible, accepting a body, content type, and status code.
func buildHTTPResponse(body string, contentType string, status string) string {
	return "HTTP/1.1 " + status + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"\r\n" +
		body
}


// It also returns []byte to support binary files like images.
func serveStaticFile(path string) ([]byte, string, error) {
	// Security: Clean the path to prevent directory traversal attacks (e.g., /../../etc/passwd)
	cleanPath := filepath.Clean(strings.TrimPrefix(path, "/"))
	if strings.Contains(cleanPath, "..") {
		return nil, "", fmt.Errorf("invalid path")
	}

	filePath := "public/" + cleanPath

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	// Detect the MIME type based on the file extension
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		// If the type can't be determined, default to a generic binary stream
		contentType = "application/octet-stream"
	}

	return data, contentType, nil
}

// acceptConnections remains the same, it listens for and accepts new connections.
func acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err) // Use log instead of fmt
			continue
		}
		go handleConnection(conn)
	}
}


// It takes a handler function as an argument and returns a new handler function.
func loggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(request string) string {
		startTime := time.Now()

		// Get the request line for logging
		requestLine := strings.Split(request, "\r\n")[0]

		// Call the original handler (e.g., routeRequest)
		response := next(request)

		duration := time.Since(startTime)
		responseLine := strings.Split(response, "\r\n")[0]

		// Log the details
		log.Printf("Request: \"%s\" | Response: \"%s\" | Duration: %s", requestLine, responseLine, duration)

		return response
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 2048) // Increased buffer size slightly
	n, err := conn.Read(buffer)
	if err != nil {
		// Don't print an error if the client closes the connection
		if err.Error() != "EOF" {
			log.Println("Error reading from the connection: ", err)
		}
		return
	}
	request := string(buffer[:n])

	// The routeRequest function is now "wrapped" by our logging middleware.
	wrappedHandler := loggingMiddleware(routeRequest)
	response := wrappedHandler(request)

	conn.Write([]byte(response))
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	// Create the public directory if it doesn't exist
	_ = os.Mkdir("public", 0755)

	log.Println("Server is listening on port 8080")
	log.Println("Serving files from the 'public' directory.")

	acceptConnections(listener)
}