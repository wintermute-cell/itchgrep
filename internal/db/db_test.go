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
