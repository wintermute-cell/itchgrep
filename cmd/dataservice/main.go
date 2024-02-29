package main

import (
	"fmt"
	"itchgrep/internal/fetcher"
	"itchgrep/internal/logging"
	"itchgrep/internal/storage"
	"itchgrep/pkg/models"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blevesearch/bleve"
)

func main() {
	logging.Init("", true)

	http.HandleFunc("/trigger-fetch", handleFetchTrigger)
	port := fmt.Sprintf(":%s", os.Getenv("PORT")) // as per cloud run standard
	if port == ":" {
		port = ":8080"
	}
	logging.Info("Server listening on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		logging.Fatal("Server failed to start: %v", err)
	}
}

func handleFetchTrigger(w http.ResponseWriter, r *http.Request) {
	// Ensure that we only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		fetchAndStoreAssets()
	}()

	// Respond to indicate the process has started
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Asset fetch and store initiated")
}

func fetchAndStoreAssets() {
	// FETCHING ASSETS
	assetCount, err := fetcher.GetAssetCount()
	if err != nil {
		logging.Fatal("Failed to get asset count: %v", err)
	}

	// fetch the first page to get the number of items per page
	respData, ok := fetcher.FetchAssetPage(1)
	if !ok {
		logging.Fatal("Failed to fetch first page, terminating.")
	}

	nPages := int64(math.Ceil(float64(assetCount) / float64(respData.NumItems)))

	var wg sync.WaitGroup
	assetsChan := make(chan []models.Asset, int(nPages))

	var pagesFetched atomic.Int64
	var pagesInProgress atomic.Int64

	for i := int64(1); i <= nPages; i++ {
		wg.Add(1)
		go func(pageNum int64) {
			defer pagesFetched.Add(1)
			defer pagesInProgress.Add(-1)
			defer wg.Done()
			pagesInProgress.Add(1)
			time.Sleep(time.Second * time.Duration(rand.Int63n(nPages/9+1))) // this spreads out the requests
			data, ok := fetcher.FetchAssetPage(pageNum)
			if !ok {
				return
			}
			assets, err := fetcher.ParseAssetPage(data, pageNum) // we include pageNum, as it indicates popularity
			if err != nil {
				logging.Error("Failed to parse asset page: %v", err)
				return
			}
			assetsChan <- assets
		}(i)
	}

	// every 5 seconds, print the progress
	quitProgressLog := make(chan bool)
	go func() {
		for {
			select {
			case <-quitProgressLog:
				return
			default:
				time.Sleep(5 * time.Second)
				logging.Info("Pages fetched: %d/%d, in progress: %d", pagesFetched.Load(), nPages, pagesInProgress.Load())
			}
		}
	}()

	// close the channel when all the assets are fetched.
	// we do this in a goroutine so that we don't block the main thread
	go func() {
		wg.Wait()
		quitProgressLog <- true
		close(assetsChan)
	}()

	var assets []models.Asset
	for assetPage := range assetsChan {
		assets = append(assets, assetPage...)
	}
	logging.Info("Successfully fetched %d assets", len(assets))

	// CREATING INDEX
	logging.Info("Creating index...")
	newIndexMapping := bleve.NewIndexMapping() // TODO: customize as needed
	newIndex, err := bleve.New(storage.IndexDirName, newIndexMapping)
	defer os.RemoveAll(storage.IndexDirName) // After we are done, no matter if clean or with error, we remove the index, since it is uploaded to storage.
	if err != nil {
		logging.Error("Failed to create index: %v", err)
		return
	}

	logging.Info("Created new empty index at %s", storage.IndexDirName)

	// first, convert the assets to IndexedAssets, which are smaller and used for indexing
	var smolAssets []models.IndexedAsset = make([]models.IndexedAsset, len(assets))
	for i, asset := range assets {
		smolAssets[i] = models.IndexedAsset{
			GameId:        asset.GameId,
			Title:         asset.Title,
			Author:        asset.Author,
			Description:   asset.Description,
			InvPopularity: asset.InvPopularity,
		}
	}

	// indexing the assets in batches
	b := newIndex.NewBatch()
	assetsIndexed := 0
	for i, asset := range smolAssets {
		assetsIndexed += 1
		err := b.Index(asset.GameId, asset)
		if err != nil {
			newIndex.Close() // clean up the failed new index
			logging.Error("Failed to index asset, cancelling indexing: %v", err)
			return
		}
		if i%1500 == 0 && i != 0 { // we index in batches of 1500
			logging.Info("Batching assets: %d/%d", i, len(smolAssets))
			newIndex.Batch(b)
			b.Reset()
		}
	}
	newIndex.Batch(b) // batch the remaining assets into the index
	newIndex.Close()  // close the index
	logging.Info("Successfully indexed %d assets", assetsIndexed)

	// STORING INDEX
	logging.Info("Storing index in cloud storage file")
	dir, err := os.ReadDir(storage.IndexDirName)
	if err != nil {
		logging.Error("Failed to read dir: %v", err)
	}
	for _, entry := range dir {
		logging.Info("DEBUG: entry: %s", entry.Name())
	}

	err = storage.PutFS(storage.IndexDirName, storage.IndexArchiveName)
	if err != nil {
		logging.Error("Failed to put index: %v", err)
		return
	}
	logging.Info("Successfully stored index")

	// STORING ASSETS
	logging.Info("Storing assets in cloud storage file")
	err = storage.PutAssets(assets)
	if err != nil {
		logging.Error("Failed to put assets, stopping here and not putting index: %v", err)
		return
	}
	logging.Info("Successfully stored assets")

}
