package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func routeRequest(request string) string {
	lines := strings.Split(request, "\r\n")
	if len(lines) == 0 {
		return "HTTP/1.1 400 Bad Request\r\n\r\nBad Request"
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return "HTTP/1.1 400 Bad Request\r\n\r\nBad Request"
	}

	method := parts[0]
	path := parts[1]

	if method != "GET" {
		return "HTTP/1.1 405 Method Not Allowed\r\n\r\nMethod Not Allowed"
	}

	// Simple routes
	if path == "/" {
		return buildHTTPResponse("Welcome to the homepage!")
	} else if path == "/about" {
		return buildHTTPResponse("This is the about page.")
	} else {
		// Try to serve static file
		content, err := serveStaticFile(path)
		if err != nil {
			return "HTTP/1.1 404 Not Found\r\n\r\nFile Not Found"
		}
		return buildHTTPResponse(content)
	}
}

func buildHTTPResponse(body string) string {
	return "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"\r\n" +
		body
}

func serveStaticFile(path string) (string, error) {
	cleanPath := strings.TrimPrefix(path, "/")
	filePath := "public/" + cleanPath
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}


func acceptConnections(listener net.Listener) {
	for true {
		// Wait for a client to connect
		conn, err := listener.Accept()
		if  err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()

	// Redaing raw byte-data from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from the connection: ", err)
		return 
	}
	request := string(buffer[:n])

	fmt.Println("========Request Received========")
	fmt.Println(request)

	response := routeRequest(request)

	conn.Write([]byte(response))
}

func main() {
	// Start listening for incoming connections on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server: ",  err)
		os.Exit(1)
	}
	fmt.Println("Server is listening on port 8080")

	// Accept incoming connections continously
	acceptConnections(listener)
}