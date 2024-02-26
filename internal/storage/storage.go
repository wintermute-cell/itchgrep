package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	bucketName = "itchgrep-data"
	fileName   = "assets.json"
)

func createClient(ctx context.Context) (*storage.Client, error) {
	local := os.Getenv("RUN_LOCAL") == "true"
	logging.Info("RUN_LOCAL: %v", local)

	if local {
		os.Setenv("STORAGE_EMULATOR_HOST", "http://fake-gcs-server:4443")
		return storage.NewClient(
			ctx,
			option.WithEndpoint("http://fake-gcs-server:4443/storage/v1/"),
			storage.WithJSONReads())
	} else {
		return storage.NewClient(ctx)
	}
}

// PutAssets writes the provided assets to a Google Cloud Storage bucket as a JSON file.
func PutAssets(assets []models.Asset) error {
	ctx := context.Background()

	client, err := createClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(bucketName)

	// Convert assets slice to JSON
	assetsJSON, err := json.Marshal(assets)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	// Create a new writer to write the JSON data
	obj := bkt.Object(fileName)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(assetsJSON); err != nil {
		return fmt.Errorf("Writer.Write: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

// GetAssets fetches the assets JSON file from the Google Cloud Storage bucket and unmarshals it into a slice of Assets.
func GetAssets() ([]models.Asset, error) {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(fileName)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object.NewReader: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	var assets []models.Asset
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %v", err)
	}

	return assets, nil
}
