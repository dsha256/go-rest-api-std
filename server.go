package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
)

// Server is the album HTTP server.
type Server struct {
	db  Database
	log *log.Logger
}

// NewServer creates a new server using the given database implementation.
func NewServer(db Database, log *log.Logger) *Server {
	return &Server{db: db, log: log}
}

// Regex to match "/albums/:id" (id must be one or more non-slash chars).
var reAlbumsID = regexp.MustCompile(`^/albums/([^/]+)$`)

// ServeHTTP routes the request and calls the correct handler based on the URL
// and HTTP method. It writes a 404 Not Found if the request URL is unknown,
// or 405 Method Not Allowed if the request method is invalid.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	s.log.Printf("%s %s", r.Method, path)

	var id string

	switch {
	case path == "/albums":
		switch r.Method {
		case "GET":
			s.getAlbums(w, r)
		case "POST":
			s.addAlbum(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
		}

	case match(path, reAlbumsID, &id):
		switch r.Method {
		case "GET":
			s.getAlbumByID(w, r, id)
		default:
			w.Header().Set("Allow", "GET")
			s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
		}

	default:
		s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
	}
}

func (s *Server) getAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := s.db.GetAlbums()
	if err != nil {
		s.log.Printf("error fetching albums: %v", err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, albums)
}

func (s *Server) addAlbum(w http.ResponseWriter, r *http.Request) {
	var album Album
	if !s.readJSON(w, r, &album) {
		return
	}

	// Validate the input and build a map of validation issues
	type validationIssue struct {
		Error   string `json:"error"`
		Message string `json:"message,omitempty"`
	}
	issues := make(map[string]any)
	if album.ID == "" {
		issues["id"] = validationIssue{"required", ""}
	}
	if album.Title == "" {
		issues["title"] = validationIssue{"required", ""}
	}
	if album.Artist == "" {
		issues["artist"] = validationIssue{"required", ""}
	}
	if album.Price < 0 || album.Price >= 100000 {
		issues["price"] = validationIssue{"out-of-range", "price must be between 0 and $1000"}
	}
	if len(issues) > 0 {
		s.jsonError(w, http.StatusBadRequest, ErrorValidation, issues)
		return
	}

	err := s.db.AddAlbum(album)
	if errors.Is(err, ErrAlreadyExists) {
		s.jsonError(w, http.StatusConflict, ErrorAlreadyExists, nil)
		return
	} else if err != nil {
		s.log.Printf("error adding album ID %q: %v", album.ID, err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}

	s.writeJSON(w, http.StatusCreated, album)
}

func (s *Server) getAlbumByID(w http.ResponseWriter, r *http.Request, id string) {
	album, err := s.db.GetAlbumByID(id)
	if errors.Is(err, ErrDoesNotExist) {
		s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
		return
	} else if err != nil {
		s.log.Printf("error fetching album ID %q: %v", id, err)
		s.jsonError(w, http.StatusInternalServerError, ErrorDatabase, nil)
		return
	}
	s.writeJSON(w, http.StatusOK, album)
}

// writeJSON marshals v to JSON and writes it to the response, handling
// errors as appropriate. It also sets the Content-Type header to
// "application/json".
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		s.log.Printf("error marshaling JSON: %v", err)
		http.Error(w, `{"error":"`+ErrorInternal+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		// Very unlikely to happen, but log any error (not much more we can do)
		s.log.Printf("error writing JSON: %v", err)
	}
}

// jsonError writes a structured error as JSON to the response, with
// optional structured data in the "data" field.
func (s *Server) jsonError(w http.ResponseWriter, status int, error string, data map[string]any) {
	response := struct {
		Status int            `json:"status"`
		Error  string         `json:"error"`
		Data   map[string]any `json:"data,omitempty"`
	}{
		Status: status,
		Error:  error,
		Data:   data,
	}
	s.writeJSON(w, status, response)
}

// readJSON reads the request body and unmarshal it from JSON, handling
// errors as appropriate. It returns true on success; the caller should
// return from the handler early if it returns false.
func (s *Server) readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Printf("error reading JSON body: %v", err)
		s.jsonError(w, http.StatusInternalServerError, ErrorInternal, nil)
		return false
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		data := map[string]any{"message": err.Error()}
		s.jsonError(w, http.StatusBadRequest, ErrorMalformedJSON, data)
		return false
	}
	return true
}
