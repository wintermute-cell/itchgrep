package db

import (
	"itchgrep/pkg/models"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Ensure tests run against DynamoDB Local
	os.Setenv("DYNAMO_LOCAL", "true")

	// Setup code (e.g., create client, create table)
	svc, err := CreateDynamoClient(true)
	if err != nil {
		panic("Failed to create DynamoDB client: " + err.Error())
	}

	err = CrateAssetsTableIfNotExists(svc)
	if err != nil {
		panic("Failed to create Assets table: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Teardown code (e.g., delete table)

	os.Exit(code)
}

func TestPutAndGetAsset(t *testing.T) {
	svc, _ := CreateDynamoClient(true)

	tests := []struct {
		name    string
		asset   models.Asset
		wantErr bool
	}{
		{
			name: "ValidAsset",
			asset: models.Asset{
				GameId:      "test1",
				Title:       "Test Asset",
				Author:      "Test Author",
				Description: "Test Description",
				Link:        "https://example.com",
				ThumbUrl:    "https://example.com/thumb.jpg",
			},
			wantErr: false,
		},
		{
			name: "ValidAssetMissingFields",
			asset: models.Asset{
				GameId:      "test1",
				Title:       "",
				Author:      "",
				Description: "",
				Link:        "",
				ThumbUrl:    "",
			},
			wantErr: false,
		},
		{
			name: "ValidAssetMissingId",
			asset: models.Asset{
				GameId:      "",
				Title:       "",
				Author:      "",
				Description: "",
				Link:        "",
				ThumbUrl:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test PutAsset
			err := PutAsset(svc, tt.asset)
			if tt.wantErr {
				require.Error(t, err, "PutAsset expected to fail")
			} else {
				require.NoError(t, err, "PutAsset failed unexpectedly")

				// Test GetAsset only if put was successful
				asset, err := GetAsset(svc, tt.asset.GameId)
				require.NoError(t, err, "GetAsset failed unexpectedly")

				// Verify the asset
				require.Equal(t, tt.asset.GameId, asset.GameId)
				require.Equal(t, tt.asset.Title, asset.Title)
				require.Equal(t, tt.asset.Author, asset.Author)
				require.Equal(t, tt.asset.Description, asset.Description)
				require.Equal(t, tt.asset.Link, asset.Link)
				require.Equal(t, tt.asset.ThumbUrl, asset.ThumbUrl)
			}
		})
	}
}

func TestPutAssets(t *testing.T) {
	svc, err := CreateDynamoClient(true)
	require.NoError(t, err, "Failed to create DynamoDB client")

	tests := []struct {
		name    string
		assets  []models.Asset
		wantErr bool
	}{
		{
			name: "ValidAssets",
			assets: []models.Asset{
				{
					GameId:      "test1",
					Title:       "Test Asset 1",
					Author:      "Test Author 1",
					Description: "Test Description 1",
					Link:        "https://example.com/1",
					ThumbUrl:    "https://example.com/thumb1.jpg",
				},
				{
					GameId:      "test2",
					Title:       "Test Asset 2",
					Author:      "Test Author 2",
					Description: "Test Description 2",
					Link:        "https://example.com/2",
					ThumbUrl:    "https://example.com/thumb2.jpg",
				},
			},
			wantErr: false,
		},
		{
			name:    "EmptyAssetList",
			assets:  []models.Asset{},
			wantErr: true, // Assuming you handle empty slices as an error
		},
		{
			name: "ValidAndInvalidAsset",
			assets: []models.Asset{
				{
					GameId:      "test3",
					Title:       "Test Asset 3",
					Author:      "Test Author 3",
					Description: "Test Description 3",
					Link:        "https://example.com/3",
					ThumbUrl:    "https://example.com/thumb3.jpg",
				},
				{
					GameId:      "", // Invalid due to missing GameId
					Title:       "Invalid Asset",
					Author:      "Invalid Author",
					Description: "Invalid Description",
					Link:        "https://example.com/invalid",
					ThumbUrl:    "https://example.com/thumbinvalid.jpg",
				},
			},
			wantErr: true,
		},
		{
			name: "DuplicateIDs",
			assets: []models.Asset{
				{
					GameId:      "test-duplicate",
					Title:       "Test Asset Duplicate 1",
					Author:      "Test Author Duplicate 1",
					Description: "Test Description Duplicate 1",
					Link:        "https://example.com/duplicate1",
					ThumbUrl:    "https://example.com/thumbduplicate1.jpg",
				},
				{
					GameId:      "test-duplicate", // Intentional duplicate ID
					Title:       "Test Asset Duplicate 2",
					Author:      "Test Author Duplicate 2",
					Description: "Test Description Duplicate 2",
					Link:        "https://example.com/duplicate2",
					ThumbUrl:    "https://example.com/thumbduplicate2.jpg",
				},
			},
			wantErr: true, // Expect an error due to duplicate GameId
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PutAssets(svc, tt.assets)
			if tt.wantErr {
				require.Error(t, err, "PutAssets expected to fail")
			} else {
				require.NoError(t, err, "PutAssets failed unexpectedly")

				// If PutAssets was successful, verify each asset was correctly stored
				if !tt.wantErr {
					for _, asset := range tt.assets {
						retrievedAsset, err := GetAsset(svc, asset.GameId)
						require.NoError(t, err, "GetAsset failed unexpectedly")
						require.Equal(t, asset, retrievedAsset, "Retrieved asset does not match the original")
					}
				}
			}
		})
	}
}
