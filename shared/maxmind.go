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

	"github.com/spf13/viper"
)

const maxFileSize = 100 << 20 // 100 MB

func DownloadGeoLiteDB() (string, error) {
	return downloadGeoLiteEdition("GeoLite2-City")
}

func DownloadGeoLiteASNDB() (string, error) {
	return downloadGeoLiteEdition("GeoLite2-ASN")
}

func downloadGeoLiteEdition(editionID string) (string, error) {
	licenseKey := viper.GetString("MAXMIND_LICENSE_ID")
	if licenseKey == "" {
		return "", fmt.Errorf("MAXMIND_LICENSE_ID environment variable is not set")
	}

	if err := os.MkdirAll("geoip", 0755); err != nil {
		return "", fmt.Errorf("failed to create geoip folder: %w", err)
	}

	url := fmt.Sprintf(
		"https://download.maxmind.com/app/geoip_download?edition_id=%s&license_key=%s&suffix=tar.gz",
		editionID, licenseKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", editionID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

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

		if header.Typeflag == tar.TypeDir {
			continue
		}

		if !strings.HasSuffix(header.Name, ".mmdb") {
			continue
		}

		if header.Size > maxFileSize {
			return "", fmt.Errorf("tar entry %s claims size %d which exceeds limit of %d bytes", header.Name, header.Size, maxFileSize)
		}

		baseName := filepath.Base(header.Name)
		destPath := filepath.Join("geoip", baseName)

		outFile, err := os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		written, err := io.Copy(outFile, io.LimitReader(tarReader, maxFileSize))
		if err != nil {
			outFile.Close()
			return "", fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
		if written >= maxFileSize {
			outFile.Close()
			return "", fmt.Errorf("file %s exceeds maximum allowed size of %d bytes, aborting", destPath, maxFileSize)
		}
		outFile.Close()

		mmdbFilename = baseName
	}

	if mmdbFilename == "" {
		return "", fmt.Errorf("no .mmdb file found in the downloaded archive")
	}

	return filepath.Join("geoip", mmdbFilename), nil
}
