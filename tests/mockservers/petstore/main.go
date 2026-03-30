package main

import (
	"bytes"
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

type Pet struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Species string `json:"species"`
	Age     *int   `json:"age,omitempty"`
}

type newPet struct {
	Name    string `json:"name"`
	Species string `json:"species"`
	Age     *int   `json:"age,omitempty"`
}

type petPatch struct {
	Name *string `json:"name,omitempty"`
	Age  *int    `json:"age,omitempty"`
}

type jsonPatchOp struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value,omitempty"`
	From  string          `json:"from,omitempty"`
}

type Owner struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type errorBody struct {
	Message string `json:"message"`
}

type petstoreStore struct {
	mu        sync.RWMutex
	pets      map[int]Pet
	nextPetID int
	owners    map[int]Owner
}

func newPetstoreStore() *petstoreStore {
	s := &petstoreStore{
		pets:      make(map[int]Pet),
		nextPetID: 1,
		owners:    make(map[int]Owner),
	}
	s.owners[1] = Owner{ID: 1, Name: "Alice"}
	s.owners[2] = Owner{ID: 2, Name: "Bob"}
	return s
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Message: msg})
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func (s *petstoreStore) listPets(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 {
			limit = n
		}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	pets := make([]Pet, 0, len(s.pets))
	for _, p := range s.pets {
		pets = append(pets, p)
		if limit > 0 && len(pets) >= limit {
			break
		}
	}
	writeJSON(w, http.StatusOK, pets)
}

func (s *petstoreStore) createPet(w http.ResponseWriter, r *http.Request) {
	var np newPet
	if err := json.NewDecoder(r.Body).Decode(&np); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := Pet{ID: s.nextPetID, Name: np.Name, Species: np.Species, Age: np.Age}
	s.pets[s.nextPetID] = p
	s.nextPetID++
	writeJSON(w, http.StatusCreated, p)
}

func (s *petstoreStore) getPet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.pets[id]
	if !exists {
		writeError(w, http.StatusNotFound, "pet not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *petstoreStore) updatePet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var np newPet
	if err := json.NewDecoder(r.Body).Decode(&np); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pets[id]; !exists {
		writeError(w, http.StatusNotFound, "pet not found")
		return
	}
	p := Pet{ID: id, Name: np.Name, Species: np.Species, Age: np.Age}
	s.pets[id] = p
	writeJSON(w, http.StatusOK, p)
}

func (s *petstoreStore) patchPet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	p, exists := s.pets[id]
	if !exists {
		writeError(w, http.StatusNotFound, "pet not found")
		return
	}

	if r.ContentLength != 0 {
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		trimmed := bytes.TrimLeft(raw, " \t\r\n")
		if len(trimmed) > 0 && trimmed[0] == '[' {
			// RFC 6902 JSON Patch
			var ops []jsonPatchOp
			if err := json.Unmarshal(raw, &ops); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON Patch body")
				return
			}
			for _, op := range ops {
				if op.Op == "replace" {
					switch op.Path {
					case "/name":
						var v string
						if err := json.Unmarshal(op.Value, &v); err == nil {
							p.Name = v
						}
					case "/age":
						var v int
						if err := json.Unmarshal(op.Value, &v); err == nil {
							p.Age = &v
						}
					}
				}
			}
		} else {
			// JSON Merge Patch
			var patch petPatch
			if err := json.Unmarshal(raw, &patch); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			if patch.Name != nil {
				p.Name = *patch.Name
			}
			if patch.Age != nil {
				p.Age = patch.Age
			}
		}
	}

	s.pets[id] = p
	writeJSON(w, http.StatusOK, p)
}

func (s *petstoreStore) deletePet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pets[id]; !exists {
		writeError(w, http.StatusNotFound, "pet not found")
		return
	}
	delete(s.pets, id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *petstoreStore) listOwners(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	owners := make([]Owner, 0, len(s.owners))
	for _, o := range s.owners {
		owners = append(owners, o)
	}
	writeJSON(w, http.StatusOK, owners)
}

func (s *petstoreStore) getOwner(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, exists := s.owners[id]
	if !exists {
		writeError(w, http.StatusNotFound, "owner not found")
		return
	}
	writeJSON(w, http.StatusOK, o)
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
		Use:   "petstore-mock",
		Short: "In-memory mock server for the Petstore API",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := newPetstoreStore()

			mux := http.NewServeMux()
			mux.HandleFunc("GET /pets", store.listPets)
			mux.HandleFunc("POST /pets", store.createPet)
			mux.HandleFunc("GET /pets/{id}", store.getPet)
			mux.HandleFunc("PUT /pets/{id}", store.updatePet)
			mux.HandleFunc("PATCH /pets/{id}", store.patchPet)
			mux.HandleFunc("DELETE /pets/{id}", store.deletePet)
			mux.HandleFunc("GET /owners", store.listOwners)
			mux.HandleFunc("GET /owners/{id}", store.getOwner)

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

	cmd.Flags().IntVar(&port, "port", 8080, "TCP port to listen on (0 = OS-assigned)")
	return cmd
}
