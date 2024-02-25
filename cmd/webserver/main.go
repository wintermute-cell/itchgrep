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
	cacheLifetimeStr := os.Getenv("CACHE_LIFETIME_HOURS")
	cacheLifetime, err := strconv.ParseFloat(cacheLifetimeStr, 64)
	logging.Info("CACHE_LIFETIME_HOURS: %v", cacheLifetime)
	if err != nil {
		logging.Error("Invalid CACHE_LIFETIME_HOURS, defaulting to 24h: %s", cacheLifetimeStr)
		cacheLifetime = 24
	}
	pageSizeStr := os.Getenv("PAGE_SIZE")
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	logging.Info("PAGE_SIZE: %v", pageSize)
	if err != nil {
		logging.Error("Invalid PAGE_SIZE, defaulting to 36: %s", pageSizeStr)
		pageSize = 36
	}
	localDb := os.Getenv("DYNAMO_LOCAL") == "true"
	logging.Info("DYNAMO_LOCAL: %v", localDb)
	c := cache.NewCache(cacheLifetime, pageSize, localDb)
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
	h := web.NewHandler(cache)
	r.Use(logMiddleware)
	r.Get("/", h.HandleIndex)
	r.Get("/assets/{page}", h.HandleGetAssetPage)
	r.Post("/query/{page}", h.HandleQuery)

	// SERVER
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":8080" // Default port to listen on
	}
	logging.Info("Server started at port %s", port)
	http.ListenAndServe(port, r)
}
