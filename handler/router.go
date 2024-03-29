package handler

import (
	"encoding/json"
	"fmt"
	"github.com/foae/dimago/clients/cacoo"
	"github.com/foae/dimago/clients/github"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Config struct {
	GithubClient github.Communicator
	CacooClient  cacoo.Communicator
}

type Handler struct {
	githubClient github.Communicator
	cacooClient  cacoo.Communicator
}

func NewHandler(cfg Config) http.Handler {
	h := &Handler{
		githubClient: cfg.GithubClient,
		cacooClient:  cfg.CacooClient,
	}

	r := chi.NewRouter()
	r.Get("/", h.welcome)
	r.Post("/", h.fetchProject)

	return r
}

func (h *Handler) welcome(w http.ResponseWriter, r *http.Request) {
	m := struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{
		Message: `OK`,
		Status:  http.StatusOK,
	}

	rsp, err := json.Marshal(m)
	if err != nil {
		log.Printf("handler: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(m.Status)
	_, _ = fmt.Fprintf(w, "%s", rsp)
}

func (h *Handler) fetchProject(w http.ResponseWriter, r *http.Request) {
	var body struct{ URL string `json:"url"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.githubClient.RetrieveProject(body.URL); err != nil {
		log.Printf("could not fetch project: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _ = fmt.Fprintf(w, `OK`)
}
