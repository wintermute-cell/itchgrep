package storage

import (
	"crypto/rand"
	"itchgrep/pkg/models"
	"os"
	"testing"
	"time"

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

func TestGetAssetsUpdateTime(t *testing.T) {
	os.Setenv("RUN_LOCAL", "true")

	// Define a slice of Asset for testing
	testAssets := []models.Asset{
		{GameId: "1", Title: "Asset 1", Author: "Author 1", Description: "Description 1", Link: "http://example.com/1", ThumbUrl: "http://example.com/thumb1.jpg"},
		{GameId: "2", Title: "Asset 2", Author: "Author 2", Description: "Description 2", Link: "http://example.com/2", ThumbUrl: "http://example.com/thumb2.jpg"},
	}

	timePrePut := time.Now()
	time.Sleep(1 * time.Second)

	// Put Assets
	err := PutAssets(testAssets)
	require.NoError(t, err, "PutAssets should not fail")

	time.Sleep(1 * time.Second)
	timePostPut := time.Now()

	// Test GetAssetsUpdateTime
	updateTime, err := GetAssetsUpdateTime()
	require.NoError(t, err, "GetAssetsUpdateTime should not fail")

	assert.True(t, timePrePut.Before(updateTime), "Update time should be after the time of PutAssets")
	assert.True(t, timePostPut.After(updateTime), "Update time should be before the time of GetAssetsUpdateTime")
}

func TestPutAndGetFSWithSingleEmptyDirectory(t *testing.T) {
	os.Setenv("RUN_LOCAL", "true")

	nameInStorage := "testDirInStorage.gz.tar"

	testDir := t.TempDir()
	err := PutFS(testDir, nameInStorage)
	require.NoError(t, err, "PutFS should not fail")

	err = os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}

	outPath, err := GetFS(nameInStorage, ".")
	require.NoError(t, err, "GetFS should not fail")

	assert.DirExists(t, outPath, "Retrieved directory should exist")
	os.RemoveAll(outPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutAndGetFSWithDirectoryContainingFile(t *testing.T) {
	os.Setenv("RUN_LOCAL", "true")

	nameInStorage := "testDirInStorage.gz.tar"

	fileSizeMegabytes := 100
	fileContents := make([]byte, fileSizeMegabytes*1024*1024)
	rand.Read(fileContents)

	testDir := t.TempDir()
	testFile, err := os.Create(testDir + "/testFile")
	if err != nil {
		t.Fatal(err)
	}
	// fill file with data to reach the desired size
	if _, err := testFile.Write(fileContents); err != nil {
		t.Fatal(err)
	}
	// assert size
	fi, err := testFile.Stat()
	if err != nil || fi.Size() != int64(fileSizeMegabytes*1024*1024) {
		t.Fatal(err)
	}

	err = PutFS(testDir, nameInStorage)
	require.NoError(t, err, "PutFS should not fail")

	err = os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}

	outPath, err := GetFS(nameInStorage, ".")
	t.Cleanup(func() { os.RemoveAll(outPath) })
	require.NoError(t, err, "GetFS should not fail")

	assert.DirExists(t, outPath, "Retrieved directory should exist")

	// assert file size
	fi, err = os.Stat(outPath + "/testFile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(fileSizeMegabytes*1024*1024), fi.Size(), "Retrieved file size should match the original file size")

	// assert file contents
	retrievedFileContents, err := os.ReadFile(outPath + "/testFile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fileContents, retrievedFileContents, "Retrieved file contents should match the original file contents")
}

func TestPutAndGetFSWithMissingFile(t *testing.T) {
	os.Setenv("RUN_LOCAL", "true")

	nameInStorage := "testDirInStorage.gz.tar"

	err := PutFS("some/path/to/a/nonexistent/file", nameInStorage)
	require.NoError(t, err, "PutFS should not fail")

	outPath, err := GetFS(nameInStorage, ".") // this should simply not extract any files
	require.NoError(t, err, "GetFS should not fail")

	assert.Equal(t, "", outPath, "Retrieved path should be empty")
}
