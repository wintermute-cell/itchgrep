package models

import "fmt"

// Asset represents a game asset.
// Assets are stored in the DynamoDB table.
type Asset struct {
	GameId      string
	Title       string
	Author      string
	Description string
	Link        string
	ThumbUrl    string
}

func (a Asset) String() string {
	return fmt.Sprintf("GameId: %s\nName: %s\nAuthor: %s\nDescription: %s\nLink: %s\nThumbUrl: %s\n", a.GameId, a.Title, a.Author, a.Description, a.Link, a.ThumbUrl)
}
