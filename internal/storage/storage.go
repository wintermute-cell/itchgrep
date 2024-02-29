package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itchgrep/internal/logging"
	"itchgrep/pkg/models"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/mholt/archiver/v4"
	"google.golang.org/api/option"
)

const (
	BucketName       = "itchgrep-data"
	DataFileName     = "assets.json"
	IndexDirName     = "index.bleve"
	IndexArchiveName = "index.bleve.gz.tar"
)

var ArchiveFormat = archiver.CompressedArchive{
	Compression: archiver.Gz{},
	Archival:    archiver.Tar{},
}

func createClient(ctx context.Context) (*storage.Client, error) {
	local := os.Getenv("RUN_LOCAL") == "true"
	logging.Info("RUN_LOCAL: %v", local)
	test := os.Getenv("RUN_TEST") == "true"
	logging.Info("RUN_TEST: %v", test)

	if local {
		if !test {
			os.Setenv("STORAGE_EMULATOR_HOST", "http://fake-gcs-server:4443") // name of the docker container
			logging.Info("Using address: http://fake-gcs-server:4443")
			return storage.NewClient(
				ctx,
				option.WithEndpoint("http://fake-gcs-server:4443/storage/v1/"),
				storage.WithJSONReads())
		} else { // if we are running tests, this is not running in a container
			os.Setenv("STORAGE_EMULATOR_HOST", "http://localhost:4443")
			logging.Info("Using address: http://localhost:4443")
			return storage.NewClient(
				ctx,
				option.WithEndpoint("http://localhost:4443/storage/v1/"),
				storage.WithJSONReads())
		}
	} else {
		logging.Info("Using production GCS client.")
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

	bkt := client.Bucket(BucketName)

	// Convert assets slice to JSON
	assetsJSON, err := json.Marshal(assets)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	// Create a new writer to write the JSON data
	obj := bkt.Object(DataFileName)
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

	bkt := client.Bucket(BucketName)
	obj := bkt.Object(DataFileName)
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

func GetAssetsUpdateTime() (time.Time, error) {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(BucketName)
	objAttrs, err := bkt.Object(DataFileName).Attrs(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("Object.Attrs: %v", err)
	}

	return objAttrs.Updated, nil
}

// PutFS writes the provided directory or file to a Google Cloud Storage
// bucket as a compressed archive.
func PutFS(dirPath, nameInStorage string) error {
	ctx := context.Background()

	client, err := createClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(BucketName)

	// COMPRESSING INDEX DIRECTORY
	fileMapping, _ := archiver.FilesFromDisk(nil, map[string]string{
		dirPath: filepath.Base(dirPath),
	})

	// create temp dir for the archive file to be created in
	archiveFileHandle, err := os.Create(nameInStorage)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer os.RemoveAll(nameInStorage)

	err = ArchiveFormat.Archive(context.Background(), archiveFileHandle, fileMapping)
	if err != nil {
		return fmt.Errorf("format.Archive: %v", err)
	}
	archiveFileHandle.Close()

	// load zip file as bytes
	archiveBytes, err := os.ReadFile(nameInStorage)
	if err != nil {
		return fmt.Errorf("zipFile.Read: %v", err)
	}

	logging.Debug("Archive size: %d", len(archiveBytes))

	// Create a new writer to write the index zip file
	obj := bkt.Object(nameInStorage)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(archiveBytes); err != nil {
		return fmt.Errorf("Writer.Write: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

// GetFS fetches the directory from the Google Cloud Storage bucket and
// extracts it to the local filesystem. It returns the path of the file or
// directory in the archive.
// Returns an empty string if the archive is empty.
func GetFS(nameInStorage, targetPath string) (string, error) {
	ctx := context.Background()
	client, err := createClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bkt := client.Bucket(BucketName)
	obj := bkt.Object(nameInStorage)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("Object.NewReader: %v", err)
	}
	defer r.Close()

	// we check what the first file/directory is in the archive, and return
	// that path, since there can only ever be one root directory or file.
	rootDir := ""
	rootFile := ""
	// nil as the third argument to Extract means that all files will be extracted
	err = ArchiveFormat.Extract(context.Background(), r, nil, func(ctx context.Context, file archiver.File) error {
		rel := filepath.Clean(file.NameInArchive)
		abs := filepath.Join(targetPath, rel)

		mode := file.Mode()

		switch {
		case mode.IsRegular():
			f, err := os.Create(abs)
			if err != nil {
				return err
			}
			defer f.Close()
			fReader, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(f, fReader)
			if rootFile == "" && rootDir == "" {
				rootFile = abs
			}
			return err
		case mode.IsDir():
			if rootDir == "" && rootFile == "" {
				rootDir = abs
			}
			return os.MkdirAll(abs, 0o755)
		default:
			return fmt.Errorf("archive contained entry %s of unsupported file type %v", file.Name(), mode)
		}
	})
	if err != nil {
		return "", fmt.Errorf("Extract: %v", err)
	}

	// if the first extraction is a directory, return that, otherwise return
	// the first file
	if rootDir != "" {
		return rootDir, nil
	} else {
		return rootFile, nil
	}
}
