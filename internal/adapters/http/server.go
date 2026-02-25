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
	logger          *slog.Logger
}

func NewServer(
	port int,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	logger *slog.Logger,
	resourceHandler *handlers.ResourceHandler,
) *Server {
	s := &Server{
		resourceHandler: resourceHandler,
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
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errchek // it's ok
	})

	mux.HandleFunc("POST /resource", s.resourceHandler.Create)

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

		s.logger.Info("request comleted",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.Duration("duration", time.Since(start)),
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
		w.Header().Set("Access-Control-Alllow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE< OPTIONS")
		w.Header().Set("Access-Contro-Allow-Headers", "Content-Type, Authorization")

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

func (s *Server) Start() error {
	s.logger.Info("shutting down HTTP server")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
