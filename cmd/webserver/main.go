package main

import (
	"fmt"
	"itchgrep/internal/cache"
	"itchgrep/internal/logging"
	"itchgrep/internal/web"
	"net/http"
	"os"
	"runtime/debug"
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
	c := cache.NewCache(cacheLifetime, pageSize)
	c.RefreshDataCache()
	return c
}

func setMemoryLimit() {
	memoryLimitStr := os.Getenv("MEMORY_LIMIT_MB")
	memoryLimit, err := strconv.ParseInt(memoryLimitStr, 10, 64)
	if err != nil {
		logging.Error("Invalid MEMORY_LIMIT_MB, defaulting to 1024MB: %s", memoryLimitStr)
		memoryLimit = 1024
	}
	logging.Info("Setting Memory Limit to: %v MB", memoryLimit)
	debug.SetMemoryLimit(1024 * 1024 * memoryLimit) // 512MB
}

func main() {
	// LOGGING
	logging.Init("", true)

	// MEMORY LIMIT
	setMemoryLimit()

	// CACHE INIT
	cache := initializeCache()

	// HANDLERS
	r := chi.NewRouter()
	r.Use(logMiddleware)

	h := web.NewHandler(cache)
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
