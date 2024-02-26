package storage

import (
	"itchgrep/pkg/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutAndGetAssets(t *testing.T) {
	os.Setenv("RUN_LOCAL", "true")

	// Define a slice of Asset for testing
	testAssets := []models.Asset{
		{GameId: "1", Title: "Asset 1", Author: "Author 1", Description: "Description 1", Link: "http://example.com/1", ThumbUrl: "http://example.com/thumb1.jpg"},
		{GameId: "2", Title: "Asset 2", Author: "Author 2", Description: "Description 2", Link: "http://example.com/2", ThumbUrl: "http://example.com/thumb2.jpg"},
	}

	// Test PutAssets
	err := PutAssets(testAssets)
	require.NoError(t, err, "PutAssets should not fail")

	// Test GetAssets
	retrievedAssets, err := GetAssets()
	require.NoError(t, err, "GetAssets should not fail")

	// Verify that the retrieved assets match the original test assets
	assert.Equal(t, testAssets, retrievedAssets, "Retrieved assets should match the original test assets")
}
