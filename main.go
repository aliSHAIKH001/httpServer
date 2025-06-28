package main

import (
	"fmt"
	"net"
	"os"
)


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

	// Writing a basic HTTP response back to the client
	responseBody := "Hello, World!"
	response := "HTTP/1.1 200 OK\r\n" + 
		"Content-Type: text/plain\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(responseBody)) +
		"\r\n" +
		responseBody

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