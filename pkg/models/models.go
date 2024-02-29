package models

import "fmt"

// Asset represents a game asset.
// Assets are stored in the DynamoDB table.
type Asset struct {
	GameId        string
	Title         string
	Author        string
	Description   string
	Link          string
	ThumbUrl      string
	InvPopularity int64 // inverse popularity, derived from page number of the asset
}

func (a Asset) String() string {
	return fmt.Sprintf("GameId: %s, Title: %s, Author: %s, Description: %s, Link: %s, ThumbUrl: %s, InvPopularity: %d", a.GameId, a.Title, a.Author, a.Description, a.Link, a.ThumbUrl, a.InvPopularity)
}

// IndexedAsset is a smaller version of Asset, used for indexing.
type IndexedAsset struct {
	GameId        string
	Title         string
	Author        string
	Description   string
	InvPopularity int64
}

func (a IndexedAsset) String() string {
	return fmt.Sprintf("GameId: %s, Title: %s, Author: %s, Description: %s, InvPopularity: %d", a.GameId, a.Title, a.Author, a.Description, a.InvPopularity)
}
