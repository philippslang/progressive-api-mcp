package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

type newBook struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type errorBody struct {
	Message string `json:"message"`
}

type bookstoreStore struct {
	mu         sync.RWMutex
	books      map[int]Book
	nextBookID int
}

func newBookstoreStore() *bookstoreStore {
	return &bookstoreStore{
		books:      make(map[int]Book),
		nextBookID: 1,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Message: msg})
}

func (s *bookstoreStore) listBooks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	books := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		books = append(books, b)
	}
	writeJSON(w, http.StatusOK, books)
}

func (s *bookstoreStore) createBook(w http.ResponseWriter, r *http.Request) {
	var nb newBook
	if err := json.NewDecoder(r.Body).Decode(&nb); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	b := Book{ID: s.nextBookID, Title: nb.Title, Author: nb.Author}
	s.books[s.nextBookID] = b
	s.nextBookID++
	writeJSON(w, http.StatusCreated, b)
}

func (s *bookstoreStore) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, exists := s.books[id]
	if !exists {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "bookstore-mock",
		Short: "In-memory mock server for the Bookstore API",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := newBookstoreStore()

			mux := http.NewServeMux()
			mux.HandleFunc("GET /books", store.listBooks)
			mux.HandleFunc("POST /books", store.createBook)
			mux.HandleFunc("GET /books/{id}", store.getBook)

			ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			srv := &http.Server{Handler: mux}
			go func() { _ = srv.Serve(ln) }()
			fmt.Printf("Listening on http://%s\n", ln.Addr())

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	}

	cmd.Flags().IntVar(&port, "port", 9090, "TCP port to listen on (0 = OS-assigned)")
	return cmd
}
