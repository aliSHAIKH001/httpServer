// server.go
// This file defines the core Server and Router components. It's the engine
// that listens for connections, uses the router to find the correct handler,
// and manages the overall server lifecycle including graceful shutdown.

package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type HandlerFunc func(w ResponseWriter, r *Request)
type Middleware func(next HandlerFunc) HandlerFunc

// Router holds the mappings of routes to their handlers.
type Router struct {
	routes         map[string]map[string]HandlerFunc
	notFoundHandler HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
		notFoundHandler: func(w ResponseWriter, r *Request) {
			httpError(w, 404) // The default not found handler-version
		},
	}
}

// Registers the handlers 
func (rt *Router) Handle(method, path string, handler HandlerFunc) {
	if rt.routes[method] == nil {
		rt.routes[method] = make(map[string]HandlerFunc)
	}
	rt.routes[method][path] = handler
}

func (rt *Router) SetNotFoundHandler(handler HandlerFunc) {
	rt.notFoundHandler = handler
}

func (rt *Router) findHandler(method, path string) HandlerFunc {
	if methodHandlers, ok := rt.routes[method]; ok {
		if handler, ok := methodHandlers[path]; ok {
			return handler
		}
	}
	return rt.notFoundHandler
}

// Server is the core of our web server.
type Server struct {
	Addr       string
	router     *Router
	middleware []Middleware
	wg         sync.WaitGroup
}

func NewServer(addr string) *Server {
	return &Server{
		Addr:   addr,
		router: NewRouter(),
	}
}

// Called by our server from main file, this in turn calls the routers handle function above
func (s *Server) Handle(method, path string, handler HandlerFunc) {
	s.router.Handle(method, path, handler)
}

func (s *Server) Use(mw Middleware) {
	s.middleware = append(s.middleware, mw)
}

func (s *Server) SetNotFoundHandler(handler HandlerFunc) {
	s.router.SetNotFoundHandler(handler)
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	shutdownCtx, shutdownRelease := context.WithCancel(context.Background())
	go s.handleShutdownSignal(shutdownRelease)

	for {
		select {
		case <-shutdownCtx.Done():
			s.wg.Wait()
			return nil
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := listener.Accept()
			if err != nil {
				// If the deadline is hit, a timeout error occurs and we loop again to avoid being stuck.
				if os.IsTimeout(err) {
					continue
				}
				return err
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// Exists/runs in the background and shuts down the server after a shudown-signal like ctrl + C, etc.
func (s *Server) handleShutdownSignal(release func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutdown signal received, stopping new connections.")
	release()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	req, err := parseRequest(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		httpError(newResponse(conn), 400)
		return
	}

	handler := s.router.findHandler(req.Method, req.Path)

	// Wraps all the middlewares we have, like an onion layer around the main handler.
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}

	// newResponse function creates a Response struct
	resp := newResponse(conn)
	handler(resp, req)
}
