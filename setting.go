package rapid

import (
	"os"
	"path/filepath"
)

type (
	Setting interface {
		// location where the download will be placed
		DownloadLocation() string

		// location where the data for this application will be stored
		DataLocation() string

		// max retry will be executed when there is an Error downloading
		MaxRetry() int

		// logger provider that will be used to log something, e.g file, std, etc
		LoggerProvider() string

		// minimum size in MB for a chunk
		MinChunkSize() int64

		HttpClient() string
	}

	setting struct {
		downloadLocation string
		dataLocation     string
		maxRetry         int
		loggerProvider   string
		minChunkSize     int64
		httpClient       string
	}
)

func DefaultSetting() Setting {
	home, _ := os.UserHomeDir()

	// location
	data := filepath.Join(home, ".rapid")
	download := filepath.Join(home, "Downloads")

	os.MkdirAll(data, os.ModePerm)

	return &setting{
		downloadLocation: download,
		dataLocation:     data,
		maxRetry:         3,
		loggerProvider:   LoggerStdOut,
		minChunkSize:     1024 * 1024 * 5, // 5 MB
	}
}

func (s *setting) DownloadLocation() string {
	return s.downloadLocation
}

func (s *setting) DataLocation() string {
	return s.dataLocation
}

func (s *setting) MaxRetry() int {
	return s.maxRetry
}

func (s *setting) LoggerProvider() string {
	return s.loggerProvider
}

func (s *setting) MinChunkSize() int64 {
	return s.minChunkSize
}

func (s *setting) HttpClient() string {
	return s.httpClient
}
