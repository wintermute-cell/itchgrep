package main

import (
	"fmt"
	"itchgrep/internal/cache"
	"itchgrep/internal/logging"
	"itchgrep/internal/web"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.Info("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func initializeCache() *cache.Cache {
	pageSizeStr := os.Getenv("PAGE_SIZE")
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	logging.Info("PAGE_SIZE: %v", pageSize)
	if err != nil {
		logging.Error("Invalid PAGE_SIZE, defaulting to 36: %s", pageSizeStr)
		pageSize = 36
	}
	c := cache.NewCache(pageSize)
	c.RefreshDataCache()
	return c
}

func main() {
	// LOGGING
	logging.Init("", true)

	// CACHE INIT
	cache := initializeCache()

	// HANDLERS
	r := chi.NewRouter()
	r.Use(logMiddleware)

	h := web.NewHandler(cache)
	r.Get("/", h.HandleIndex)
	r.Get("/assets/{page}", h.HandleGetAssetPage)
	r.Post("/query/{page}", h.HandleQuery)
	r.Get("/about", h.HandleAbout)

	// SERVER
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":8080" // Default port to listen on
	}
	logging.Info("Server started at port %s", port)

	http.ListenAndServe(port, r)
}
