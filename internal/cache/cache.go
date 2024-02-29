package cache

import (
	"errors"
	"itchgrep/internal/logging"
	"itchgrep/internal/storage"
	"itchgrep/pkg/models"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

type Cache struct {
	cacheLock sync.RWMutex

	dataMap map[string]models.Asset
	data    []models.Asset
	index   bleve.Index

	// the time the data was last updated on the server.
	// if we check if the current time is greater than this time, we know the
	// cache is expired
	dataUpdatedTime time.Time

	// the cache can be retrieved as chunks/pages
	pageSize int64
}

func NewCache(pageSize int64) *Cache {
	return &Cache{
		dataMap:         make(map[string]models.Asset),
		cacheLock:       sync.RWMutex{},
		pageSize:        pageSize,
		dataUpdatedTime: time.Time{},
	}
}

func (c *Cache) IsCacheExpired() bool {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	// if we never updated the cache, it is expired
	if c.dataUpdatedTime.IsZero() {
		return true
	}

	// otherwise, we check if the data on the server is newer than the data in the cache
	storageUpdateTime, err := storage.GetAssetsUpdateTime()
	if err != nil {
		logging.Error("Failed to get assets update time: %v", err)
		return false
	}
	return c.dataUpdatedTime.Before(storageUpdateTime)
}

func (c *Cache) RefreshDataCache() error {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	// we fetch this here already, since we can just stop if we fail to fetch even this
	newServerUpdateTime, err := storage.GetAssetsUpdateTime()
	if err != nil {
		return err
	}

	// fetch asset data
	preFetchTime := time.Now()
	newData, err := storage.GetAssets()
	if err != nil || newData == nil {
		return err
	}
	fetchTime := time.Since(preFetchTime)
	logging.Info("Fetched %d assets in %v", len(newData), fetchTime)

	// TODO: this whole procedure is pretty unsafe. we should not close the
	// index before we are sure we have a new one to replace it with.
	// But most likely we need to close the old index before we can open a new one at the same path.
	// We should probably use a temporary path for the new index and then move it to the correct path.

	// fetch index data
	preFetchTime = time.Now()
	if c.index != nil {
		c.index.Close()
	}
	indexPath, err := storage.GetFS(storage.IndexArchiveName, ".")
	c.index, err = bleve.Open(indexPath)
	if err != nil {
		return err
	}

	fetchTime = time.Since(preFetchTime)
	logging.Info("Fetched and opened index in %v", fetchTime)

	// overwrite the old data with the new data
	c.data = newData
	c.dataMap = make(map[string]models.Asset, len(newData)) // we also save it as a map, so we can easily match searches from the index
	for _, asset := range newData {
		c.dataMap[asset.GameId] = asset
	}
	c.dataUpdatedTime = newServerUpdateTime
	return nil
}

func buildFuzzyQuery(queryString string, fuzzyness int, prefixLen int) *query.DisjunctionQuery {
	titleQuery := bleve.NewMatchQuery(queryString)
	titleQuery.SetField("Title")
	titleQuery.SetBoost(3)
	titleQuery.SetPrefix(prefixLen)
	titleQuery.SetFuzziness(fuzzyness)
	descriptionQuery := bleve.NewMatchQuery(queryString)
	descriptionQuery.SetField("Description")
	descriptionQuery.SetBoost(2)
	descriptionQuery.SetPrefix(prefixLen)
	descriptionQuery.SetFuzziness(fuzzyness)
	authorQuery := bleve.NewMatchQuery(queryString)
	authorQuery.SetField("Author")
	authorQuery.SetBoost(1)
	authorQuery.SetPrefix(prefixLen)
	authorQuery.SetFuzziness(fuzzyness)

	// Combine queries with a disjunction (OR) query
	query := bleve.NewDisjunctionQuery(titleQuery, descriptionQuery, authorQuery)
	return query
}

func buildExactQuery(queryString string) *query.DisjunctionQuery {
	titleQuery := bleve.NewMatchQuery(queryString)
	titleQuery.SetField("Title")
	titleQuery.SetBoost(3)
	descriptionQuery := bleve.NewMatchQuery(queryString)
	descriptionQuery.SetField("Description")
	descriptionQuery.SetBoost(2)
	authorQuery := bleve.NewMatchQuery(queryString)
	authorQuery.SetField("Author")
	authorQuery.SetBoost(1)

	// Combine queries with a disjunction (OR) query
	query := bleve.NewDisjunctionQuery(titleQuery, descriptionQuery, authorQuery)
	return query
}

func (c *Cache) QueryCache(queryString string, pageIndex int64) ([]models.Asset, error) {
	// check for stale cache, refresh if needed
	if c.IsCacheExpired() {
		if err := c.RefreshDataCache(); err != nil {
			return nil, err
		}
	}

	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	veryFuzzyQuery := buildFuzzyQuery(queryString, 1, 2)
	veryFuzzyQuery.SetBoost(2)
	fuzzyQuery := buildFuzzyQuery(queryString, 1, 4)
	fuzzyQuery.SetBoost(4)
	exactQuery := buildExactQuery(queryString)
	exactQuery.SetBoost(6)
	query := bleve.NewDisjunctionQuery(veryFuzzyQuery, fuzzyQuery, exactQuery)

	from := (int(pageIndex) - 1) * int(c.pageSize)
	searchRequest := bleve.NewSearchRequestOptions(query, int(c.pageSize), from, false)

	//searchRequest.Highlight = bleve.NewHighlight()
	searchRequest.Fields = []string{"Title", "Author", "Description"}
	searchRequest.SortBy([]string{"-_score", "InvPopularity"})

	searchResult, err := c.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	logging.Info("Got %d hits for query \"%s\"", searchResult.Total, queryString)

	var matchedAssets []models.Asset
	for _, hit := range searchResult.Hits {
		matchedAssets = append(matchedAssets, c.dataMap[hit.ID])
	}

	return matchedAssets, nil
}

func (c *Cache) Page(pageNum int64) ([]models.Asset, error) {

	// TODO: maybe we dont even have to check for a stale cache, since most
	// people won't be using the page function a lot

	// check for stale cache, refresh if needed
	if c.IsCacheExpired() {
		if err := c.RefreshDataCache(); err != nil {
			return nil, err
		}
	}

	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()
	start := pageNum * c.pageSize
	end := start + c.pageSize
	if start > int64(len(c.data)) {
		return nil, errors.New("Page out of range")
	} else if end > int64(len(c.data)) {
		end = int64(len(c.data))
	}
	return c.data[start:end], nil
}
