package web

import (
	"itchgrep/internal/cache"
	"itchgrep/internal/logging"
	"itchgrep/internal/web/templates"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type handler struct {
	cache *cache.Cache
}

func NewHandler(cache *cache.Cache) *handler {
	return &handler{
		cache: cache,
	}
}

func (h *handler) HandleHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func handle404(w http.ResponseWriter, r *http.Request) {
	component := templates.Layout("TODO", templates.Error404())
	component.Render(r.Context(), w)
}

func (h *handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		handle404(w, r)
	} else {
		component := templates.Layout("TODO", templates.Index())
		component.Render(r.Context(), w)
	}
}

func (h *handler) HandleGetAssetPage(w http.ResponseWriter, r *http.Request) {
	pageNum, err := strconv.ParseInt(chi.URLParam(r, "page"), 10, 64)
	if err != nil {
		logging.Error("Error parsing page: %s", err)
		http.Error(w, "Invalid request, page number not found", http.StatusBadRequest)
		return
	}
	assets, err := h.cache.Page(pageNum)
	if err != nil {
		logging.Error("Error fetching page: %s", err)
		http.Error(w, "Error fetching page", http.StatusBadRequest)
		return
	}

	component := templates.Layout("TODO", templates.AssetPage(pageNum, assets, false, ""))
	component.Render(r.Context(), w)
}

func (h *handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	if query == "" {
		// this shouldn't happen as long as the form is set up correctly
		http.Error(w, "Empty Query", http.StatusBadRequest)
		return
	}

	pageNum, err := strconv.ParseInt(chi.URLParam(r, "page"), 10, 64)
	if err != nil {
		logging.Error("Error parsing page: %s", err)
		http.Error(w, "Invalid request, page number not found", http.StatusBadRequest)
		return
	}

	assets, err := h.cache.QueryCache(query, pageNum)
	if err != nil {
		logging.Error("Error searching: %s", err)
		http.Error(w, "Error searching", http.StatusBadRequest)
		return
	}

	component := templates.Layout("TODO", templates.AssetPage(pageNum, assets, true, query))
	component.Render(r.Context(), w)
}
