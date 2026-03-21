package shared

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadGeoLiteDB() (string, error) {
	licenseKey := os.Getenv("MAXMIND_LICENSE_ID")
	if licenseKey == "" {
		return "", fmt.Errorf("MAXMIND_LICENSE_ID environment variable is not set")
	}

	// Remove and recreate the geoip folder
	if err := os.RemoveAll("geoip"); err != nil {
		return "", fmt.Errorf("failed to remove geoip folder: %w", err)
	}
	if err := os.MkdirAll("geoip", 0755); err != nil {
		return "", fmt.Errorf("failed to create geoip folder: %w", err)
	}

	// Download the tar.gz
	url := fmt.Sprintf(
		"https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz",
		licenseKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download GeoLite2-City: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Extract the tar.gz directly from the response body
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var mmdbFilename string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Only extract .mmdb files
		if !strings.HasSuffix(header.Name, ".mmdb") {
			continue
		}

		// Flatten the path, only use the base filename
		baseName := filepath.Base(header.Name)
		destPath := filepath.Join("geoip", baseName)

		outFile, err := os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return "", fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
		outFile.Close()

		mmdbFilename = baseName
	}

	if mmdbFilename == "" {
		return "", fmt.Errorf("no .mmdb file found in the downloaded archive")
	}

	return filepath.Join("geoip", mmdbFilename), nil
}
