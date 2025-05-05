package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/gost/log"
	"github.com/hashicorp/go-version"
	"github.com/oschwald/maxminddb-golang"
)

const metadataChunkSize = 128 * 1024

var metadataStartMarker = []byte("\xAB\xCD\xEFMaxMind.com")

type MMDBRepository struct {
	cityDbSourceUrl string
	ispDbSourceUrl  string
	cityDbPath      string
	ispDbPath       string
}

func NewMMDBRepository(cityDbSourceUrl, ispDbSourceUrl, cityDbPath, ispDbPath string) *MMDBRepository {
	return &MMDBRepository{
		cityDbSourceUrl: cityDbSourceUrl,
		ispDbSourceUrl:  ispDbSourceUrl,
		cityDbPath:      cityDbPath,
		ispDbPath:       ispDbPath,
	}
}

func (r *MMDBRepository) extractVersion(metadataBuf []byte) (*version.Version, error) {
	metadataDecoder := maxmind.Decoder{Buffer: metadataBuf}

	var metadata maxminddb.Metadata
	rvMetadata := reflect.ValueOf(&metadata)
	_, err := metadataDecoder.Decode(0, rvMetadata, 0)
	if err != nil {
		return nil, fmt.Errorf("metadata decoding failed: %w", err)
	}

	return version.NewVersion(fmt.Sprintf("%d.%d", metadata.BinaryFormatMajorVersion, metadata.BinaryFormatMinorVersion))
}

func (r *MMDBRepository) fileMetadataVersion(path string) (*version.Version, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	startOffset := fileSize - metadataChunkSize
	if fileSize < metadataChunkSize {
		startOffset = 0
	}

	readSize := fileSize - startOffset
	buffer := make([]byte, readSize)
	_, err = file.ReadAt(buffer, startOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read file chunk: %w", err)
	}

	metadataIndex := bytes.LastIndex(buffer, metadataStartMarker)
	if metadataIndex == -1 {
		return nil, fmt.Errorf("metadata marker not found")
	}

	metadataStart := metadataIndex + len(metadataStartMarker)

	return r.extractVersion(buffer[metadataStart:])
}

func (r *MMDBRepository) DownloadCityDb(ctx context.Context, update ...bool) error {
	if r.cityDbPath == "" {
		return fmt.Errorf("city database path is not set")
	}

	if r.cityDbSourceUrl == "" {
		return fmt.Errorf("city database source URL is not set")
	}

	if len(update) == 0 || !update[0] {
		if _, err := os.Stat(r.cityDbPath); err == nil {
			log.FromContext(ctx).Info("City database already exists, skipping download")
			return nil
		}
	}

	log.FromContext(ctx).Infof("Downloading city database from %s to %s", r.cityDbSourceUrl, r.cityDbPath)
	return r.downloadMMDb(ctx, r.cityDbSourceUrl, r.cityDbPath)
}

func (r *MMDBRepository) DownloadISPDb(ctx context.Context, update ...bool) error {
	if r.ispDbPath == "" {
		return fmt.Errorf("isp database path is not set")
	}

	if r.ispDbSourceUrl == "" {
		return fmt.Errorf("isp database source URL is not set")
	}

	if len(update) == 0 || !update[0] {
		if _, err := os.Stat(r.ispDbPath); err == nil {
			log.FromContext(ctx).Info("ISP database already exists, skipping download")
			return nil
		}
	}

	log.FromContext(ctx).Infof("Downloading ISP database from %s to %s", r.ispDbSourceUrl, r.ispDbPath)
	return r.downloadMMDb(ctx, r.ispDbSourceUrl, r.ispDbPath)
}

func (r *MMDBRepository) downloadMMDb(ctx context.Context, sourceUrl, savePath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", sourceUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (r *MMDBRepository) CheckCityDbUpdates(ctx context.Context) (bool, error) {
	if r.cityDbPath == "" {
		return false, fmt.Errorf("city database path is not set")
	}

	if r.cityDbSourceUrl == "" {
		return false, fmt.Errorf("city database source URL is not set")
	}

	return r.checkMMDBUpdates(ctx, r.cityDbSourceUrl, r.cityDbPath)
}

func (r *MMDBRepository) CheckISPDbUpdates(ctx context.Context) (bool, error) {
	if r.ispDbPath == "" {
		return false, fmt.Errorf("isp database path is not set")
	}

	if r.ispDbSourceUrl == "" {
		return false, fmt.Errorf("isp database source URL is not set")
	}

	return r.checkMMDBUpdates(ctx, r.ispDbSourceUrl, r.ispDbPath)
}

func (r *MMDBRepository) checkMMDBUpdates(ctx context.Context, sourceUrl, localPath string) (bool, error) {
	res, err := r.downloadRange(ctx, sourceUrl, -metadataChunkSize)
	if err != nil {
		return false, err
	}

	metadataStart := bytes.LastIndex(res, metadataStartMarker)

	if metadataStart == -1 {
		return false, fmt.Errorf("metadata start marker not found in the response")
	}

	metadataStart += len(metadataStartMarker)

	remoteVersion, err := r.extractVersion(res[metadataStart:])
	if err != nil {
		return false, fmt.Errorf("failed to extract version from response: %w", err)
	}

	localVersion, err := r.fileMetadataVersion(localPath)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from local file: %w", err)
	}

	return remoteVersion.GreaterThan(localVersion), nil
}

func (r *MMDBRepository) downloadRange(ctx context.Context, sourceUrl string, chunkSize int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sourceUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d", chunkSize))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("server does not support range requests")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
