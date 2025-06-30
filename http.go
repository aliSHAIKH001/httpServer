// http.go
// This file contains types and functions related to the HTTP protocol itself.
// It defines the structure of a Request and a ResponseWriter, and includes
// the logic for parsing an incoming raw request from a TCP connection.

package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Request represents a parsed HTTP request, this is passed to handlers as one of the arguments.
type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    string
	Conn    net.Conn
}

// ResponseWriter is an interface used by an HTTP handler to construct an HTTP response.
type ResponseWriter interface {
	SetHeader(key, value string)
	WriteHeader(statusCode int)
	Write(data []byte) (int, error)
	Status() int
}

type response struct {
	conn        net.Conn
	headers     map[string]string
	statusCode  int
	wroteHeader bool
}

func newResponse(conn net.Conn) *response {
	return &response{
		conn:    conn,
		headers: make(map[string]string),
		statusCode: 200,
	}
}

func (rw *response) SetHeader(key, value string) {
	rw.headers[key] = value
}

func (rw *response) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = statusCode
	statusText := StatusText(statusCode)

	// For status info
	fmt.Fprintf(rw.conn, "HTTP/1.1 %d %s\r\n", rw.statusCode, statusText)
	// Next in line are the headers
	for key, value := range rw.headers {
		fmt.Fprintf(rw.conn, "%s: %s\r\n", key, value)
	}
	// Now the end of headers
	fmt.Fprint(rw.conn, "\r\n")
	rw.wroteHeader = true
}

// Main function that writes to the client 
func (rw *response) Write(data []byte) (int, error) {
	if !rw.wroteHeader {
		if _, ok := rw.headers["Content-Length"]; !ok {
			rw.SetHeader("Content-Length", fmt.Sprintf("%d", len(data)))
		}
		rw.WriteHeader(rw.statusCode)
	}
	return rw.conn.Write(data)
}

func (rw *response) Status() int {
	return rw.statusCode
}

func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line")
	}

	req := &Request{
		Method: parts[0], Path: parts[1], Version: parts[2],
		Headers: make(map[string]string), Conn: conn,
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil { return nil, err }
		line = strings.TrimSpace(line)
		if line == "" { break }
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 { continue }
		req.Headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
	}

	if contentLengthStr, ok := req.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLengthStr)
		if err != nil { return nil, fmt.Errorf("invalid Content-Length: %v", err) }
		
		if length > 0 {
			body := make([]byte, length)
			_, err := reader.Read(body)
			if err != nil { return nil, err }
			req.Body = string(body)
		}
	}
	return req, nil
}

func StatusText(code int) string {
	switch code {
	case 200: return "OK"
	case 400: return "Bad Request"
	case 404: return "Not Found"
	case 405: return "Method Not Allowed"
	case 500: return "Internal Server Error"
	default: return ""
	}
}

func httpError(w ResponseWriter, code int) {
	w.SetHeader("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%d %s", code, StatusText(code))
}
