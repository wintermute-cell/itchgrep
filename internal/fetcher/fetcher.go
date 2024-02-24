package fetcher

import (
	"encoding/json"
	"fmt"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type itchResponse struct {
	NumItems int64  `json:"num_items"`
	Page     int64  `json:"page"`
	Content  string `json:"content"`
}

func ParseAssetPage(respData itchResponse) ([]models.Asset, error) {
	// parse html
	queryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(respData.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// iterate over each asset
	assets := make([]models.Asset, 0)
	queryDoc.Find(".game_cell").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		gameId, _ := s.Attr("data-game_id")
		title := s.Find(".title").Text()
		author := s.Find(".game_author").Children().First().Text()
		description := s.Find(".game_text").Text()
		linkNode := s.Find(".thumb_link")
		link, _ := linkNode.Attr("href")
		thumbUrl, _ := linkNode.Children().First().Attr("data-lazy_src")
		assets = append(assets, models.Asset{
			GameId:      gameId,
			Title:       title,
			Author:      author,
			Description: description,
			Link:        link,
			ThumbUrl:    thumbUrl,
		})
	})
	return assets, nil
}

func FetchAssetPage(pageNum int64) (itchResponse, bool) {
	for {
		// Construct the URL with the page number
		url := fmt.Sprintf("https://itch.io/game-assets?page=%d&format=json", pageNum)
		resp, err := http.Get(url)
		if err != nil {
			logging.Error("Failed to fetch data: %v", err)
			return itchResponse{}, false
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var respData itchResponse
			if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
				logging.Error("Failed to decode response: %v", err)
				return itchResponse{}, false
			}
			return respData, true
		} else if resp.StatusCode == 429 {
			// Too Many Requests, wait a second and retry
			time.Sleep(time.Duration(rand.Int63n(7)+3) * time.Second)
			continue
		} else {
			logging.Error("Unexpected status code: %d %s", resp.StatusCode, resp.Status)
			resp.Body.Close() // Ensure response body is closed in this case
			return itchResponse{}, false
		}
	}
}

func GetAssetCount() (int64, error) {
	for {
		resp, err := http.Get("https://itch.io/game-assets")
		if err != nil {
			return 0, fmt.Errorf("failed to fetch page: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			queryDoc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return 0, fmt.Errorf("failed to parse HTML: %w", err)
			}

			// parse "(53,665 results)" -> 53665
			resultCountStr := queryDoc.Find(".game_count").Text()
			re := regexp.MustCompile(`[\d,]+`)
			match := re.FindString(resultCountStr)
			numberStr := strings.ReplaceAll(match, ",", "")
			number, err := strconv.ParseInt(numberStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse result count: %w", err)
			}
			return number, nil
		} else if resp.StatusCode == 429 {
			// Too Many Requests, wait a second and retry
			time.Sleep(4 * time.Second)
			continue
		} else {
			resp.Body.Close() // Ensure the response body is closed before returning
			return 0, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
		}
	}
}
