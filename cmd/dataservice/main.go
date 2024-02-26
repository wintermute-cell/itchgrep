package main

import (
	"itchgrep/internal/fetcher"
	"itchgrep/internal/logging"
	"itchgrep/internal/storage"
	"itchgrep/pkg/models"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	logging.Init("", true)

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
			time.Sleep(time.Second * time.Duration(rand.Int63n(nPages/18+1)))
			data, ok := fetcher.FetchAssetPage(pageNum)
			if !ok {
				return
			}
			assets, err := fetcher.ParseAssetPage(data)
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

	err = storage.PutAssets(assets)
	if err != nil {
		logging.Error("Failed to put asset: %v", err)
	}
}
