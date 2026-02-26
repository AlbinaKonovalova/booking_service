package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AlbinaKonovalova/booking_service/internal/adapters/http/handlers"
)

// Server представляет HTTP сервер
type Server struct {
	server          *http.Server
	resourceHandler *handlers.ResourceHandler
	bookingHandler  *handlers.BookingHandler
	logger          *slog.Logger
}

// NewServer создает новый HTTP сервер
func NewServer(
	port int,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	logger *slog.Logger,
	resourceHandler *handlers.ResourceHandler,
	bookingHandler *handlers.BookingHandler,
) *Server {
	s := &Server{
		resourceHandler: resourceHandler,
		bookingHandler:  bookingHandler,
		logger:          logger,
	}

	mux := s.setupRoutes()

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.withMiddleware(mux),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return s
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck // it's ok
	})

	// Resource endpoints
	mux.HandleFunc("POST /resource", s.resourceHandler.Create)
	mux.HandleFunc("DELETE /resource/{id}", s.resourceHandler.Delete)
	mux.HandleFunc("GET /resource/{id}/bookings", s.bookingHandler.ListByResource)

	// Booking endpoints
	mux.HandleFunc("POST /booking", s.bookingHandler.Create)
	mux.HandleFunc("POST /booking/{id}/confirm", s.bookingHandler.Confirm)
	mux.HandleFunc("POST /booking/{id}/cancel", s.bookingHandler.Cancel)

	return mux
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return s.loggingMiddleware(s.recoveryMiddleware(s.corsMiddleware(next)))
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		s.logger.Info("request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.String("duration", time.Since(start).String()),
		)
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("panic recovered", slog.Any("error", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", slog.String("addr", s.server.Addr))
	return s.server.ListenAndServe()
}

// Shutdown gracefully останавливает HTTP сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
