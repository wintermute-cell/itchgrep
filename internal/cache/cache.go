package cache

import (
	"errors"
	"itchgrep/internal/db"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/blevesearch/bleve"
)

type dataCache struct {
	data      []models.Asset
	updatedAt time.Time
}

type Cache struct {
	dataCache     dataCache
	cacheLifetime float64
	cacheLock     sync.RWMutex
	pageSize      int64
	dynamoClient  *dynamodb.Client
	index         bleve.Index
}

func NewCache(lifetime float64, pageSize int64, localDb bool) *Cache {
	dynamoClient, err := db.CreateDynamoClient(localDb)
	if err != nil {
		logging.Fatal("Failed to create DynamoDB client: %v", err)
	}

	return &Cache{
		dataCache:     dataCache{},
		cacheLifetime: lifetime,
		cacheLock:     sync.RWMutex{},
		pageSize:      pageSize,
		dynamoClient:  dynamoClient,
	}
}

func (c *Cache) IsCacheExpired() bool {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	return time.Since(c.dataCache.updatedAt).Hours() > c.cacheLifetime
}

func (c *Cache) InitIndex() error {
	indexMapping := bleve.NewIndexMapping()
	// Customize the index mapping as needed

	var err error
	c.index, err = bleve.NewMemOnly(indexMapping)
	return err
}

func (c *Cache) reIndexAssets(assets []models.Asset) (bleve.Index, error) {
	// Create a new index for the re-indexing process, so we don't break the
	// old one in case of errors
	newIndexMapping := bleve.NewIndexMapping() // TODO: customize as needed
	newIndex, err := bleve.NewMemOnly(newIndexMapping)
	if err != nil {
		return nil, err
	}

	// Index each asset in the new index
	for _, asset := range assets {
		if err := newIndex.Index(asset.GameId, asset); err != nil {
			newIndex.Close() // clean up the failed new index
			return nil, err
		}
	}

	return newIndex, nil
}

func (c *Cache) RefreshDataCache() error {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	newData, err := db.GetAllAssets(c.dynamoClient)
	if err != nil {
		return err
	}

	// TODO: we might want to check if the data has changed before re-indexing
	// and otherwise just return nil

	// Re-index the new data
	newIndex, err := c.reIndexAssets(newData)
	if err != nil {
		return err
	}

	if c.index != nil {
		c.index.Close() // close the old index
	}
	c.index = newIndex
	c.dataCache.data = newData
	c.dataCache.updatedAt = time.Now()
	return nil
}

func (c *Cache) QueryCache(queryString string, pageIndex int64) ([]models.Asset, error) {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	// TODO: construct a query that boosts matches on title, then description, then author
	query := bleve.NewQueryStringQuery(queryString)
	from := (int(pageIndex) - 1) * int(c.pageSize)
	searchRequest := bleve.NewSearchRequestOptions(query, int(c.pageSize), from, false) // TODO: adjust size as needed

	searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Fields = []string{"Title", "Author", "Description", "Link", "ThumbUrl"}

	searchResult, err := c.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var matchedAssets []models.Asset
	for _, hit := range searchResult.Hits {
		matchedAssets = append(matchedAssets, models.Asset{
			GameId:      hit.ID,
			Title:       hit.Fields["Title"].(string),
			Author:      hit.Fields["Author"].(string),
			Description: hit.Fields["Description"].(string),
			Link:        hit.Fields["Link"].(string),
			ThumbUrl:    hit.Fields["ThumbUrl"].(string),
		})
	}

	return matchedAssets, nil
}

func (c *Cache) Page(pageNum int64) ([]models.Asset, error) {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()
	start := pageNum * c.pageSize
	end := start + c.pageSize
	if start > int64(len(c.dataCache.data)) {
		return nil, errors.New("Page out of range")
	} else if end > int64(len(c.dataCache.data)) {
		end = int64(len(c.dataCache.data))
	}
	return c.dataCache.data[start:end], nil
}
